// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

// Pool executes bounded concurrent A2S transactions over reusable worker sockets.
//
// One worker owns one transaction at a time.
// The pool never multiplexes concurrent transactions over the same UDP socket.
// Use New to construct a pool; the zero value is not ready for queries.
type Pool struct {
	root       context.Context
	cancel     context.CancelFunc
	jobs       chan workItem
	queueSpace chan struct{}
	done       chan struct{}
	resolver   func(context.Context, string) ([]netip.Addr, error)
	workers    sync.WaitGroup
	timeout    time.Duration
	stateMu    sync.RWMutex
	closeOnce  sync.Once
	bufferSize uint16
	closed     bool
	window     int
}

// workItem is one accepted query and its private completion path.
type workItem struct {
	ctx     context.Context
	result  chan queryResult
	stream  *streamSink
	cleanup context.CancelFunc
	target  netip.AddrPort
	request Request
	index   uint64
	query   transport.QueryType
}

// queryResult is the worker-to-caller completion value.
type queryResult struct {
	err    error
	packet transport.Packet
	meta   transport.QueryMeta
}

// New creates a pool with reusable UDP workers.
func New(options ...Option) (*Pool, error) {
	config := config{
		workers:    DefaultWorkers,
		queueSize:  DefaultQueueSize,
		timeout:    DefaultTimeout,
		bufferSize: DefaultBufferSize,
	}

	for _, option := range options {
		if option == nil {
			return nil, errors.Join(ErrInvalidConfig, errors.New("nil option"))
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}

	root, cancel := context.WithCancel(context.Background())
	p := &Pool{
		root:       root,
		cancel:     cancel,
		jobs:       make(chan workItem, config.queueSize),
		queueSpace: make(chan struct{}, 1),
		done:       make(chan struct{}),
		timeout:    config.timeout,
		bufferSize: config.bufferSize,
		resolver:   lookupHost,
		window:     config.workers + config.queueSize,
	}

	p.workers.Add(config.workers)
	for index := 0; index < config.workers; index++ {
		go p.worker()
	}

	return p, nil
}

// Query submits one query to the pool and waits for its result.
//
// The pool timeout starts when Query begins submission,
// so queue wait is part of the total budget.
// A resolved target is queried directly without DNS.
func (p *Pool) Query(
	ctx context.Context,
	target netip.AddrPort,
	query transport.QueryType,
) (transport.Packet, transport.QueryMeta, error) {
	if p == nil {
		return transport.Packet{}, transport.QueryMeta{}, ErrClosed
	}
	if ctx == nil {
		return transport.Packet{}, transport.QueryMeta{}, ErrNilContext
	}
	if err := p.ensureOpen(); err != nil {
		return transport.Packet{}, transport.QueryMeta{}, err
	}
	if err := validateTarget(target); err != nil {
		return transport.Packet{}, transport.QueryMeta{}, err
	}
	if !supportedQuery(query) {
		return transport.Packet{}, transport.QueryMeta{},
			fmtQueryError(query)
	}

	jobCtx, cancel := p.queryContext(ctx)
	defer cancel()

	job := workItem{
		ctx:    jobCtx,
		target: normalizeTarget(target),
		query:  query,
		result: make(chan queryResult, 1),
	}
	if err := p.enqueue(jobCtx, job); err != nil {
		return transport.Packet{}, transport.QueryMeta{}, err
	}

	select {
	case result := <-job.result:
		return result.packet, result.meta, result.err

	case <-p.done:
		return transport.Packet{}, transport.QueryMeta{}, ErrClosed

	case <-ctx.Done():
		return transport.Packet{}, transport.QueryMeta{}, ctx.Err()

	case <-jobCtx.Done():
		if errors.Is(p.root.Err(), context.Canceled) {
			return transport.Packet{}, transport.QueryMeta{}, ErrClosed
		}
		return transport.Packet{}, transport.QueryMeta{}, jobCtx.Err()
	}
}

// Close stops workers, cancels accepted queries, and releases worker sockets.
// It is safe to call Close more than once.
func (p *Pool) Close() error {
	if p == nil || p.root == nil {
		return nil
	}

	p.closeOnce.Do(func() {
		p.stateMu.Lock()
		p.closed = true
		p.stateMu.Unlock()
		p.cancel()
		close(p.done)
		p.workers.Wait()
		p.cleanupQueuedJobs()
	})

	return nil
}

// cleanupQueuedJobs releases stream contexts left in the queue during shutdown.
func (p *Pool) cleanupQueuedJobs() {
	for {
		select {
		case job := <-p.jobs:
			if job.cleanup != nil {
				job.cleanup()
			}
		default:
			return
		}
	}
}

// queryContext creates a context bounded by the pool root, caller, and timeout.
func (p *Pool) queryContext(ctx context.Context) (context.Context, context.CancelFunc) {
	queryCtx, cancel := context.WithCancel(p.root)
	stopCaller := context.AfterFunc(ctx, cancel)

	deadline, hasDeadline := ctx.Deadline()
	if p.timeout > 0 {
		poolDeadline := time.Now().Add(p.timeout)
		if !hasDeadline || poolDeadline.Before(deadline) {
			deadline = poolDeadline
			hasDeadline = true
		}
	}

	if hasDeadline {
		var deadlineCancel context.CancelFunc
		queryCtx, deadlineCancel = context.WithDeadline(queryCtx, deadline)
		return queryCtx, func() {
			stopCaller()
			deadlineCancel()
			cancel()
		}
	}

	return queryCtx, func() {
		stopCaller()
		cancel()
	}
}

// enqueue adds an item while preserving bounded backpressure.
func (p *Pool) enqueue(ctx context.Context, item workItem) error {
	for {
		p.stateMu.RLock()
		if p.closed {
			p.stateMu.RUnlock()
			return ErrClosed
		}

		select {
		case p.jobs <- item:
			p.stateMu.RUnlock()
			return nil

		default:
			p.stateMu.RUnlock()
		}

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-p.done:
			return ErrClosed

		case <-p.queueSpace:
		}
	}
}

// validateTarget rejects unresolved, unspecified, and zero-port targets.
func validateTarget(target netip.AddrPort) error {
	if !target.IsValid() || target.Port() == 0 || target.Addr().IsUnspecified() {
		return ErrInvalidTarget
	}

	return nil
}

// normalizeTarget removes IPv4-mapped IPv6 representation differences.
func normalizeTarget(target netip.AddrPort) netip.AddrPort {
	return netip.AddrPortFrom(target.Addr().Unmap(), target.Port())
}

// supportedQuery reports whether the transport engine can execute the query.
func supportedQuery(query transport.QueryType) bool {
	switch query {
	case
		transport.InfoRequest,
		transport.PlayerRequest,
		transport.RulesRequest,
		transport.ChallengeRequest,
		transport.PingRequest:
		return true

	default:
		return false
	}
}

// fmtQueryError keeps the invalid query value available to callers.
func fmtQueryError(query transport.QueryType) error {
	return fmt.Errorf("%w: 0x%X", ErrInvalidQuery, query)
}
