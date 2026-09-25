// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestStreamReturnsCompletionOrderAndClosesOutput(t *testing.T) {
	response := transport.AppendPacket(nil, transport.Packet{Type: transport.ResponseInfo})
	responder := newResponder(t, response, false)

	p, err := New(WithWorkers(1), WithTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	requests := make(chan Request, 2)
	requests <- Request{Target: responder.target, Query: transport.InfoRequest}
	requests <- Request{Target: responder.target, Query: transport.InfoRequest}
	close(requests)

	results := p.Stream(context.Background(), requests)
	var got []Result
	for result := range results {
		got = append(got, result)
	}

	if len(got) != 2 {
		t.Fatalf("received %d results, want 2", len(got))
	}
	for index, result := range got {
		if result.Index != uint64(index) {
			t.Fatalf("result %d index = %d", index, result.Index)
		}
		if result.Request.Query != transport.InfoRequest || result.Err != nil {
			t.Fatalf("result %d = %+v", index, result)
		}
		if result.Packet.Type != transport.ResponseInfo {
			t.Fatalf("result %d packet type = %X", index, result.Packet.Type)
		}
	}
}

func TestStreamReportsInvalidRequestWithoutStopping(t *testing.T) {
	response := transport.AppendPacket(nil, transport.Packet{Type: transport.ResponseInfo})
	responder := newResponder(t, response, false)

	p, err := New(WithWorkers(1), WithTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	requests := make(chan Request, 2)
	requests <- Request{}
	requests <- Request{Target: responder.target, Query: transport.InfoRequest}
	close(requests)

	results := p.Stream(context.Background(), requests)
	var invalid, valid bool
	for result := range results {
		if result.Index == 0 {
			invalid = errors.Is(result.Err, ErrInvalidTarget)
		}
		if result.Index == 1 {
			valid = result.Err == nil && result.Packet.Type == transport.ResponseInfo
		}
	}

	if !invalid || !valid {
		t.Fatalf("stream results did not preserve invalid and valid requests: invalid=%t valid=%t", invalid, valid)
	}
}

func TestStreamCancellationClosesOutputWithOpenInput(t *testing.T) {
	responder := newSilentResponder(t)
	p, err := New(WithWorkers(1), WithTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	requests := make(chan Request, 1)
	requests <- Request{Target: responder.target, Query: transport.InfoRequest}
	results := p.Stream(ctx, requests)
	responder.waitRequest(t)

	cancel()
	select {
	case _, ok := <-results:
		if ok {
			for range results {
			}
			t.Fatal("stream returned a result after cancellation")
		}
	case <-time.After(time.Second):
		t.Fatal("stream output did not close after cancellation")
	}
}

func TestStreamNilContextReturnsClosedOutput(t *testing.T) {
	p, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	results := p.Stream(nil, nil)
	if _, ok := <-results; ok {
		t.Fatal("nil-context stream returned a result")
	}
}

func TestStreamWindowIsBounded(t *testing.T) {
	response := transport.AppendPacket(nil, transport.Packet{Type: transport.ResponseInfo})
	responder := newResponder(t, response, false)
	p, err := New(WithWorkers(1), WithQueueSize(2))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	if p.window != 3 {
		t.Fatalf("stream window = %d, want 3", p.window)
	}
	ctx, cancel := context.WithCancel(context.Background())
	requests := make(chan Request, p.window)
	for range p.window {
		requests <- Request{Target: responder.target, Query: transport.InfoRequest}
	}
	close(requests)
	results := p.Stream(ctx, requests)
	for range p.window {
		responder.waitSourcePort(t)
	}

	queryCtx, queryCancel := context.WithTimeout(context.Background(), time.Second)
	defer queryCancel()
	if _, _, err := p.Query(queryCtx, responder.target, transport.InfoRequest); err != nil {
		t.Fatalf("pool query was blocked by unread stream output: %v", err)
	}

	cancel()
	for range results {
	}
}
