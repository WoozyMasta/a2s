// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package server

import (
	"errors"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestNormalizeResponseTyped(t *testing.T) {
	tests := []struct {
		name     string
		query    transport.QueryType
		response Response
		wantType transport.ResponseType
	}{
		{
			name:     "info",
			query:    transport.InfoRequest,
			response: InfoResponse{Info: a2s.Info{Format: a2s.InfoFormat(transport.ResponseInfo)}},
			wantType: transport.ResponseInfo,
		},
		{
			name:     "players",
			query:    transport.PlayerRequest,
			response: PlayersResponse{Players: []a2s.Player{{Name: "player"}}},
			wantType: transport.ResponsePlayers,
		},
		{
			name:     "rules",
			query:    transport.RulesRequest,
			response: RulesResponse{Rules: a2s.Rules{{Name: "hostname", Value: "server"}}},
			wantType: transport.ResponseRules,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			packet, err := NormalizeResponse(&Request{Query: transport.Request{Type: test.query}}, test.response)
			if err != nil {
				t.Fatalf("normalizeResponse() error = %v", err)
			}
			if packet.Type != test.wantType {
				t.Errorf("normalizeResponse() type = 0x%X, want 0x%X", packet.Type, test.wantType)
			}
		})
	}
}

func TestNormalizeResponsePacketIsLossless(t *testing.T) {
	payload := []byte("opaque response payload")
	want := transport.Packet{Type: transport.ResponseRules, Payload: payload}

	got, err := NormalizeResponse(
		&Request{Query: transport.Request{Type: transport.RulesRequest}},
		PacketResponse{Packet: want},
	)
	if err != nil {
		t.Fatalf("normalizeResponse() error = %v", err)
	}
	if got.Type != want.Type || &got.Payload[0] != &want.Payload[0] {
		t.Fatalf("normalizeResponse() changed packet: got %#v, want %#v", got, want)
	}
}

func TestNormalizeResponseRejectsIncompatibleResponse(t *testing.T) {
	_, err := NormalizeResponse(
		&Request{Query: transport.Request{Type: transport.InfoRequest}},
		PlayersResponse{},
	)
	if !errors.Is(err, ErrResponseQuery) {
		t.Fatalf("normalizeResponse() error = %v, want ErrResponseQuery", err)
	}
}

func TestNormalizeResponseWrapsEncoderError(t *testing.T) {
	_, err := NormalizeResponse(
		&Request{Query: transport.Request{Type: transport.RulesRequest}},
		RulesResponse{Rules: a2s.Rules{{Name: "invalid\x00name"}}},
	)
	if !errors.Is(err, ErrResponse) {
		t.Fatalf("normalizeResponse() error = %v, want ErrResponse", err)
	}
	if !errors.Is(err, a2s.ErrRuleEncode) {
		t.Fatalf("normalizeResponse() error = %v, want a2s.ErrRuleEncode", err)
	}
}

func TestNormalizeResponseAllowsProtocolPackets(t *testing.T) {
	tests := []struct {
		query        transport.QueryType
		responseType transport.ResponseType
	}{
		{query: transport.InfoRequest, responseType: transport.ResponseChallenge},
		{query: transport.PlayerRequest, responseType: transport.ResponseChallenge},
		{query: transport.RulesRequest, responseType: transport.ResponseChallenge},
		{query: transport.ChallengeRequest, responseType: transport.ResponseChallenge},
		{query: transport.PingRequest, responseType: transport.ResponsePing},
	}

	for _, test := range tests {
		_, err := NormalizeResponse(
			&Request{Query: transport.Request{Type: test.query}},
			PacketResponse{Packet: transport.Packet{Type: test.responseType}},
		)
		if err != nil {
			t.Errorf("normalizeResponse(%#v) error = %v", test, err)
		}
	}
}
