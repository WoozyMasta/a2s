// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"context"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/server"
)

// Rate defines a token-bucket request limit over a time window.
// Requests set to zero disable the limit.
type Rate struct {
	// Requests is the maximum token-bucket capacity and request count per window.
	Requests uint32
	// Window is the token refill interval.
	Window time.Duration
}

// RateLimitConfig configures independent global and per-client limits.
type RateLimitConfig struct {
	// Global limits requests across all downstream clients.
	Global Rate
	// Client limits requests for each IPv4 address or IPv6 /64 prefix.
	Client Rate
}

// RateLimiter limits decoded downstream requests before challenge handling.
// RateLimiter is safe for concurrent use by server workers.
type RateLimiter struct {
	nextCleanup   time.Time
	clients       map[clientKey]*clientBucket
	now           func() time.Time
	global        tokenBucket
	globalRate    Rate
	clientRate    Rate
	mu            sync.Mutex
	globalEnabled bool
}

// clientKey identifies a client by its IPv4 address or IPv6 /64 prefix.
type clientKey struct {
	address netip.Addr
}

// clientBucket stores one client's tokens and activity timestamp.
type clientBucket struct {
	lastSeen time.Time
	bucket   tokenBucket
}

// tokenBucket stores the current token balance and refill timestamp.
type tokenBucket struct {
	updatedAt time.Time
	tokens    float64
}

// NewRateLimiter creates a token-bucket limiter from the supplied limits.
// A limit with zero Requests is disabled and does not require a window.
func NewRateLimiter(config RateLimitConfig) (*RateLimiter, error) {
	return newRateLimiter(config, time.Now)
}

// newRateLimiter creates a limiter with an injectable clock for deterministic tests.
func newRateLimiter(config RateLimitConfig, now func() time.Time) (*RateLimiter, error) {
	if now == nil {
		return nil, fmt.Errorf("%w: clock is nil", ErrRateLimit)
	}
	if err := validateRate(config.Global, "global"); err != nil {
		return nil, err
	}
	if err := validateRate(config.Client, "client"); err != nil {
		return nil, err
	}

	limiter := &RateLimiter{
		globalRate: config.Global,
		clientRate: config.Client,
		clients:    make(map[clientKey]*clientBucket),
		now:        now,
	}
	if config.Global.Requests > 0 {
		limiter.global = newTokenBucket(config.Global, now())
		limiter.globalEnabled = true
	}

	return limiter, nil
}

// Middleware returns server middleware that drops requests over either limit.
// Dropped requests return server.ErrDrop and produce no response packet.
func (l *RateLimiter) Middleware() server.Middleware {
	return func(next server.Handler) server.Handler {
		return server.HandlerFunc(func(ctx context.Context, request *server.Request) (server.Response, error) {
			if l == nil {
				return nil, fmt.Errorf("%w: limiter is nil", ErrRateLimit)
			}
			if request == nil {
				return nil, fmt.Errorf("%w: request is nil", ErrRateLimit)
			}
			if !l.allow(request.Remote.Addr()) {
				return nil, server.ErrDrop
			}

			return next.Handle(ctx, request)
		})
	}
}

// allow consumes global and per-client tokens in that order.
func (l *RateLimiter) allow(address netip.Addr) bool {
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.globalEnabled && !l.global.allow(now, l.globalRate) {
		return false
	}
	if l.clientRate.Requests == 0 {
		return true
	}

	key := makeClientKey(address)
	state, ok := l.clients[key]
	if !ok {
		l.cleanupClients(now)
		state = &clientBucket{
			bucket:   newTokenBucket(l.clientRate, now),
			lastSeen: now,
		}
		l.clients[key] = state
	} else {
		state.lastSeen = now
	}

	return state.bucket.allow(now, l.clientRate)
}

// cleanupClients removes clients idle for approximately two client windows.
func (l *RateLimiter) cleanupClients(now time.Time) {
	if !l.nextCleanup.IsZero() && now.Before(l.nextCleanup) {
		return
	}

	expiry := 2 * l.clientRate.Window
	for key, state := range l.clients {
		if now.Sub(state.lastSeen) >= expiry {
			delete(l.clients, key)
		}
	}
	l.nextCleanup = now.Add(l.clientRate.Window)
}

// makeClientKey groups IPv4 clients by address and IPv6 clients by /64.
func makeClientKey(address netip.Addr) clientKey {
	address = address.Unmap()
	if address.Is4() {
		return clientKey{address: address}
	}
	if address.Is6() {
		prefix, _ := address.Prefix(64)
		return clientKey{address: prefix.Masked().Addr()}
	}

	return clientKey{}
}

// newTokenBucket creates a full bucket for a configured rate.
func newTokenBucket(rate Rate, now time.Time) tokenBucket {
	return tokenBucket{
		tokens:    float64(rate.Requests),
		updatedAt: now,
	}
}

// allow refills and consumes one token when available.
func (b *tokenBucket) allow(now time.Time, rate Rate) bool {
	if elapsed := now.Sub(b.updatedAt); elapsed > 0 {
		refill := float64(elapsed) / float64(rate.Window) * float64(rate.Requests)
		b.tokens += refill
		if b.tokens > float64(rate.Requests) {
			b.tokens = float64(rate.Requests)
		}
		b.updatedAt = now
	}
	if b.tokens < 1 {
		return false
	}

	b.tokens--
	return true
}

// validateRate validates only enabled limits.
func validateRate(rate Rate, name string) error {
	if rate.Requests > 0 && rate.Window <= 0 {
		return fmt.Errorf("%w: %s window must be positive", ErrRateLimit, name)
	}

	return nil
}
