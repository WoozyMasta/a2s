// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func TestWithMiddlewareRunsBeforeChallengeGate(t *testing.T) {
	trace := make([]string, 0, 3)
	provider := &middlewareTestProvider{trace: &trace}

	configured, err := New(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			trace = append(trace, "handler")
			return InfoResponse{}, nil
		}),
		WithChallengeProvider(provider),
		WithMiddleware(func(next Handler) Handler {
			return HandlerFunc(func(ctx context.Context, request *Request) (Response, error) {
				trace = append(trace, "middleware before")
				response, err := next.Handle(ctx, request)
				trace = append(trace, "middleware after")
				return response, err
			})
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	response, err := configured.Handler.Handle(context.Background(), testMiddlewareRequest())
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if _, ok := response.(PacketResponse); !ok {
		t.Fatalf("Handle() response = %T, want PacketResponse", response)
	}

	want := []string{"middleware before", "challenge", "middleware after"}
	if len(trace) != len(want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
	for index := range want {
		if trace[index] != want[index] {
			t.Fatalf("trace = %v, want %v", trace, want)
		}
	}
	if provider.issued != 1 {
		t.Fatalf("Issue() calls = %d, want 1", provider.issued)
	}
}

func TestWithMiddlewareCanDropBeforeChallengeGate(t *testing.T) {
	provider := &middlewareTestProvider{}
	handlerCalls := 0
	configured, err := New(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			handlerCalls++
			return InfoResponse{}, nil
		}),
		WithChallengeProvider(provider),
		WithMiddleware(func(Handler) Handler {
			return HandlerFunc(func(context.Context, *Request) (Response, error) {
				return nil, ErrDrop
			})
		}),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if _, err := configured.Handler.Handle(context.Background(), testMiddlewareRequest()); !errors.Is(err, ErrDrop) {
		t.Fatalf("Handle() error = %v, want ErrDrop", err)
	}
	if provider.issued != 0 {
		t.Fatalf("Issue() calls = %d, want 0", provider.issued)
	}
	if handlerCalls != 0 {
		t.Fatalf("handler calls = %d, want 0", handlerCalls)
	}
}

func TestWithMiddlewarePreservesDeclarationOrder(t *testing.T) {
	trace := make([]string, 0, 5)
	wrap := func(name string) Middleware {
		return func(next Handler) Handler {
			return HandlerFunc(func(ctx context.Context, request *Request) (Response, error) {
				trace = append(trace, name+" before")
				response, err := next.Handle(ctx, request)
				trace = append(trace, name+" after")
				return response, err
			})
		}
	}

	configured, err := New(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			trace = append(trace, "handler")
			return PacketResponse{Packet: a2s.Packet{Type: a2s.ResponsePing}}, nil
		}),
		WithChallengePolicy(NoChallengePolicy()),
		WithMiddleware(wrap("first")),
		WithMiddleware(wrap("second")),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if _, err := configured.Handler.Handle(context.Background(), &Request{
		Remote: netip.MustParseAddrPort("127.0.0.1:27015"),
		Query:  a2s.Request{Type: a2s.PingRequest},
	}); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}

	want := []string{"first before", "second before", "handler", "second after", "first after"}
	if len(trace) != len(want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
	for index := range want {
		if trace[index] != want[index] {
			t.Fatalf("trace = %v, want %v", trace, want)
		}
	}
}

func TestWithMiddlewareRejectsNilMiddleware(t *testing.T) {
	configured, err := New(
		HandlerFunc(func(context.Context, *Request) (Response, error) {
			return nil, nil
		}),
		WithMiddleware(nil),
	)
	if configured != nil {
		t.Fatal("New() returned a server for nil middleware")
	}
	if !errors.Is(err, ErrServer) {
		t.Fatalf("New() error = %v, want ErrServer", err)
	}
}

type middlewareTestProvider struct {
	trace  *[]string
	issued int
}

func (p *middlewareTestProvider) Issue(netip.AddrPort) a2s.Challenge {
	p.issued++
	if p.trace != nil {
		*p.trace = append(*p.trace, "challenge")
	}

	return a2s.Challenge{1, 2, 3, 4}
}

func (*middlewareTestProvider) Validate(netip.AddrPort, a2s.Challenge) bool {
	return false
}
