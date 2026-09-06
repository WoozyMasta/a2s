// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"bytes"
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

func TestRateLimiterLocalUDPSingleClientFlood(t *testing.T) {
	handler := &e2eResponseHandler{
		packets: map[a2s.QueryType]a2s.Packet{
			a2s.PingRequest: {Type: a2s.ResponsePing},
		},
	}
	limiter := mustNewE2ERateLimiter(t, RateLimitConfig{
		Client: Rate{Requests: 3, Window: time.Hour},
	})
	downstream := startFloodE2EServer(t, handler,
		server.WithChallengePolicy(server.NoChallengePolicy()),
		server.WithMiddleware(limiter.Middleware()),
	)

	sendUDPFlood(t, downstream.addr, a2s.Request{Type: a2s.PingRequest}, 1, 32)
	assertE2ECount(t, func() int { return handler.Count(a2s.PingRequest) }, 3)
}

func TestRateLimiterLocalUDPGlobalFlood(t *testing.T) {
	handler := &e2eResponseHandler{
		packets: map[a2s.QueryType]a2s.Packet{
			a2s.PingRequest: {Type: a2s.ResponsePing},
		},
	}
	limiter := mustNewE2ERateLimiter(t, RateLimitConfig{
		Global: Rate{Requests: 4, Window: time.Hour},
		Client: Rate{Requests: 100, Window: time.Hour},
	})
	downstream := startFloodE2EServer(t, handler,
		server.WithChallengePolicy(server.NoChallengePolicy()),
		server.WithMiddleware(limiter.Middleware()),
	)

	sendUDPFlood(t, downstream.addr, a2s.Request{Type: a2s.PingRequest}, 32, 1)
	assertE2ECount(t, func() int { return handler.Count(a2s.PingRequest) }, 4)
	if got := len(limiter.clients); got > 4 {
		t.Fatalf("client state size = %d, want at most 4 admitted clients", got)
	}
}

func TestRateLimiterLocalUDPIgnoresSourcePort(t *testing.T) {
	handler := &e2eResponseHandler{
		packets: map[a2s.QueryType]a2s.Packet{
			a2s.PingRequest: {Type: a2s.ResponsePing},
		},
	}
	limiter := mustNewE2ERateLimiter(t, RateLimitConfig{
		Client: Rate{Requests: 1, Window: time.Hour},
	})
	downstream := startFloodE2EServer(t, handler,
		server.WithChallengePolicy(server.NoChallengePolicy()),
		server.WithMiddleware(limiter.Middleware()),
	)

	// Each request uses a fresh source port, but all clients share 127.0.0.1.
	sendUDPFlood(t, downstream.addr, a2s.Request{Type: a2s.PingRequest}, 32, 1)
	assertE2ECount(t, func() int { return handler.Count(a2s.PingRequest) }, 1)
}

func TestRateLimiterRunsBeforeChallengeGate(t *testing.T) {
	provider := &e2eChallengeProvider{token: a2s.Challenge{5, 6, 7, 8}}
	handler := &e2eResponseHandler{
		packets: map[a2s.QueryType]a2s.Packet{
			a2s.InfoRequest: {Type: a2s.ResponseInfo},
		},
	}
	limiter := mustNewE2ERateLimiter(t, RateLimitConfig{
		Global: Rate{Requests: 2, Window: time.Hour},
	})
	downstream := startFloodE2EServer(t, handler,
		server.WithChallengeProvider(provider),
		server.WithMiddleware(limiter.Middleware()),
	)

	sendUDPFlood(t, downstream.addr, a2s.Request{Type: a2s.InfoRequest}, 1, 32)
	assertE2ECount(t, func() int { return int(provider.issued.Load()) }, 2)
	time.Sleep(20 * time.Millisecond)
	if got := handler.Count(a2s.InfoRequest); got != 0 {
		t.Fatalf("handler calls = %d, want 0 for challenge-only flood", got)
	}
	if got := provider.issued.Load(); got != 2 {
		t.Fatalf("issued challenges = %d, want 2", got)
	}
}

