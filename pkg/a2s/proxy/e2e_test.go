// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

func TestProxyCachedRulesRoundTripAvoidsRelay(t *testing.T) {
	provider := &e2eChallengeProvider{token: a2s.Challenge{1, 2, 3, 4}}
	upstreamHandler := &e2eResponseHandler{packets: map[a2s.QueryType]a2s.Packet{
		a2s.RulesRequest: {
			Type:    a2s.ResponseRules,
			Payload: []byte("opaque A3SB rules payload"),
		},
	}}
	upstream := startE2EServer(t, upstreamHandler, provider, nil)
	upstreamClient := newE2EClient(t, upstream.addr)

	cached, _, err := upstreamClient.Query(context.Background(), a2s.RulesRequest)
	if err != nil {
		t.Fatalf("upstream Query() error = %v", err)
	}

	cache := mustCache(t, a2s.RulesRequest)
	if err := cache.Store(a2s.RulesRequest, cached, time.Now()); err != nil {
		t.Fatalf("cache.Store() error = %v", err)
	}

	relayClient := newE2EClient(t, upstream.addr)
	relay := &countingUpstream{client: relayClient}
	downstreamProvider := &e2eChallengeProvider{token: a2s.Challenge{5, 6, 7, 8}}
	handler := mustHandlerWithProvider(t, cache, relay, downstreamProvider, false)
	downstream := startE2EServer(t, handler, downstreamProvider, nil)
	downstreamClient := newE2EClient(t, downstream.addr)

	got, _, err := downstreamClient.Query(context.Background(), a2s.RulesRequest)
	if err != nil {
		t.Fatalf("downstream Query() error = %v", err)
	}
	if got.Type != cached.Type || !bytes.Equal(got.Payload, cached.Payload) {
		t.Fatalf("cached packet = %#v, want %#v", got, cached)
	}
	if got := relay.calls.Load(); got != 0 {
		t.Fatalf("relay calls = %d, want 0", got)
	}
	if got := upstreamHandler.Count(a2s.RulesRequest); got != 1 {
		t.Fatalf("upstream rules calls = %d, want 1", got)
	}
	if downstreamProvider.issued.Load() == 0 {
		t.Fatal("downstream challenge provider was not used")
	}
}

func TestProxyLiveRelayUsesIndependentChallengeFlows(t *testing.T) {
	upstreamProvider := &e2eChallengeProvider{token: a2s.Challenge{1, 2, 3, 4}}
	upstreamHandler := &e2eResponseHandler{packets: map[a2s.QueryType]a2s.Packet{
		a2s.RulesRequest: {
			Type:    a2s.ResponseRules,
			Payload: []byte("live rules"),
		},
	}}
	upstream := startE2EServer(t, upstreamHandler, upstreamProvider, nil)

	relayClient := newE2EClient(t, upstream.addr)
	relay := &countingUpstream{client: relayClient}
	cache := mustCache(t)
	downstreamProvider := &e2eChallengeProvider{token: a2s.Challenge{5, 6, 7, 8}}
	handler := mustHandlerWithProvider(t, cache, relay, downstreamProvider, false)
	downstream := startE2EServer(t, handler, downstreamProvider, nil)
	downstreamClient := newE2EClient(t, downstream.addr)

	got, _, err := downstreamClient.Query(context.Background(), a2s.RulesRequest)
	if err != nil {
		t.Fatalf("downstream Query() error = %v", err)
	}
	if string(got.Payload) != "live rules" {
		t.Fatalf("live payload = %q, want live rules", got.Payload)
	}
	if got := relay.calls.Load(); got != 1 {
		t.Fatalf("relay calls = %d, want 1", got)
	}
	if got := upstreamHandler.Count(a2s.RulesRequest); got != 1 {
		t.Fatalf("upstream rules calls = %d, want 1", got)
	}
	if upstreamProvider.issued.Load() == 0 || downstreamProvider.issued.Load() == 0 {
		t.Fatalf(
			"challenge providers were not both used: upstream=%d downstream=%d",
			upstreamProvider.issued.Load(),
			downstreamProvider.issued.Load(),
		)
	}
	if upstreamProvider.token == downstreamProvider.token {
		t.Fatal("upstream and downstream challenge tokens unexpectedly match")
	}
}

