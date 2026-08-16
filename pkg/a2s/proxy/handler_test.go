// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package proxy

import (
	"context"
	"errors"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

func TestNewHandlerValidatesConfiguration(t *testing.T) {
	cache := mustCache(t, a2s.InfoRequest)
	relay := &handlerUpstream{}
	provider := &handlerChallengeProvider{}

	tests := []struct {
		name   string
		cache  *Cache
		relay  Upstream
		config HandlerConfig
	}{
		{name: "nil cache", relay: relay, config: HandlerConfig{ChallengeProvider: provider}},
		{name: "nil relay", cache: cache, config: HandlerConfig{ChallengeProvider: provider}},
		{name: "nil provider", cache: cache, relay: relay},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewHandler(test.cache, test.relay, test.config); !errors.Is(err, ErrHandler) {
				t.Fatalf("NewHandler() error = %v, want ErrHandler", err)
			}
		})
	}
}

func TestHandlerServesCachedQueriesWithoutRelay(t *testing.T) {
	tests := []struct {
		name     string
		query    a2s.QueryType
		response a2s.ResponseType
		payload  []byte
	}{
		{name: "info", query: a2s.InfoRequest, response: a2s.ResponseInfo, payload: []byte("info")},
		{name: "players", query: a2s.PlayerRequest, response: a2s.ResponsePlayers, payload: []byte("players")},
		{name: "rules", query: a2s.RulesRequest, response: a2s.ResponseRules, payload: []byte("opaque A3SB rules")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := mustCache(t, test.query)
			if err := cache.Store(test.query, a2s.Packet{
				Type:    test.response,
				Payload: test.payload,
			}, time.Time{}); err != nil {
				t.Fatalf("Store() error = %v", err)
			}
			relay := &handlerUpstream{}
			handler := mustHandler(t, cache, relay, false)

			response, err := handler.Handle(context.Background(), requestFor(test.query))
			if err != nil {
				t.Fatalf("Handle() error = %v", err)
			}
			packet := responsePacket(t, response)
			if packet.Type != test.response || string(packet.Payload) != string(test.payload) {
				t.Fatalf("packet = %#v, want type %x payload %q", packet, test.response, test.payload)
			}
			if got := relay.Count(test.query); got != 0 {
				t.Fatalf("relay calls = %d, want 0", got)
			}
		})
	}
}

func TestHandlerDropsMissingCachedQuery(t *testing.T) {
	cache := mustCache(t, a2s.RulesRequest)
	relay := &handlerUpstream{}
	handler := mustHandler(t, cache, relay, false)

	response, err := handler.Handle(context.Background(), requestFor(a2s.RulesRequest))
	if !errors.Is(err, server.ErrDrop) || response != nil {
		t.Fatalf("Handle() = %#v, %v; want dropped response", response, err)
	}
	if got := relay.Count(a2s.RulesRequest); got != 0 {
		t.Fatalf("relay calls = %d, want 0", got)
	}
}

func TestHandlerRelaysUncachedQueries(t *testing.T) {
	queries := []struct {
		query    a2s.QueryType
		response a2s.ResponseType
	}{
		{query: a2s.InfoRequest, response: a2s.ResponseInfo},
		{query: a2s.PlayerRequest, response: a2s.ResponsePlayers},
		{query: a2s.RulesRequest, response: a2s.ResponseRules},
	}
	relay := &handlerUpstream{}
	handler := mustHandler(t, mustCache(t), relay, false)

	for _, test := range queries {
		response, err := handler.Handle(context.Background(), requestFor(test.query))
		if err != nil {
			t.Fatalf("Handle(0x%X) error = %v", test.query, err)
		}
		if packet := responsePacket(t, response); packet.Type != test.response {
			t.Fatalf("Handle(0x%X) response type = %x, want %x", test.query, packet.Type, test.response)
		}
		if got := relay.Count(test.query); got != 1 {
			t.Fatalf("relay calls for 0x%X = %d, want 1", test.query, got)
		}
	}
}

func TestHandlerDropsRelayFailure(t *testing.T) {
	relay := &handlerUpstream{err: errors.New("upstream unavailable")}
	handler := mustHandler(t, mustCache(t), relay, false)

	response, err := handler.Handle(context.Background(), requestFor(a2s.InfoRequest))
	if !errors.Is(err, server.ErrDrop) || response != nil {
		t.Fatalf("Handle() = %#v, %v; want dropped response", response, err)
	}
}

