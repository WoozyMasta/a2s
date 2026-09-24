// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

// pollerRetryDelay is the fixed base delay between refresh retries.
const pollerRetryDelay = time.Second

// cacheableQueries lists every query type supported by Cache and Poller.
var cacheableQueries = [...]a2s.QueryType{
	a2s.InfoRequest,
	a2s.PlayerRequest,
	a2s.RulesRequest,
}

// Upstream is the minimal client contract required by Poller and Handler.
type Upstream interface {
	Query(context.Context, a2s.QueryType) (a2s.Packet, a2s.QueryMeta, error)
}

// PollerConfig configures refresh timing and state notifications.
type PollerConfig struct {
	// OnStateChange receives inactive and recovered transitions.
	// The callback may be called concurrently and must return promptly.
	OnStateChange func(StateChange)
	// TTL is the delay between refreshes while a query is active.
	TTL time.Duration
	// InactiveTTL is the delay between probes after a query becomes inactive.
	// Zero inherits TTL.
	InactiveTTL time.Duration
	// Jitter is the maximum random delay added to every wait.
	Jitter time.Duration
	// Retries is the number of attempts after the initial failed refresh.
	Retries int
}

// StateChange describes a cache query becoming inactive or recovering.
type StateChange struct {
	// Err is the failure that caused an inactive transition, or nil on recovery.
	Err error
	// Query identifies the cache entry whose state changed.
	Query a2s.QueryType
	// Active reports whether the query is currently available.
	Active bool
}

// Poller periodically refreshes enabled cache entries from an upstream client.
//
// Each enabled query has an independent lifecycle.
// Upstream failures invalidate only the affected entry
// after retries are exhausted; they do not stop Run.
type Poller struct {
	upstream Upstream                                  // Upstream client used for logical queries.
	cache    *Cache                                    // Cache whose enabled entries are refreshed.
	wait     func(context.Context, time.Duration) bool // Interruptible wait helper.
	jitter   func(time.Duration) time.Duration         // Bounded jitter helper.
	config   PollerConfig                              // Refresh timing and notification settings.
	runMu    sync.Mutex                                // Serializes Poller.Run calls.
	run      bool                                      // Whether a polling run is active.
}

// NewPoller creates a poller for the enabled entries in cache.
func NewPoller(cache *Cache, upstream Upstream, config PollerConfig) (*Poller, error) {
	if cache == nil {
		return nil, fmt.Errorf("%w: cache is nil", ErrPoller)
	}
	if upstream == nil {
		return nil, fmt.Errorf("%w: upstream is nil", ErrPoller)
	}
	if config.TTL <= 0 {
		return nil, fmt.Errorf("%w: TTL must be positive", ErrPoller)
	}
	if config.InactiveTTL < 0 {
		return nil, fmt.Errorf("%w: inactive TTL must not be negative", ErrPoller)
	}
	if config.Jitter < 0 {
		return nil, fmt.Errorf("%w: jitter must not be negative", ErrPoller)
	}
	if config.Retries < 0 {
		return nil, fmt.Errorf("%w: retries must not be negative", ErrPoller)
	}
	if config.InactiveTTL == 0 {
		config.InactiveTTL = config.TTL
	}

	return &Poller{
		cache:    cache,
		upstream: upstream,
		config:   config,
		wait:     waitContext,
		jitter:   randomJitter,
	}, nil
}

// Run refreshes all enabled cache entries until ctx is canceled.
//
// Context cancellation is a normal shutdown and returns nil.
// A poller must not be run concurrently with itself.
func (p *Poller) Run(ctx context.Context) error {
	if p == nil {
		return fmt.Errorf("%w: poller is nil", ErrPoller)
	}
	if ctx == nil {
		return fmt.Errorf("%w: context is nil", ErrPoller)
	}

	p.runMu.Lock()
	if p.run {
		p.runMu.Unlock()
		return fmt.Errorf("%w: poller is already running", ErrPoller)
	}

	p.run = true
	p.runMu.Unlock()
	defer func() {
		p.runMu.Lock()
		p.run = false
		p.runMu.Unlock()
	}()

	runCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	var group sync.WaitGroup
	for _, query := range cacheableQueries {
		if !p.cache.Enabled(query) {
			continue
		}

		group.Go(func() {
			if err := p.runQuery(runCtx, query); err != nil {
				cancel(err)
			}
		})
	}
	group.Wait()

	if cause := context.Cause(runCtx); cause != nil &&
		!errors.Is(cause, context.Canceled) &&
		!errors.Is(cause, context.DeadlineExceeded) {
		return cause
	}

	return nil
}

