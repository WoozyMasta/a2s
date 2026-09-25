// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

var (
	benchmarkPacketSink   transport.Packet
	benchmarkResponseSink server.Response
)

func BenchmarkCacheLoadInfo(b *testing.B) {
	cache := benchmarkCache(b, transport.InfoRequest, []byte("benchmark info"))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		packet, ok := cache.Load(transport.InfoRequest)
		if !ok {
			b.Fatal("Load() returned no INFO packet")
		}
		benchmarkPacketSink = packet
	}
}

func BenchmarkCacheLoadRulesLarge(b *testing.B) {
	cache := benchmarkCache(b, transport.RulesRequest, bytes.Repeat([]byte("rule=value\x00"), 512))

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		packet, ok := cache.Load(transport.RulesRequest)
		if !ok {
			b.Fatal("Load() returned no RULES packet")
		}
		benchmarkPacketSink = packet
	}
}

func BenchmarkHandlerCachedHit(b *testing.B) {
	for _, test := range []struct {
		name    string
		query   transport.QueryType
		payload []byte
	}{
		{name: "info", query: transport.InfoRequest, payload: []byte("benchmark info")},
		{
			name:    "rules-large",
			query:   transport.RulesRequest,
			payload: bytes.Repeat([]byte("rule=value\x00"), 512),
		},
	} {
		b.Run(test.name, func(b *testing.B) {
			cache := benchmarkCache(b, test.query, test.payload)
			handler, err := NewHandler(cache, &handlerUpstream{}, HandlerConfig{
				ChallengeProvider: &handlerChallengeProvider{},
			})
			if err != nil {
				b.Fatalf("NewHandler() error = %v", err)
			}

			request := requestFor(test.query)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				response, err := handler.Handle(context.Background(), request)
				if err != nil {
					b.Fatal(err)
				}
				benchmarkResponseSink = response
			}
		})
	}
}

func BenchmarkCachePublication(b *testing.B) {
	packet := transport.Packet{Type: transport.ResponseInfo, Payload: []byte("benchmark info")}

	b.Run("store", func(b *testing.B) {
		cache, err := NewCache([]transport.QueryType{transport.InfoRequest})
		if err != nil {
			b.Fatalf("NewCache() error = %v", err)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			if err := cache.Store(transport.InfoRequest, packet); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("invalidate", func(b *testing.B) {
		cache := benchmarkCache(b, transport.InfoRequest, packet.Payload)

		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			if err := cache.Invalidate(transport.InfoRequest); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkCachedLocalRoundTrip(b *testing.B) {
	cache := benchmarkCache(b, transport.InfoRequest, []byte("benchmark info"))
	handler, err := NewHandler(cache, &handlerUpstream{}, HandlerConfig{
		ChallengeProvider: &handlerChallengeProvider{},
	})
	if err != nil {
		b.Fatalf("NewHandler() error = %v", err)
	}

	instance, err := server.New(
		handler,
		server.WithChallengePolicy(server.NoChallengePolicy()),
		server.WithWorkers(1),
	)
	if err != nil {
		b.Fatalf("server.New() error = %v", err)
	}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("ListenPacket() error = %v", err)
	}

	serveErr := make(chan error, 1)
	go func() { serveErr <- instance.Serve(conn) }()

	client, err := a2s.NewWithAddr(conn.LocalAddr().(*net.UDPAddr), a2s.WithTimeout(time.Second))
	if err != nil {
		_ = conn.Close()
		_ = instance.Shutdown(context.Background())
		<-serveErr
		b.Fatalf("NewWithAddr() error = %v", err)
	}
	b.Cleanup(func() {
		_ = client.Close()
		_ = instance.Shutdown(context.Background())
		_ = conn.Close()
		<-serveErr
	})

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		packet, _, err := client.Query(context.Background(), transport.InfoRequest)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkPacketSink = packet
	}
}

func benchmarkCache(b *testing.B, query transport.QueryType, payload []byte) *Cache {
	b.Helper()

	cache, err := NewCache([]transport.QueryType{query})
	if err != nil {
		b.Fatalf("NewCache() error = %v", err)
	}

	response := transport.ResponseInfo
	if query == transport.RulesRequest {
		response = transport.ResponseRules
	}
	if query == transport.PlayerRequest {
		response = transport.ResponsePlayers
	}
	if err := cache.Store(query, transport.Packet{Type: response, Payload: payload}); err != nil {
		b.Fatalf("Store() error = %v", err)
	}

	return cache
}