func TestRateLimiterLocalUDPCachedLargeResponseFlood(t *testing.T) {
	cache := mustCache(t, a2s.RulesRequest)
	payload := bytes.Repeat([]byte("cached-rule="), 1024)
	if err := cache.Store(a2s.RulesRequest, a2s.Packet{
		Type:    a2s.ResponseRules,
		Payload: payload,
	}, time.Now()); err != nil {
		t.Fatalf("cache.Store() error = %v", err)
	}
	relay := &handlerUpstream{}
	handler := mustHandler(t, cache, relay, false)
	var accepted atomic.Int32
	limiter := mustNewE2ERateLimiter(t, RateLimitConfig{
		Global: Rate{Requests: 4, Window: time.Hour},
	})
	downstream := startFloodE2EServer(t, handler,
		server.WithChallengePolicy(server.NoChallengePolicy()),
		server.WithMiddleware(limiter.Middleware(), countAccepted(&accepted)),
	)

	sendUDPFlood(t, downstream.addr, a2s.Request{
		Type:         a2s.RulesRequest,
		Challenge:    a2s.InitialChallenge,
		HasChallenge: true,
	}, 1, 3)

	client := newE2EClient(t, downstream.addr)
	got, _, err := client.Query(context.Background(), a2s.RulesRequest)
	if err != nil {
		t.Fatalf("cached Rules Query() error = %v", err)
	}
	if got.Type != a2s.ResponseRules || !bytes.Equal(got.Payload, payload) {
		t.Fatalf("cached packet = %#v, want rules payload of %d bytes", got, len(payload))
	}
	assertE2ECount(t, func() int { return int(accepted.Load()) }, 4)
	if got := relay.Count(a2s.RulesRequest); got != 0 {
		t.Fatalf("upstream rules calls = %d, want 0 for cached response", got)
	}
}

func TestRateLimiterPreservesCachedAndLocalResponsesDuringRelay(t *testing.T) {
	relay := newBlockingUpstream()
	cache := mustCache(t, a2s.InfoRequest)
	if err := cache.Store(a2s.InfoRequest, a2s.Packet{
		Type:    a2s.ResponseInfo,
		Payload: []byte("cached info"),
	}, time.Now()); err != nil {
		t.Fatalf("cache.Store() error = %v", err)
	}
	handler := mustHandler(t, cache, relay, true)
	limiter := mustNewE2ERateLimiter(t, RateLimitConfig{
		Global: Rate{Requests: 100, Window: time.Hour},
	})
	downstream := startFloodE2EServer(t, handler,
		server.WithChallengePolicy(server.NoChallengePolicy()),
		server.WithWorkers(8),
		server.WithMiddleware(limiter.Middleware()),
	)

	relayClient := newE2EClient(t, downstream.addr)
	relayDone := make(chan error, 1)
	go func() {
		_, _, err := relayClient.Query(context.Background(), a2s.PlayerRequest)
		relayDone <- err
	}()
	waitForRelayStart(t, relay)

	client := newE2EClient(t, downstream.addr)
	got, _, err := client.Query(context.Background(), a2s.InfoRequest)
	if err != nil {
		t.Fatalf("cached INFO query error = %v", err)
	}
	if string(got.Payload) != "cached info" {
		t.Fatalf("cached INFO payload = %q, want cached info", got.Payload)
	}

	got, _, err = client.Query(context.Background(), a2s.PingRequest)
	if err != nil {
		t.Fatalf("local PING query error = %v", err)
	}
	if got.Type != a2s.ResponsePing {
		t.Fatalf("local PING response type = %X, want %X", got.Type, a2s.ResponsePing)
	}
	if got := relay.calls.Load(); got != 1 {
		t.Fatalf("relay calls during cached/local queries = %d, want 1", got)
	}

	close(relay.release)
	if err := <-relayDone; err != nil {
		t.Fatalf("live relay query error = %v", err)
	}
}

