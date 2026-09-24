// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"context"
	"errors"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

func TestNewRateLimiterValidatesEnabledWindows(t *testing.T) {
	tests := []struct {
		name   string
		config RateLimitConfig
		want   bool
	}{
		{
			name: "global window required",
			config: RateLimitConfig{
				Global: Rate{Requests: 1},
			},
			want: true,
		},
		{
			name: "client window required",
			config: RateLimitConfig{
				Client: Rate{Requests: 1},
			},
			want: true,
		},
		{
			name: "disabled zero window",
			config: RateLimitConfig{
				Global: Rate{Window: -time.Second},
				Client: Rate{Window: -time.Second},
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			limiter, err := NewRateLimiter(test.config)
			if test.want {
				if limiter != nil || !errors.Is(err, ErrRateLimit) {
					t.Fatalf("NewRateLimiter() = %v, %v; want ErrRateLimit", limiter, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewRateLimiter() error = %v", err)
			}
		})
	}
}

func TestRateLimiterGlobalAllowance(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := mustTestRateLimiter(t, RateLimitConfig{
		Global: Rate{Requests: 3, Window: time.Second},
	}, &now)

	for index := 0; index < 3; index++ {
		if !limiter.allow(netip.MustParseAddr("192.0.2.1")) {
			t.Fatalf("allow(%d) = false, want true", index)
		}
	}
	if limiter.allow(netip.MustParseAddr("192.0.2.1")) {
		t.Fatal("fourth request was allowed past the global limit")
	}

	now = now.Add(time.Second)
	if !limiter.allow(netip.MustParseAddr("192.0.2.2")) {
		t.Fatal("request was not allowed after global refill")
	}
}

func TestRateLimiterClientRejectionDoesNotConsumeGlobalToken(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := mustTestRateLimiter(t, RateLimitConfig{
		Global: Rate{Requests: 2, Window: time.Second},
		Client: Rate{Requests: 1, Window: time.Second},
	}, &now)

	first := netip.MustParseAddr("192.0.2.1")
	second := netip.MustParseAddr("192.0.2.2")
	if !limiter.allow(first) {
		t.Fatal("first client request was not allowed")
	}
	if limiter.allow(first) {
		t.Fatal("client exceeded its limit")
	}
	if !limiter.allow(second) {
		t.Fatal("second client was blocked by a rejected first-client request")
	}
}

func TestRateLimiterClientUsesIPv4AddressWithoutPort(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := mustTestRateLimiter(t, RateLimitConfig{
		Client: Rate{Requests: 1, Window: time.Second},
	}, &now)

	if !limiter.allow(netip.MustParseAddr("192.0.2.1")) {
		t.Fatal("first request was not allowed")
	}
	if limiter.allow(netip.MustParseAddr("192.0.2.1")) {
		t.Fatal("same IPv4 address bypassed the client limit")
	}
	if !limiter.allow(netip.MustParseAddr("192.0.2.2")) {
		t.Fatal("different IPv4 address was not allowed")
	}

	if first, second := makeClientKey(netip.MustParseAddr("192.0.2.1")),
		makeClientKey(netip.MustParseAddr("192.0.2.1")); first != second {
		t.Fatal("same IPv4 address produced different client keys")
	}
}

func TestRateLimiterClientGroupsIPv6By64(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := mustTestRateLimiter(t, RateLimitConfig{
		Client: Rate{Requests: 1, Window: time.Second},
	}, &now)

	if !limiter.allow(netip.MustParseAddr("2001:db8:1:2::1")) {
		t.Fatal("first IPv6 request was not allowed")
	}
	if limiter.allow(netip.MustParseAddr("2001:db8:1:2::2")) {
		t.Fatal("same IPv6 /64 bypassed the client limit")
	}
	if !limiter.allow(netip.MustParseAddr("2001:db8:1:3::1")) {
		t.Fatal("different IPv6 /64 was not allowed")
	}
}

func TestRateLimiterDisabledLimitsDoNotCreateClientState(t *testing.T) {
	limiter := mustTestRateLimiter(t, RateLimitConfig{}, nil)
	for index := 1; index <= 100; index++ {
		address := testIPv4(byte(index))
		if !limiter.allow(address) {
			t.Fatalf("allow(%s) = false, want true", address)
		}
	}
	if got := len(limiter.clients); got != 0 {
		t.Fatalf("client state size = %d, want 0", got)
	}
}

func TestRateLimiterRefillsClientBucket(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := mustTestRateLimiter(t, RateLimitConfig{
		Client: Rate{Requests: 2, Window: time.Second},
	}, &now)

	address := netip.MustParseAddr("192.0.2.1")
	if !limiter.allow(address) || !limiter.allow(address) {
		t.Fatal("initial burst was not allowed")
	}
	if limiter.allow(address) {
		t.Fatal("request exceeded empty bucket")
	}

	now = now.Add(500 * time.Millisecond)
	if !limiter.allow(address) {
		t.Fatal("request was not allowed after partial refill")
	}
	if limiter.allow(address) {
		t.Fatal("request exceeded partially refilled bucket")
	}
}

func TestRateLimiterExpiresStaleClients(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := mustTestRateLimiter(t, RateLimitConfig{
		Client: Rate{Requests: 1, Window: time.Second},
	}, &now)

	first := netip.MustParseAddr("192.0.2.1")
	second := netip.MustParseAddr("192.0.2.2")
	if !limiter.allow(first) {
		t.Fatal("first client request was not allowed")
	}

	now = now.Add(3 * time.Second)
	if !limiter.allow(second) {
		t.Fatal("second client request was not allowed")
	}
	if got := len(limiter.clients); got != 1 {
		t.Fatalf("client state size = %d, want 1 after expiry", got)
	}
	if !limiter.allow(first) {
		t.Fatal("expired client did not receive a new bucket")
	}
}

func TestRateLimiterGlobalAdmissionBoundsClientState(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := mustTestRateLimiter(t, RateLimitConfig{
		Global: Rate{Requests: 2, Window: time.Second},
		Client: Rate{Requests: 1, Window: time.Second},
	}, &now)

	for index := 1; index <= 100; index++ {
		address := testIPv4(byte(index))
		limiter.allow(address)
	}
	if got := len(limiter.clients); got != 2 {
		t.Fatalf("client state size = %d, want 2 admitted clients", got)
	}
}

func TestRateLimiterMaxClientsBoundsPerClientState(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := mustTestRateLimiter(t, RateLimitConfig{
		Client:     Rate{Requests: 1, Window: time.Second},
		MaxClients: 2,
	}, &now)

	if !limiter.allow(testIPv4(1)) || !limiter.allow(testIPv4(2)) {
		t.Fatal("initial clients were not admitted")
	}
	if limiter.allow(testIPv4(3)) {
		t.Fatal("new client was admitted past MaxClients")
	}
	if got := len(limiter.clients); got != 2 {
		t.Fatalf("client state size = %d, want 2", got)
	}
}

func TestRateLimiterExpiredClientsFreeBoundedState(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := mustTestRateLimiter(t, RateLimitConfig{
		Client:     Rate{Requests: 1, Window: time.Second},
		MaxClients: 1,
	}, &now)

	if !limiter.allow(testIPv4(1)) {
		t.Fatal("first client was not admitted")
	}
	now = now.Add(3 * time.Second)
	if !limiter.allow(testIPv4(2)) {
		t.Fatal("new client did not replace expired state")
	}
	if got := len(limiter.clients); got != 1 {
		t.Fatalf("client state size = %d, want 1", got)
	}
	if _, ok := limiter.clients[makeClientKey(testIPv4(1))]; ok {
		t.Fatal("expired client state was retained")
	}
}

func TestRateLimiterMiddlewareDropsExcessRequests(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := mustTestRateLimiter(t, RateLimitConfig{
		Global: Rate{Requests: 1, Window: time.Second},
	}, &now)
	calls := 0
	handler := limiter.Middleware()(server.HandlerFunc(func(context.Context, *server.Request) (server.Response, error) {
		calls++
		return server.PacketResponse{Packet: a2s.Packet{Type: a2s.ResponsePing}}, nil
	}))
	request := &server.Request{
		Remote: netip.MustParseAddrPort("192.0.2.1:27015"),
		Query:  a2s.Request{Type: a2s.PingRequest},
	}

	if _, err := handler.Handle(context.Background(), request); err != nil {
		t.Fatalf("first Handle() error = %v", err)
	}
	if _, err := handler.Handle(context.Background(), request); !errors.Is(err, server.ErrDrop) {
		t.Fatalf("second Handle() error = %v, want ErrDrop", err)
	}
	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}
}

func TestRateLimiterIsSafeForConcurrentAccess(t *testing.T) {
	limiter, err := NewRateLimiter(RateLimitConfig{
		Global: Rate{Requests: 1000, Window: time.Second},
		Client: Rate{Requests: 1000, Window: time.Second},
	})
	if err != nil {
		t.Fatalf("NewRateLimiter() error = %v", err)
	}

	var wait sync.WaitGroup
	for range 100 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			limiter.allow(netip.MustParseAddr("192.0.2.1"))
		}()
	}
	wait.Wait()
}

func testIPv4(lastOctet byte) netip.Addr {
	return netip.AddrFrom4([4]byte{192, 0, 2, lastOctet})
}

func mustTestRateLimiter(t *testing.T, config RateLimitConfig, now *time.Time) *RateLimiter {
	t.Helper()
	clock := time.Now
	if now != nil {
		clock = func() time.Time { return *now }
	}

	limiter, err := newRateLimiter(config, clock)
	if err != nil {
		t.Fatalf("newRateLimiter() error = %v", err)
	}
	return limiter
}
