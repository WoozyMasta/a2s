// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func TestNewCacheEnablesSelectedQueries(t *testing.T) {
	cache, err := NewCache([]a2s.QueryType{
		a2s.InfoRequest,
		a2s.RulesRequest,
		a2s.InfoRequest,
	})
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	if !cache.Enabled(a2s.InfoRequest) || !cache.Enabled(a2s.RulesRequest) {
		t.Fatal("selected queries are not enabled")
	}
	if cache.Enabled(a2s.PlayerRequest) {
		t.Fatal("unselected query is enabled")
	}
}

func TestNewCacheRejectsUnsupportedQuery(t *testing.T) {
	_, err := NewCache([]a2s.QueryType{a2s.PingRequest})
	if !errors.Is(err, ErrCacheQuery) {
		t.Fatalf("NewCache() error = %v, want ErrCacheQuery", err)
	}
}

func TestCacheStoreAndLoadOwnPayload(t *testing.T) {
	cache, err := NewCache([]a2s.QueryType{a2s.InfoRequest})
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	payload := []byte("info")
	if err := cache.Store(a2s.InfoRequest, a2s.Packet{
		Type:    a2s.ResponseInfo,
		Payload: payload,
	}); err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	payload[0] = 'X'
	got, ok := cache.Load(a2s.InfoRequest)
	if !ok {
		t.Fatal("Load() reports missing entry")
	}
	if !bytes.Equal(got.Payload, []byte("info")) {
		t.Fatalf("Load() payload = %q, want info", got.Payload)
	}

	got.Payload[0] = 'Y'
	again, ok := cache.Load(a2s.InfoRequest)
	if !ok || !bytes.Equal(again.Payload, []byte("info")) {
		t.Fatalf("cache payload changed through loaded packet: %q", again.Payload)
	}
}

func TestCacheLoadReturnsOnlyValidEntries(t *testing.T) {
	cache, err := NewCache([]a2s.QueryType{
		a2s.InfoRequest,
		a2s.PlayerRequest,
	})
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	if _, ok := cache.Load(a2s.InfoRequest); ok {
		t.Fatal("empty cache returned an INFO entry")
	}
	if err := cache.Store(a2s.InfoRequest, infoPacket()); err != nil {
		t.Fatalf("Store(INFO) error = %v", err)
	}
	if err := cache.Invalidate(a2s.InfoRequest); err != nil {
		t.Fatalf("Invalidate(INFO) error = %v", err)
	}
	if _, ok := cache.Load(a2s.InfoRequest); ok {
		t.Fatal("invalidated INFO entry was returned")
	}
}

func TestCacheEntriesAreIndependent(t *testing.T) {
	cache, err := NewCache([]a2s.QueryType{
		a2s.InfoRequest,
		a2s.PlayerRequest,
	})
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	if err := cache.Store(a2s.InfoRequest, infoPacket()); err != nil {
		t.Fatalf("Store(INFO) error = %v", err)
	}
	if err := cache.Store(a2s.PlayerRequest, a2s.Packet{
		Type:    a2s.ResponsePlayers,
		Payload: []byte("players"),
	}); err != nil {
		t.Fatalf("Store(PLAYER) error = %v", err)
	}
	if err := cache.Invalidate(a2s.InfoRequest); err != nil {
		t.Fatalf("Invalidate(INFO) error = %v", err)
	}

	if _, ok := cache.Load(a2s.InfoRequest); ok {
		t.Fatal("invalidated INFO entry was returned")
	}
	if packet, ok := cache.Load(a2s.PlayerRequest); !ok || string(packet.Payload) != "players" {
		t.Fatalf("PLAYER entry = %#v, %v", packet, ok)
	}
}

func TestCacheRejectsDisabledAndMismatchedPackets(t *testing.T) {
	cache, err := NewCache([]a2s.QueryType{a2s.InfoRequest})
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	if err := cache.Store(a2s.PlayerRequest, a2s.Packet{Type: a2s.ResponsePlayers}); !errors.Is(err, ErrCacheDisabled) {
		t.Fatalf("Store(disabled) error = %v, want ErrCacheDisabled", err)
	}
	if err := cache.Store(a2s.InfoRequest, a2s.Packet{Type: a2s.ResponseRules}); !errors.Is(err, ErrCachePacket) {
		t.Fatalf("Store(mismatched) error = %v, want ErrCachePacket", err)
	}
	if err := cache.Invalidate(a2s.PingRequest); !errors.Is(err, ErrCacheQuery) {
		t.Fatalf("Invalidate(unsupported) error = %v, want ErrCacheQuery", err)
	}
}

func TestCacheConcurrentAccess(t *testing.T) {
	cache, err := NewCache([]a2s.QueryType{a2s.InfoRequest})
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			packet := a2s.Packet{
				Type:    a2s.ResponseInfo,
				Payload: []byte{byte(i)},
			}
			for j := 0; j < 100; j++ {
				if err := cache.Store(a2s.InfoRequest, packet); err != nil {
					t.Errorf("Store() error = %v", err)
					return
				}
				_, _ = cache.Load(a2s.InfoRequest)
				if err := cache.Invalidate(a2s.InfoRequest); err != nil {
					t.Errorf("Invalidate() error = %v", err)
					return
				}
			}
		}(i)
	}
	group.Wait()
}

func infoPacket() a2s.Packet {
	return a2s.Packet{
		Type:    a2s.ResponseInfo,
		Payload: []byte("info"),
	}
}