func TestHandlerAnswersGetChallengeLocally(t *testing.T) {
	provider := &handlerChallengeProvider{challenge: a2s.Challenge{1, 2, 3, 4}}
	relay := &handlerUpstream{}
	handler, err := NewHandler(mustCache(t), relay, HandlerConfig{ChallengeProvider: provider})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	request := requestFor(a2s.ChallengeRequest)
	response, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	packet := responsePacket(t, response)
	if packet.Type != a2s.ResponseChallenge || string(packet.Payload) != string(provider.challenge[:]) {
		t.Fatalf("challenge packet = %#v, want %x", packet, provider.challenge)
	}
	if provider.remote != request.Remote {
		t.Fatalf("provider remote = %s, want %s", provider.remote, request.Remote)
	}
	if got := relay.Count(a2s.ChallengeRequest); got != 0 {
		t.Fatalf("relay calls = %d, want 0", got)
	}
}

func TestHandlerRelaysPingByDefault(t *testing.T) {
	relay := &handlerUpstream{}
	handler := mustHandler(t, mustCache(t), relay, false)

	response, err := handler.Handle(context.Background(), requestFor(a2s.PingRequest))
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if packet := responsePacket(t, response); packet.Type != a2s.ResponsePing {
		t.Fatalf("PING response type = %x, want %x", packet.Type, a2s.ResponsePing)
	}
	if got := relay.Count(a2s.PingRequest); got != 1 {
		t.Fatalf("relay calls = %d, want 1", got)
	}
}

func TestHandlerAnswersLocalPingWithoutRelay(t *testing.T) {
	relay := &handlerUpstream{}
	handler := mustHandler(t, mustCache(t), relay, true)

	response, err := handler.Handle(context.Background(), requestFor(a2s.PingRequest))
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	packet := responsePacket(t, response)
	if packet.Type != a2s.ResponsePing || len(packet.Payload) != 0 {
		t.Fatalf("local PING packet = %#v, want empty ACK", packet)
	}
	if got := relay.Count(a2s.PingRequest); got != 0 {
		t.Fatalf("relay calls = %d, want 0", got)
	}
}

func TestHandlerDropsUnknownRequest(t *testing.T) {
	handler := mustHandler(t, mustCache(t), &handlerUpstream{}, false)
	request := requestFor(a2s.QueryType(0xFE))

	response, err := handler.Handle(context.Background(), request)
	if !errors.Is(err, server.ErrDrop) || response != nil {
		t.Fatalf("Handle() = %#v, %v; want dropped response", response, err)
	}
}

func mustHandler(t *testing.T, cache *Cache, relay Upstream, localPing bool) *Handler {
	t.Helper()
	handler, err := NewHandler(cache, relay, HandlerConfig{
		LocalPing:         localPing,
		ChallengeProvider: &handlerChallengeProvider{},
	})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	return handler
}

func requestFor(query a2s.QueryType) *server.Request {
	return &server.Request{
		Remote: netip.MustParseAddrPort("127.0.0.1:27015"),
		Query:  a2s.Request{Type: query},
	}
}

func responsePacket(t *testing.T, response server.Response) a2s.Packet {
	t.Helper()
	packetResponse, ok := response.(server.PacketResponse)
	if !ok {
		t.Fatalf("response = %T, want server.PacketResponse", response)
	}
	return packetResponse.Packet
}

type handlerUpstream struct {
	mu    sync.Mutex
	calls map[a2s.QueryType]int
	err   error
}

func (u *handlerUpstream) Query(_ context.Context, query a2s.QueryType) (a2s.Packet, a2s.QueryMeta, error) {
	u.mu.Lock()
	if u.calls == nil {
		u.calls = make(map[a2s.QueryType]int)
	}
	u.calls[query]++
	u.mu.Unlock()

	if u.err != nil {
		return a2s.Packet{}, a2s.QueryMeta{}, u.err
	}

	return responsePacketFor(query), a2s.QueryMeta{}, nil
}

func (u *handlerUpstream) Count(query a2s.QueryType) int {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.calls[query]
}

func responsePacketFor(query a2s.QueryType) a2s.Packet {
	switch query {
	case a2s.InfoRequest:
		return a2s.Packet{Type: a2s.ResponseInfo, Payload: []byte("info")}
	case a2s.PlayerRequest:
		return a2s.Packet{Type: a2s.ResponsePlayers, Payload: []byte("players")}
	case a2s.RulesRequest:
		return a2s.Packet{Type: a2s.ResponseRules, Payload: []byte("rules")}
	case a2s.PingRequest:
		return a2s.Packet{Type: a2s.ResponsePing, Payload: []byte("pong\x00")}
	default:
		return a2s.Packet{}
	}
}

type handlerChallengeProvider struct {
	challenge a2s.Challenge
	remote    netip.AddrPort
}

func (p *handlerChallengeProvider) Issue(remote netip.AddrPort) a2s.Challenge {
	p.remote = remote
	return p.challenge
}

func (*handlerChallengeProvider) Validate(netip.AddrPort, a2s.Challenge) bool {
	return true
}
