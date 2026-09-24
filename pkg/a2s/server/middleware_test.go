// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import (
	"context"
	"errors"
	"net/netip"
	"sync"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func TestMiddlewareModifiesTypedResponse(t *testing.T) {
	base := HandlerFunc(func(context.Context, *Request) (Response, error) {
		return InfoResponse{Info: a2s.Info{Name: "original"}}, nil
	})
	modifier := Middleware(func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, request *Request) (Response, error) {
			response, err := next.Handle(ctx, request)
			if err != nil {
				return nil, err
			}

			info := response.(InfoResponse)
			info.Info.Name += " modified"
			return info, nil
		})
	})

	response, err := modifier(base).Handle(context.Background(), testMiddlewareRequest())
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	info, ok := response.(InfoResponse)
	if !ok || info.Info.Name != "original modified" {
		t.Fatalf("response = %#v, want modified InfoResponse", response)
	}
}

func TestMiddlewareImplementsSimpleInfoCache(t *testing.T) {
	var calls int
	base := HandlerFunc(func(context.Context, *Request) (Response, error) {
		calls++
		return InfoResponse{Info: a2s.Info{Name: "cached"}}, nil
	})

	var mu sync.Mutex
	var cached Response
	var hasCached bool
	cache := Middleware(func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, request *Request) (Response, error) {
			mu.Lock()
			if hasCached {
				response := cached
				mu.Unlock()
				return response, nil
			}
			mu.Unlock()

			response, err := next.Handle(ctx, request)
			if err != nil {
				return nil, err
			}
			mu.Lock()
			cached = response
			hasCached = true
			mu.Unlock()
			return response, nil
		})
	})

	handler := cache(base)
	for range 2 {
		response, err := handler.Handle(context.Background(), testMiddlewareRequest())
		if err != nil {
			t.Fatalf("Handle() error = %v", err)
		}
		if response.(InfoResponse).Info.Name != "cached" {
			t.Fatalf("cached response = %#v", response)
		}
	}
	if calls != 1 {
		t.Fatalf("base handler calls = %d, want 1", calls)
	}
}

func TestMiddlewareCanDropUnauthorizedRequest(t *testing.T) {
	var calls int
	base := HandlerFunc(func(context.Context, *Request) (Response, error) {
		calls++
		return InfoResponse{Info: a2s.Info{Name: "allowed"}}, nil
	})
	allowed := netip.MustParseAddrPort("127.0.0.1:27016")
	acl := Middleware(func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, request *Request) (Response, error) {
			if request.Remote != allowed {
				return nil, ErrDrop
			}
			return next.Handle(ctx, request)
		})
	})

	if _, err := acl(base).Handle(context.Background(), &Request{Remote: netip.MustParseAddrPort("127.0.0.1:1"), Query: a2s.Request{Type: a2s.InfoRequest}}); !errors.Is(err, ErrDrop) {
		t.Fatalf("unauthorized error = %v, want ErrDrop", err)
	}
	response, err := acl(base).Handle(context.Background(), &Request{Remote: allowed, Query: a2s.Request{Type: a2s.InfoRequest}})
	if err != nil {
		t.Fatalf("authorized error = %v", err)
	}
	if response.(InfoResponse).Info.Name != "allowed" || calls != 1 {
		t.Fatalf("authorized response = %#v, calls = %d", response, calls)
	}
}

func TestMiddlewarePreservesCompositionOrder(t *testing.T) {
	var trace []string
	base := HandlerFunc(func(context.Context, *Request) (Response, error) {
		trace = append(trace, "handler")
		return InfoResponse{}, nil
	})
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

	handler := wrap("outer")(wrap("inner")(base))
	if _, err := handler.Handle(context.Background(), testMiddlewareRequest()); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	want := []string{"outer before", "inner before", "handler", "inner after", "outer after"}
	if len(trace) != len(want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
	for index := range want {
		if trace[index] != want[index] {
			t.Fatalf("trace = %v, want %v", trace, want)
		}
	}
}

func testMiddlewareRequest() *Request {
	return &Request{Remote: netip.MustParseAddrPort("127.0.0.1:27016"), Query: a2s.Request{Type: a2s.InfoRequest}}
}
