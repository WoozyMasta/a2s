// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"context"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func BenchmarkPoolInfoWorkers1(b *testing.B) {
	benchmarkPoolInfo(b, 1)
}

func BenchmarkPoolInfoWorkers8(b *testing.B) {
	benchmarkPoolInfo(b, 8)
}

func BenchmarkPoolInfoWorkers32(b *testing.B) {
	benchmarkPoolInfo(b, 32)
}

func BenchmarkPoolInfoWorkers64(b *testing.B) {
	benchmarkPoolInfo(b, 64)
}

func BenchmarkPoolMixedSuccessTimeout(b *testing.B) {
	response := transport.AppendPacket(nil, transport.Packet{Type: transport.ResponseInfo})
	success := newPoolBenchmarkResponder(b, response)
	timeout := newPoolBenchmarkResponder(b, nil)
	p := newPoolBenchmark(b, 8, 10*time.Millisecond)

	b.ReportAllocs()
	b.ResetTimer()
	var index uint64
	for b.Loop() {
		target := timeout
		wantSuccess := index%2 == 0
		if wantSuccess {
			target = success
		}

		_, _, err := p.Query(context.Background(), target, transport.InfoRequest)
		if (err == nil) != wantSuccess {
			b.Fatalf("mixed query %d: error = %v, want success = %t", index, err, wantSuccess)
		}
		index++
	}
	b.StopTimer()
	b.ReportMetric(50, "success_percent")
}

func BenchmarkPoolStream(b *testing.B) {
	const (
		workers   = 8
		batchSize = 32
	)

	response := transport.AppendPacket(nil, transport.Packet{Type: transport.ResponseInfo})
	targets := make([]netip.AddrPort, batchSize)
	for index := range targets {
		targets[index] = newPoolBenchmarkResponder(b, response)
	}
	p := newPoolBenchmark(b, workers, time.Second)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		requests := make(chan Request, batchSize)
		for _, target := range targets {
			requests <- Request{Target: target, Query: transport.InfoRequest}
		}
		close(requests)

		results := p.Stream(context.Background(), requests)
		count := 0
		for result := range results {
			if result.Err != nil || result.Packet.Type != transport.ResponseInfo {
				b.Fatalf("stream result = %+v", result)
			}
			count++
		}
		if count != batchSize {
			b.Fatalf("stream results = %d, want %d", count, batchSize)
		}
	}
	b.StopTimer()
	b.ReportMetric(batchSize, "queries/op")
}

// benchmarkPoolInfo measures concurrent successful INFO throughput.
func benchmarkPoolInfo(b *testing.B, workers int) {
	const targetCount = 64

	response := transport.AppendPacket(nil, transport.Packet{Type: transport.ResponseInfo})
	targets := make([]netip.AddrPort, targetCount)
	for index := range targets {
		targets[index] = newPoolBenchmarkResponder(b, response)
	}
	p := newPoolBenchmark(b, workers, time.Second)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		index := 0
		for pb.Next() {
			if _, _, err := p.Query(
				context.Background(),
				targets[index%len(targets)],
				transport.InfoRequest,
			); err != nil {
				b.Errorf("Pool.Query() error = %v", err)
				return
			}
			index++
		}
	})
	b.StopTimer()
	b.ReportMetric(float64(workers), "workers")
}

// newPoolBenchmark creates a pool with the benchmark's explicit budget.
func newPoolBenchmark(b *testing.B, workers int, timeout time.Duration) *Pool {
	b.Helper()
	p, err := New(
		WithWorkers(workers),
		WithTimeout(timeout),
	)
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		if err := p.Close(); err != nil {
			b.Errorf("Pool.Close() error = %v", err)
		}
	})

	return p
}

// newPoolBenchmarkResponder starts a local UDP responder for one target.
func newPoolBenchmarkResponder(b *testing.B, response []byte) netip.AddrPort {
	b.Helper()
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1")})
	if err != nil {
		b.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		buffer := make([]byte, 2048)
		for {
			_, remote, err := conn.ReadFromUDP(buffer)
			if err != nil {
				return
			}
			if len(response) == 0 {
				continue
			}
			if _, err := conn.WriteToUDP(response, remote); err != nil {
				return
			}
		}
	}()

	b.Cleanup(func() {
		_ = conn.Close()
		<-done
	})
	return netip.MustParseAddrPort(conn.LocalAddr().String())
}
