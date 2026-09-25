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

	"github.com/woozymasta/a2s/pkg/a2s/server"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestNewHandlerValidatesConfiguration(t *testing.T) {
	cache := mustCache(t, transport.InfoRequest)
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
		query    transport.QueryType
		response transport.ResponseType
		payload  []byte
	}{
		{name: "info", query: transport.InfoRequest, response: transport.ResponseInfo, payload: []byte("info")},
		{name: "players", query: transport.PlayerRequest, response: transport.ResponsePlayers, payload: []byte("players")},
		{name: "rules", query: transport.RulesRequest, response: transport.ResponseRules, payload: []byte("opaque A3SB rules")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cache := mustCache(t, test.query)
			if err := cache.Store(test.query, transport.Packet{
				Type:    test.response,
				Payload: test.payload,
			}); err != nil {
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
	cache := mustCache(t, transport.RulesRequest)
	relay := &handlerUpstream{}
	handler := mustHandler(t, cache, relay, false)

	response, err := handler.Handle(context.Background(), requestFor(transport.RulesRequest))
	if !errors.Is(err, server.ErrDrop) || response != nil {
		t.Fatalf("Handle() = %#v, %v; want dropped response", response, err)
	}
	if got := relay.Count(transport.RulesRequest); got != 0 {
		t.Fatalf("relay calls = %d, want 0", got)
	}
}

func TestHandlerRelaysUncachedQueries(t *testing.T) {
	queries := []struct {
		query    transport.QueryType
		response transport.ResponseType
	}{
		{query: transport.InfoRequest, response: transport.ResponseInfo},
		{query: transport.PlayerRequest, response: transport.ResponsePlayers},
		{query: transport.RulesRequest, response: transport.ResponseRules},
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

	response, err := handler.Handle(context.Background(), requestFor(transport.InfoRequest))
	if !errors.Is(err, server.ErrDrop) || response != nil {
		t.Fatalf("Handle() = %#v, %v; want dropped response", response, err)
	}
}

func TestHandlerAnswersGetChallengeLocally(t *testing.T) {
	provider := &handlerChallengeProvider{challenge: transport.Challenge{1, 2, 3, 4}}
	relay := &handlerUpstream{}
	handler, err := NewHandler(mustCache(t), relay, HandlerConfig{ChallengeProvider: provider})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}

	request := requestFor(transport.ChallengeRequest)
	response, err := handler.Handle(context.Background(), request)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	packet := responsePacket(t, response)
	if packet.Type != transport.ResponseChallenge || string(packet.Payload) != string(provider.challenge[:]) {
		t.Fatalf("challenge packet = %#v, want %x", packet, provider.challenge)
	}
	if provider.remote != request.Remote {
		t.Fatalf("provider remote = %s, want %s", provider.remote, request.Remote)
	}
	if got := relay.Count(transport.ChallengeRequest); got != 0 {
		t.Fatalf("relay calls = %d, want 0", got)
	}
}

func TestHandlerRelaysPingByDefault(t *testing.T) {
	relay := &handlerUpstream{}
	handler := mustHandler(t, mustCache(t), relay, false)

	response, err := handler.Handle(context.Background(), requestFor(transport.PingRequest))
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if packet := responsePacket(t, response); packet.Type != transport.ResponsePing {
		t.Fatalf("PING response type = %x, want %x", packet.Type, transport.ResponsePing)
	}
	if got := relay.Count(transport.PingRequest); got != 1 {
		t.Fatalf("relay calls = %d, want 1", got)
	}
}

func TestHandlerAnswersLocalPingWithoutRelay(t *testing.T) {
	relay := &handlerUpstream{}
	handler := mustHandler(t, mustCache(t), relay, true)

	response, err := handler.Handle(context.Background(), requestFor(transport.PingRequest))
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	packet := responsePacket(t, response)
	if packet.Type != transport.ResponsePing || len(packet.Payload) != 0 {
		t.Fatalf("local PING packet = %#v, want empty ACK", packet)
	}
	if got := relay.Count(transport.PingRequest); got != 0 {
		t.Fatalf("relay calls = %d, want 0", got)
	}
}

func TestHandlerDropsUnknownRequest(t *testing.T) {
	handler := mustHandler(t, mustCache(t), &handlerUpstream{}, false)
	request := requestFor(transport.QueryType(0xFE))

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

func requestFor(query transport.QueryType) *server.Request {
	return &server.Request{
		Remote: netip.MustParseAddrPort("127.0.0.1:27015"),
		Query:  transport.Request{Type: query},
	}
}

func responsePacket(t *testing.T, response server.Response) transport.Packet {
	t.Helper()
	packetResponse, ok := response.(server.PacketResponse)
	if !ok {
		t.Fatalf("response = %T, want server.PacketResponse", response)
	}
	return packetResponse.Packet
}

type handlerUpstream struct {
	mu    sync.Mutex
	calls map[transport.QueryType]int
	err   error
}

func (u *handlerUpstream) Query(_ context.Context, query transport.QueryType) (transport.Packet, transport.QueryMeta, error) {
	u.mu.Lock()
	if u.calls == nil {
		u.calls = make(map[transport.QueryType]int)
	}
	u.calls[query]++
	u.mu.Unlock()

	if u.err != nil {
		return transport.Packet{}, transport.QueryMeta{}, u.err
	}

	return responsePacketFor(query), transport.QueryMeta{}, nil
}

func (u *handlerUpstream) Count(query transport.QueryType) int {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.calls[query]
}

func responsePacketFor(query transport.QueryType) transport.Packet {
	switch query {
	case transport.InfoRequest:
		return transport.Packet{Type: transport.ResponseInfo, Payload: []byte("info")}
	case transport.PlayerRequest:
		return transport.Packet{Type: transport.ResponsePlayers, Payload: []byte("players")}
	case transport.RulesRequest:
		return transport.Packet{Type: transport.ResponseRules, Payload: []byte("rules")}
	case transport.PingRequest:
		return transport.Packet{Type: transport.ResponsePing, Payload: []byte("pong\x00")}
	default:
		return transport.Packet{}
	}
}

type handlerChallengeProvider struct {
	challenge transport.Challenge
	remote    netip.AddrPort
}

func (p *handlerChallengeProvider) Issue(remote netip.AddrPort) transport.Challenge {
	p.remote = remote
	return p.challenge
}

func (*handlerChallengeProvider) Validate(netip.AddrPort, transport.Challenge) bool {
	return true
}