// runQuery owns the independent lifecycle of one cacheable query.
func (p *Poller) runQuery(ctx context.Context, query a2s.QueryType) error {
	active := false
	if _, ok := p.cache.Load(query); ok {
		active = true
	}

	if !active {
		packet, err := p.refresh(ctx, query)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err == nil {
			if err := p.cache.Store(query, packet); err != nil {
				return fmt.Errorf("store cache for query 0x%X: %w", query, err)
			}
			active = true
		} else {
			if invalidateErr := p.markInactive(query, err); invalidateErr != nil {
				return fmt.Errorf("invalidate cache for query 0x%X: %w", query, invalidateErr)
			}
		}
	}

	for {
		if active {
			if !p.wait(ctx, p.delay(p.config.TTL)) {
				return nil
			}

			packet, err := p.refresh(ctx, query)
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if err == nil {
				if err := p.cache.Store(query, packet); err != nil {
					return fmt.Errorf("store cache for query 0x%X: %w", query, err)
				}
				continue
			}

			if invalidateErr := p.markInactive(query, err); invalidateErr != nil {
				return fmt.Errorf("invalidate cache for query 0x%X: %w", query, invalidateErr)
			}
			active = false
			continue
		}

		if !p.wait(ctx, p.delay(p.config.InactiveTTL)) {
			return nil
		}

		packet, _, err := p.upstream.Query(ctx, query)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err == nil {
			if err := p.cache.Store(query, packet); err != nil {
				return fmt.Errorf("store cache for query 0x%X: %w", query, err)
			}

			p.notify(StateChange{Query: query, Active: true})
			active = true
			continue
		}
	}
}

// refresh performs an initial query and the configured number of retries.
func (p *Poller) refresh(ctx context.Context, query a2s.QueryType) (a2s.Packet, error) {
	var lastErr error
	for attempt := 0; attempt <= p.config.Retries; attempt++ {
		packet, _, err := p.upstream.Query(ctx, query)
		if err == nil {
			return packet, nil
		}

		lastErr = err
		if ctx.Err() != nil || attempt == p.config.Retries {
			break
		}

		if !p.wait(ctx, p.delay(pollerRetryDelay)) {
			return a2s.Packet{}, ctx.Err()
		}
	}

	return a2s.Packet{}, lastErr
}

// markInactive invalidates one cache entry and reports its failure transition.
func (p *Poller) markInactive(query a2s.QueryType, err error) error {
	if invalidateErr := p.cache.Invalidate(query); invalidateErr != nil {
		return invalidateErr
	}

	p.notify(StateChange{Query: query, Err: err})
	return nil
}

// notify sends a state transition when a callback is configured.
func (p *Poller) notify(change StateChange) {
	if p.config.OnStateChange != nil {
		p.config.OnStateChange(change)
	}
}

// delay adds one bounded jitter sample to base.
func (p *Poller) delay(base time.Duration) time.Duration {
	return base + p.jitter(p.config.Jitter)
}

// waitContext waits for delay or context cancellation.
func waitContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// randomJitter returns a non-negative sample no greater than upperBound.
func randomJitter(upperBound time.Duration) time.Duration {
	if upperBound <= 0 {
		return 0
	}

	if upperBound == time.Duration(1<<63-1) {
		// #nosec G404 -- jitter is timing noise, not a security value.
		return time.Duration(rand.Int63())
	}

	// #nosec G404 -- jitter is timing noise, not a security value.
	return time.Duration(rand.Int63n(int64(upperBound) + 1))
}