func TestProxyInvalidatesAndRecoversCachedQuery(t *testing.T) {
	provider := &e2eChallengeProvider{token: a2s.Challenge{1, 2, 3, 4}}
	upstreamHandler := &e2eResponseHandler{packets: map[a2s.QueryType]a2s.Packet{
		a2s.InfoRequest: {
			Type:    a2s.ResponseInfo,
			Payload: []byte("online info"),
		},
	}}
	upstream := startE2EServer(t, upstreamHandler, provider, nil)
	pollClient := newE2EClientWithTimeout(t, upstream.addr, 25*time.Millisecond)
	relayClient := newE2EClientWithTimeout(t, upstream.addr, 25*time.Millisecond)

	cache := mustCache(t, a2s.InfoRequest)
	relay := &countingUpstream{client: relayClient}
	downstreamProvider := &e2eChallengeProvider{token: a2s.Challenge{5, 6, 7, 8}}
	handler := mustHandlerWithProvider(t, cache, relay, downstreamProvider, false)
	downstream := startE2EServer(t, handler, downstreamProvider, nil)
	downstreamClient := newE2EClientWithTimeout(t, downstream.addr, 40*time.Millisecond)

	changes := make(chan StateChange, 2)
	poller := mustPoller(t, cache, pollClient, PollerConfig{
		TTL:           10 * time.Millisecond,
		InactiveTTL:   10 * time.Millisecond,
		OnStateChange: func(change StateChange) { changes <- change },
	})
	pollContext, cancelPoller := context.WithCancel(context.Background())
	pollDone := make(chan error, 1)
	go func() { pollDone <- poller.Run(pollContext) }()
	defer func() {
		cancelPoller()
		if err := <-pollDone; err != nil {
			t.Errorf("poller Run() error = %v", err)
		}
	}()

	waitForE2E(t, func() bool {
		_, ok := cache.Load(a2s.InfoRequest)
		return ok
	})

	upstreamHandler.SetOffline(true)
	waitForE2E(t, func() bool {
		for {
			select {
			case change := <-changes:
				if change.Query == a2s.InfoRequest && !change.Active {
					return true
				}
			default:
				return false
			}
		}
	})
	if _, _, err := downstreamClient.Query(context.Background(), a2s.InfoRequest); err == nil {
		t.Fatal("cached INFO query succeeded while cache was inactive")
	}
	if got := relay.calls.Load(); got != 0 {
		t.Fatalf("relay calls for cached INFO = %d, want 0", got)
	}

	upstreamHandler.SetOffline(false)
	waitForE2E(t, func() bool {
		for {
			select {
			case change := <-changes:
				if change.Query == a2s.InfoRequest && change.Active {
					return true
				}
			default:
				return false
			}
		}
	})
	got, _, err := downstreamClient.Query(context.Background(), a2s.InfoRequest)
	if err != nil {
		t.Fatalf("recovered downstream Query() error = %v", err)
	}
	if string(got.Payload) != "online info" {
		t.Fatalf("recovered payload = %q, want online info", got.Payload)
	}
}

func TestProxyPacketResponsePreservesSourceAndGoldSourceLogicalPackets(t *testing.T) {
	tests := []struct {
		name       string
		response   a2s.ResponseType
		packetizer func(*testing.T) server.Option
	}{
		{
			name:     "source",
			response: a2s.ResponseInfo,
			packetizer: func(t *testing.T) server.Option {
				packetizer, err := server.NewSourcePacketizer()
				if err != nil {
					t.Fatalf("NewSourcePacketizer() error = %v", err)
				}
				packetizer.SplitSize = 128
				return server.WithSourcePacketizer(packetizer)
			},
		},
		{
			name:     "goldsource",
			response: a2s.ResponseInfoGoldSource,
			packetizer: func(t *testing.T) server.Option {
				packetizer, err := server.NewGoldSourcePacketizer()
				if err != nil {
					t.Fatalf("NewGoldSourcePacketizer() error = %v", err)
				}
				packetizer.SplitSize = 128
				return server.WithGoldSourcePacketizer(packetizer)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := bytes.Repeat([]byte("logical packet "), 64)
			cache := mustCache(t, a2s.InfoRequest)
			if err := cache.Store(a2s.InfoRequest, a2s.Packet{
				Type:    test.response,
				Payload: payload,
			}, time.Now()); err != nil {
				t.Fatalf("cache.Store() error = %v", err)
			}

			provider := &e2eChallengeProvider{token: a2s.Challenge{5, 6, 7, 8}}
			handler := mustHandlerWithProvider(t, cache, &handlerUpstream{}, provider, false)
			downstream := startE2EServer(t, handler, provider, test.packetizer(t))
			client := newE2EClient(t, downstream.addr)

			got, _, err := client.Query(context.Background(), a2s.InfoRequest)
			if err != nil {
				t.Fatalf("Query() error = %v", err)
			}
			if got.Type != test.response || !bytes.Equal(got.Payload, payload) {
				t.Fatalf("packet = %#v, want type %x and %d payload bytes", got, test.response, len(payload))
			}
		})
	}
}

