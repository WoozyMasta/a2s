// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"context"
	"errors"
	"net/netip"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

var _ interface {
	Query(context.Context, transport.QueryType) (transport.Packet, transport.QueryMeta, error)
} = (*Endpoint)(nil)

func TestBindResolvesOnceAndEndpointReusesTarget(t *testing.T) {
	response := transport.AppendPacket(nil, transport.Packet{Type: transport.ResponseInfo})
	responder := newResponder(t, response, false)

	p, err := New(WithWorkers(1), WithTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	var calls atomic.Int32
	p.resolver = func(context.Context, string) ([]netip.Addr, error) {
		calls.Add(1)
		return []netip.Addr{responder.target.Addr()}, nil
	}

	address := "example.test:" + strconv.Itoa(int(responder.target.Port()))
	endpoint, err := p.Bind(context.Background(), address)
	if err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("resolver calls after Bind() = %d, want 1", got)
	}

	if _, _, err := endpoint.Query(context.Background(), transport.InfoRequest); err != nil {
		t.Fatalf("Endpoint.Query() error = %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("resolver calls after Endpoint.Query() = %d, want 1", got)
	}
}

func TestBindLiteralSkipsResolver(t *testing.T) {
	responder := newResponder(t, transport.AppendPacket(nil, transport.Packet{
		Type: transport.ResponseInfo,
	}), false)
	p, err := New(WithTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	var calls atomic.Int32
	p.resolver = func(context.Context, string) ([]netip.Addr, error) {
		calls.Add(1)
		return nil, errors.New("resolver must not be called")
	}

	endpoint, err := p.Bind(context.Background(), responder.target.String())
	if err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if calls.Load() != 0 {
		t.Fatal("literal Bind() called the resolver")
	}
	if _, _, err := endpoint.Query(context.Background(), transport.InfoRequest); err != nil {
		t.Fatalf("Endpoint.Query() error = %v", err)
	}
}

func TestBindRejectsInvalidTarget(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	for _, address := range []string{"", "example.test", "example.test:0", "example.test:not-a-port"} {
		t.Run(address, func(t *testing.T) {
			if _, err := p.Bind(context.Background(), address); !errors.Is(err, ErrInvalidTarget) {
				t.Fatalf("Bind(%q) error = %v, want ErrInvalidTarget", address, err)
			}
		})
	}

	for _, target := range []netip.AddrPort{
		{},
		netip.AddrPortFrom(netip.IPv4Unspecified(), 27015),
		netip.AddrPortFrom(netip.MustParseAddr("127.0.0.1"), 0),
	} {
		if _, err := p.BindAddr(target); !errors.Is(err, ErrInvalidTarget) {
			t.Fatalf("BindAddr(%v) error = %v, want ErrInvalidTarget", target, err)
		}
	}
}

func TestBindPropagatesResolverContext(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	p.resolver = func(ctx context.Context, _ string) ([]netip.Addr, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = p.Bind(ctx, "example.test:27015")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Bind() error = %v, want context.Canceled", err)
	}
}
