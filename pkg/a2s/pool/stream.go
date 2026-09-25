// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"context"
	"net/netip"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

// Request identifies one target and query submitted to Stream.
type Request struct {
	// Target is the already resolved server endpoint.
	Target netip.AddrPort
	// Query is the A2S request type to execute.
	Query transport.QueryType
}

// Result contains one completed Stream request.
type Result struct {
	// Err is the request error, if any.
	Err error
	// Request is the request associated with the result.
	Request Request
	// Packet is the logical A2S response, when one was received.
	Packet transport.Packet
	// Meta contains transaction timing and challenge metadata.
	Meta transport.QueryMeta
	// Index is the zero-based submission order within one stream.
	Index uint64
}

// streamSink is the bounded internal completion path for one stream.
type streamSink struct {
	ctx         context.Context
	completions chan streamCompletion
}

// streamCompletion carries worker output before public delivery.
type streamCompletion struct {
	request Request
	result  queryResult
	index   uint64
}

// Stream submits requests from the caller-owned channel and returns results in completion order.
// The caller closes requests; Stream closes the result channel.
// Input acceptance is bounded by the pool worker and queue capacity.
//
// Canceling ctx stops input acceptance, cancels accepted work,
// and closes the result channel without waiting for the caller to close requests.
func (p *Pool) Stream(ctx context.Context, requests <-chan Request) <-chan Result {
	output := make(chan Result, 1)
	if p == nil || ctx == nil {
		close(output)
		return output
	}
	if err := p.ensureOpen(); err != nil {
		close(output)
		return output
	}

	streamCtx, cancel := context.WithCancel(p.root)
	stopCaller := context.AfterFunc(ctx, cancel)
	sink := &streamSink{
		ctx:         streamCtx,
		completions: make(chan streamCompletion, p.window),
	}

	go func() {
		defer close(output)
		defer stopCaller()
		defer cancel()
		p.runStream(streamCtx, sink, requests, output)
	}()

	return output
}

// runStream coordinates bounded input acceptance and result delivery.
func (p *Pool) runStream(
	ctx context.Context,
	sink *streamSink,
	requests <-chan Request,
	output chan<- Result,
) {
	input := requests
	var (
		inputOpen   = true
		outstanding int
		nextIndex   uint64
	)

	for inputOpen || outstanding > 0 {
		if ctx.Err() != nil {
			return
		}

		if outstanding == p.window {
			completion, ok := receiveCompletion(ctx, sink)
			if !ok {
				return
			}
			outstanding--
			if !sendResult(ctx, output, completion) {
				return
			}
			continue
		}

		select {
		case <-ctx.Done():
			return

		case completion := <-sink.completions:
			outstanding--
			if !sendResult(ctx, output, completion) {
				return
			}

		case request, ok := <-input:
			if !ok {
				input = nil
				inputOpen = false
				continue
			}

			index := nextIndex
			nextIndex++
			outstanding++
			if err := p.submitStreamJob(ctx, sink, index, request); err != nil {
				sink.publish(streamCompletion{
					index:   index,
					request: request,
					result:  queryResult{err: err},
				})
			}
		}
	}
}

// submitStreamJob validates and enqueues one stream request.
func (p *Pool) submitStreamJob(
	ctx context.Context,
	sink *streamSink,
	index uint64,
	request Request,
) error {
	if err := validateTarget(request.Target); err != nil {
		return err
	}
	if !supportedQuery(request.Query) {
		return fmtQueryError(request.Query)
	}

	jobCtx, cancel := p.queryContext(ctx)
	job := workItem{
		ctx:     jobCtx,
		stream:  sink,
		cleanup: cancel,
		index:   index,
		request: request,
		target:  normalizeTarget(request.Target),
		query:   request.Query,
	}
	if err := p.enqueue(jobCtx, job); err != nil {
		cancel()
		return err
	}

	return nil
}

// publish sends a completion unless the stream has been canceled.
func (s *streamSink) publish(completion streamCompletion) {
	select {
	case s.completions <- completion:
	case <-s.ctx.Done():
	}
}

// receiveCompletion waits for one worker completion or stream cancellation.
func receiveCompletion(ctx context.Context, sink *streamSink) (streamCompletion, bool) {
	select {
	case completion := <-sink.completions:
		return completion, true
	case <-ctx.Done():
		return streamCompletion{}, false
	}
}

// sendResult delivers one completion without blocking cancellation.
func sendResult(ctx context.Context, output chan<- Result, completion streamCompletion) bool {
	result := Result{
		Index:   completion.index,
		Request: completion.request,
		Packet:  completion.result.packet,
		Meta:    completion.result.meta,
		Err:     completion.result.err,
	}

	select {
	case output <- result:
		return true
	case <-ctx.Done():
		return false
	}
}