type e2eServer struct {
	server   *server.Server
	conn     net.PacketConn
	addr     *net.UDPAddr
	serveErr <-chan error
}

func startE2EServer(
	t *testing.T,
	handler server.Handler,
	provider server.ChallengeProvider,
	packetizer server.Option,
) *e2eServer {
	t.Helper()
	options := []server.Option{
		server.WithChallengeProvider(provider),
		server.WithWorkers(1),
	}
	if packetizer != nil {
		options = append(options, packetizer)
	}

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
		if err := <-serveErr; !errors.Is(err, server.ErrServerClosed) {
			t.Errorf("Serve() error = %v, want ErrServerClosed", err)
		}
		_ = conn.Close()
	})

	return e2e
}

func newE2EClient(t *testing.T, addr *net.UDPAddr) *a2s.Client {
	return newE2EClientWithTimeout(t, addr, time.Second)
}

func newE2EClientWithTimeout(t *testing.T, addr *net.UDPAddr, timeout time.Duration) *a2s.Client {
	t.Helper()
	client, err := a2s.NewWithAddr(addr, a2s.WithTimeout(timeout))
	if err != nil {
		t.Fatalf("a2s.NewWithAddr() error = %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func mustHandlerWithProvider(
	t *testing.T,
	cache *Cache,
	relay Upstream,
	provider server.ChallengeProvider,
	localPing bool,
) *Handler {
	t.Helper()
	handler, err := NewHandler(cache, relay, HandlerConfig{
		LocalPing:         localPing,
		ChallengeProvider: provider,
	})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	return handler
}

func waitForE2E(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer deadline.Stop()
	defer ticker.Stop()

	for {
		if condition() {
			return
		}
		select {
		case <-ticker.C:
		case <-deadline.C:
			t.Fatal("condition was not reached")
		}
	}
}

type countingUpstream struct {
	client *a2s.Client
	calls  atomic.Int32
}

func (u *countingUpstream) Query(ctx context.Context, query a2s.QueryType) (a2s.Packet, a2s.QueryMeta, error) {
	u.calls.Add(1)
	return u.client.Query(ctx, query)
}

type e2eResponseHandler struct {
	mu      sync.Mutex
	calls   map[a2s.QueryType]int
	packets map[a2s.QueryType]a2s.Packet
	offline atomic.Bool
}

func (h *e2eResponseHandler) Handle(_ context.Context, request *server.Request) (server.Response, error) {
	h.mu.Lock()
	if h.calls == nil {
		h.calls = make(map[a2s.QueryType]int)
	}
	h.calls[request.Query.Type]++
	packet, ok := h.packets[request.Query.Type]
	h.mu.Unlock()

	if h.offline.Load() || !ok {
		return nil, server.ErrDrop
	}

	return server.PacketResponse{Packet: packet}, nil
}

func (h *e2eResponseHandler) Count(query a2s.QueryType) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.calls[query]
}

func (h *e2eResponseHandler) SetOffline(offline bool) {
	h.offline.Store(offline)
}

type e2eChallengeProvider struct {
	token  a2s.Challenge
	issued atomic.Int32
}

func (p *e2eChallengeProvider) Issue(netip.AddrPort) a2s.Challenge {
	p.issued.Add(1)
	return p.token
}

func (p *e2eChallengeProvider) Validate(_ netip.AddrPort, challenge a2s.Challenge) bool {
	return challenge == p.token
}
