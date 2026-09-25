// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"context"
	"net/netip"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/pool"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

var _ Upstream = (*pool.Endpoint)(nil)

func TestPoolEndpointRefreshesPollerCache(t *testing.T) {
	provider := &e2eChallengeProvider{token: transport.Challenge{1, 2, 3, 4}}
	handler := &e2eResponseHandler{packets: map[transport.QueryType]transport.Packet{
		transport.InfoRequest: {
			Type:    transport.ResponseInfo,
			Payload: []byte("pooled info"),
		},
	}}
	upstream := startE2EServer(t, handler, provider, nil)

	sharedPool, err := pool.New(pool.WithWorkers(1), pool.WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("pool.New() error = %v", err)
	}
	defer func() {
		if err := sharedPool.Close(); err != nil {
			t.Errorf("Pool.Close() error = %v", err)
		}
	}()

	target := netip.MustParseAddrPort(upstream.addr.String())
	endpoint, err := sharedPool.BindAddr(target)
	if err != nil {
		t.Fatalf("Pool.BindAddr() error = %v", err)
	}

	cache := mustCache(t, transport.InfoRequest)
	poller, err := NewPoller(cache, endpoint, PollerConfig{TTL: time.Hour})
	if err != nil {
		t.Fatalf("NewPoller() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- poller.Run(ctx) }()

	waitForE2E(t, func() bool {
		packet, ok := cache.Load(transport.InfoRequest)
		return ok && string(packet.Payload) == "pooled info"
	})
	cancel()

	if err := <-done; err != nil {
		t.Fatalf("Poller.Run() error = %v", err)
	}
	if got := handler.Count(transport.InfoRequest); got != 1 {
		t.Fatalf("upstream INFO calls = %d, want 1", got)
	}
}

func TestPoolEndpointRelaysThroughProxyServer(t *testing.T) {
	upstreamProvider := &e2eChallengeProvider{token: transport.Challenge{1, 2, 3, 4}}
	upstreamHandler := &e2eResponseHandler{packets: map[transport.QueryType]transport.Packet{
		transport.InfoRequest: {
			Type:    transport.ResponseInfo,
			Payload: []byte("pooled relay"),
		},
	}}
	upstream := startE2EServer(t, upstreamHandler, upstreamProvider, nil)

	sharedPool, err := pool.New(pool.WithWorkers(1), pool.WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("pool.New() error = %v", err)
	}
	defer func() {
		if err := sharedPool.Close(); err != nil {
			t.Errorf("Pool.Close() error = %v", err)
		}
	}()

	endpoint, err := sharedPool.BindAddr(netip.MustParseAddrPort(upstream.addr.String()))
	if err != nil {
		t.Fatalf("Pool.BindAddr() error = %v", err)
	}

	cache := mustCache(t)
	downstreamProvider := &e2eChallengeProvider{token: transport.Challenge{5, 6, 7, 8}}
	handler, err := NewHandler(cache, endpoint, HandlerConfig{
		ChallengeProvider: downstreamProvider,
	})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	downstream := startE2EServer(t, handler, downstreamProvider, nil)

	client := newE2EClient(t, downstream.addr)
	got, _, err := client.Query(context.Background(), transport.InfoRequest)
	if err != nil {
		t.Fatalf("downstream Query() error = %v", err)
	}
	if got.Type != transport.ResponseInfo || string(got.Payload) != "pooled relay" {
		t.Fatalf("relayed packet = %#v, want pooled relay INFO", got)
	}
	if got := upstreamHandler.Count(transport.InfoRequest); got != 1 {
		t.Fatalf("upstream INFO calls = %d, want 1", got)
	}
	if upstreamProvider.issued.Load() == 0 || downstreamProvider.issued.Load() == 0 {
		t.Fatalf(
			"challenge providers were not both used: upstream=%d downstream=%d",
			upstreamProvider.issued.Load(),
			downstreamProvider.issued.Load(),
		)
	}
}
