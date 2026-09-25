// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s/server"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestHandlerAllowsOnlyOneLiveRelay(t *testing.T) {
	relay := newBlockingUpstream()
	handler := mustHandler(t, mustCache(t), relay, false)
	firstDone := make(chan error, 1)

	go func() {
		_, err := handler.Handle(context.Background(), requestFor(transport.InfoRequest))
		firstDone <- err
	}()

	waitForRelayStart(t, relay)
	if _, err := handler.Handle(context.Background(), requestFor(transport.InfoRequest)); !errors.Is(err, server.ErrDrop) {
		t.Fatalf("concurrent relay error = %v, want ErrDrop", err)
	}
	if got := relay.calls.Load(); got != 1 {
		t.Fatalf("relay calls = %d, want 1", got)
	}

	close(relay.release)
	if err := <-firstDone; err != nil {
		t.Fatalf("first relay error = %v", err)
	}

	if _, err := handler.Handle(context.Background(), requestFor(transport.InfoRequest)); err != nil {
		t.Fatalf("relay after release error = %v", err)
	}
	if got := relay.calls.Load(); got != 2 {
		t.Fatalf("relay calls after release = %d, want 2", got)
	}
}

func TestHandlerKeepsCachedAndLocalResponsesAvailableDuringRelay(t *testing.T) {
	relay := newBlockingUpstream()
	cache := mustCache(t, transport.InfoRequest)
	if err := cache.Store(transport.InfoRequest, transport.Packet{
		Type:    transport.ResponseInfo,
		Payload: []byte("cached info"),
	}); err != nil {
		t.Fatalf("cache.Store() error = %v", err)
	}
	handler := mustHandler(t, cache, relay, true)
	firstDone := make(chan error, 1)

	go func() {
		_, err := handler.Handle(context.Background(), requestFor(transport.PlayerRequest))
		firstDone <- err
	}()
	waitForRelayStart(t, relay)

	response, err := handler.Handle(context.Background(), requestFor(transport.InfoRequest))
	if err != nil {
		t.Fatalf("cached INFO error = %v", err)
	}
	if packet := responsePacket(t, response); string(packet.Payload) != "cached info" {
		t.Fatalf("cached INFO payload = %q, want cached info", packet.Payload)
	}

	response, err = handler.Handle(context.Background(), requestFor(transport.ChallengeRequest))
	if err != nil {
		t.Fatalf("local challenge error = %v", err)
	}
	if packet := responsePacket(t, response); packet.Type != transport.ResponseChallenge {
		t.Fatalf("local challenge type = %X, want %X", packet.Type, transport.ResponseChallenge)
	}

	response, err = handler.Handle(context.Background(), requestFor(transport.PingRequest))
	if err != nil {
		t.Fatalf("local PING error = %v", err)
	}
	if packet := responsePacket(t, response); packet.Type != transport.ResponsePing {
		t.Fatalf("local PING type = %X, want %X", packet.Type, transport.ResponsePing)
	}
	if got := relay.calls.Load(); got != 1 {
		t.Fatalf("relay calls during local responses = %d, want 1", got)
	}

	close(relay.release)
	if err := <-firstDone; err != nil {
		t.Fatalf("first relay error = %v", err)
	}
}

type blockingUpstream struct {
	started  chan struct{}
	release  chan struct{}
	startOne sync.Once
	calls    atomic.Int32
}

func newBlockingUpstream() *blockingUpstream {
	return &blockingUpstream{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
}

func (u *blockingUpstream) Query(ctx context.Context, query transport.QueryType) (transport.Packet, transport.QueryMeta, error) {
	u.calls.Add(1)
	u.startOne.Do(func() { close(u.started) })
	select {
	case <-u.release:
		return responsePacketFor(query), transport.QueryMeta{}, nil
	case <-ctx.Done():
		return transport.Packet{}, transport.QueryMeta{}, ctx.Err()
	}
}

func waitForRelayStart(t *testing.T, relay *blockingUpstream) {
	t.Helper()
	select {
	case <-relay.started:
	case <-time.After(time.Second):
		t.Fatal("live relay did not start")
	}
}
