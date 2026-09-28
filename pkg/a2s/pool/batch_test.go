// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"sync/atomic"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestPoolBatchRetriesFailedRequestsInNextPass(t *testing.T) {
	response := transport.AppendPacket(nil, transport.Packet{Type: transport.ResponseInfo})
	flaky := newBatchFlakyResponder(t, response)
	stable := newResponder(t, response, false)

	p, err := New(
		WithWorkers(2),
		WithQueueSize(2),
		WithTimeout(50*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	batch, err := p.Batch(context.Background(), []Request{
		{Target: flaky.target, Query: transport.InfoRequest},
		{Target: stable.target, Query: transport.InfoRequest},
	}, BatchOptions{BonusPasses: 1})
	if err != nil {
		t.Fatalf("Batch() error = %v", err)
	}
	if batch.Recovered != 1 {
		t.Fatalf("Batch() recovered = %d, want 1", batch.Recovered)
	}
	if len(batch.Results) != 2 {
		t.Fatalf("Batch() results = %d, want 2", len(batch.Results))
	}
	for index, result := range batch.Results {
		if result.Index != uint64(index) {
			t.Fatalf("result[%d].Index = %d, want %d", index, result.Index, index)
		}
		if result.Err != nil {
			t.Fatalf("result[%d].Err = %v", index, result.Err)
		}
	}
	if attempts := flaky.requests.Load(); attempts != 2 {
		t.Fatalf("flaky responder requests = %d, want 2", attempts)
	}
}

func TestPoolBatchRetryPredicateCanRejectFailure(t *testing.T) {
	responder := newSilentResponder(t)
	p, err := New(WithWorkers(1), WithTimeout(20*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = p.Close() }()

	var predicateCalls atomic.Int32
	batch, err := p.Batch(context.Background(), []Request{{
		Target: responder.target,
		Query:  transport.InfoRequest,
	}}, BatchOptions{
		BonusPasses: 2,
		Retry: func(error) bool {
			predicateCalls.Add(1)
			return false
		},
	})
	if err != nil {
		t.Fatalf("Batch() error = %v", err)
	}
	if len(batch.Results) != 1 {
		t.Fatalf("Batch() results = %d, want 1", len(batch.Results))
	}
	if !errors.Is(batch.Results[0].Err, context.DeadlineExceeded) {
		t.Fatalf("Batch() result error = %v, want deadline", batch.Results[0].Err)
	}
	if got := predicateCalls.Load(); got != 1 {
		t.Fatalf("retry predicate calls = %d, want 1", got)
	}
}

// batchFlakyResponder ignores the first request and answers subsequent ones.
type batchFlakyResponder struct {
	target   netip.AddrPort
	conn     *net.UDPConn
	response []byte
	requests atomic.Int32
}

// newBatchFlakyResponder creates a responder for retry-pass tests.
func newBatchFlakyResponder(t *testing.T, response []byte) *batchFlakyResponder {
	t.Helper()
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		t.Fatal(err)
	}

	responder := &batchFlakyResponder{
		target:   netip.MustParseAddrPort(conn.LocalAddr().String()),
		conn:     conn,
		response: response,
	}
	t.Cleanup(func() { _ = conn.Close() })

	go responder.serve()
	return responder
}

// serve ignores the first request and responds to every later request.
func (r *batchFlakyResponder) serve() {
	buffer := make([]byte, 2048)
	for {
		n, remote, err := r.conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}
		if r.requests.Add(1) == 1 {
			continue
		}
		_, _ = r.conn.WriteToUDP(r.response, remote)
	}
}