func TestRateLimiterLocalUDPLiveRelayFlood(t *testing.T) {
	relay := newBlockingUpstream()
	handler := mustHandler(t, mustCache(t), relay, false)
	downstream := startFloodE2EServer(t, handler,
		server.WithChallengePolicy(server.NoChallengePolicy()),
		server.WithWorkers(8),
	)

	client, err := net.DialUDP("udp4", nil, downstream.addr)
	if err != nil {
		t.Fatalf("net.DialUDP() error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	request, err := a2s.AppendRequest(nil, a2s.Request{Type: a2s.InfoRequest})
	if err != nil {
		t.Fatalf("a2s.AppendRequest() error = %v", err)
	}
	if _, err := client.Write(request); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	waitForRelayStart(t, relay)
	for range 32 {
		if _, err := client.Write(request); err != nil {
			t.Fatalf("flood Write() error = %v", err)
		}
	}
	time.Sleep(20 * time.Millisecond)
	if got := relay.calls.Load(); got != 1 {
		t.Fatalf("live relay calls during flood = %d, want 1", got)
	}

	close(relay.release)
	waitForE2E(t, func() bool { return relay.calls.Load() == 1 })
}

// startFloodE2EServer starts a local UDP server with caller-selected options.
func startFloodE2EServer(
	t *testing.T,
	handler server.Handler,
	options ...server.Option,
) *e2eServer {
	t.Helper()
	instance, err := server.New(handler, options...)
	if err != nil {
		t.Fatalf("server.New() error = %v", err)
	}
	conn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.ListenPacket() error = %v", err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- instance.Serve(conn) }()

	e2e := &e2eServer{
		server:   instance,
		conn:     conn,
		addr:     conn.LocalAddr().(*net.UDPAddr),
		serveErr: serveErr,
	}
	t.Cleanup(func() {
		if err := instance.Shutdown(context.Background()); err != nil {
			t.Errorf("server.Shutdown() error = %v", err)
		}
		if err := <-serveErr; err != nil && err != server.ErrServerClosed {
			t.Errorf("Serve() error = %v, want ErrServerClosed", err)
		}
		_ = conn.Close()
	})

	return e2e
}

// sendUDPFlood sends the same request from one or more local UDP clients.
func sendUDPFlood(
	t *testing.T,
	addr *net.UDPAddr,
	request a2s.Request,
	clients int,
	requestsPerClient int,
) {
	t.Helper()
	data, err := a2s.AppendRequest(nil, request)
	if err != nil {
		t.Fatalf("a2s.AppendRequest() error = %v", err)
	}
	for range clients {
		client, err := net.DialUDP("udp4", nil, addr)
		if err != nil {
			t.Fatalf("net.DialUDP() error = %v", err)
		}
		for range requestsPerClient {
			if _, err := client.Write(data); err != nil {
				_ = client.Close()
				t.Fatalf("UDP flood Write() error = %v", err)
			}
		}
		if err := client.Close(); err != nil {
			t.Fatalf("UDP client Close() error = %v", err)
		}
	}
}

// assertE2ECount waits for an allowance to be observed and rejects extra calls.
func assertE2ECount(t *testing.T, count func() int, want int) {
	t.Helper()
	waitForE2E(t, func() bool { return count() >= want })
	time.Sleep(20 * time.Millisecond)
	if got := count(); got != want {
		t.Fatalf("accepted calls = %d, want %d", got, want)
	}
}

// countAccepted records requests that passed an outer rate-limit middleware.
func countAccepted(count *atomic.Int32) server.Middleware {
	return func(next server.Handler) server.Handler {
		return server.HandlerFunc(func(ctx context.Context, request *server.Request) (server.Response, error) {
			count.Add(1)
			return next.Handle(ctx, request)
		})
	}
}

// mustNewE2ERateLimiter creates a real limiter for local UDP integration tests.
func mustNewE2ERateLimiter(t *testing.T, config RateLimitConfig) *RateLimiter {
	t.Helper()
	limiter, err := NewRateLimiter(config)
	if err != nil {
		t.Fatalf("NewRateLimiter() error = %v", err)
	}
	return limiter
}
