// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"errors"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestRequestCodecRoundTrip(t *testing.T) {
	challenge := transport.Challenge{0x78, 0x56, 0x34, 0x12}
	tests := []struct {
		name    string
		request transport.Request
	}{
		{name: "info", request: transport.Request{Type: transport.InfoRequest}},
		{name: "info with challenge", request: transport.Request{
			Type:         transport.InfoRequest,
			Challenge:    challenge,
			HasChallenge: true,
		}},
		{name: "players", request: transport.Request{
			Type:         transport.PlayerRequest,
			Challenge:    challenge,
			HasChallenge: true,
		}},
		{name: "rules", request: transport.Request{
			Type:         transport.RulesRequest,
			Challenge:    challenge,
			HasChallenge: true,
		}},
		{name: "challenge", request: transport.Request{Type: transport.ChallengeRequest}},
		{name: "ping", request: transport.Request{Type: transport.PingRequest}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := transport.AppendRequest(nil, test.request)
			if err != nil {
				t.Fatalf("transport.AppendRequest() error = %v", err)
			}

			decoded, err := transport.DecodeRequest(encoded)
			if err != nil {
				t.Fatalf("transport.DecodeRequest() error = %v", err)
			}
			if decoded != test.request {
				t.Fatalf("decoded request = %+v, want %+v", decoded, test.request)
			}
		})
	}
}

func TestDecodeRequestFixtures(t *testing.T) {
	challenge := transport.Challenge{0x78, 0x56, 0x34, 0x12}
	tests := []struct {
		name string
		data []byte
		want transport.Request
	}{
		{
			name: "info",
			data: readProtocolFixture(t, "request_info.hex"),
			want: transport.Request{Type: transport.InfoRequest},
		},
		{
			name: "challenge",
			data: readProtocolFixture(t, "request_challenge.hex"),
			want: transport.Request{Type: transport.ChallengeRequest},
		},
		{
			name: "players",
			data: readProtocolFixture(t, "request_players_challenge.hex"),
			want: transport.Request{Type: transport.PlayerRequest, Challenge: challenge, HasChallenge: true},
		},
		{
			name: "rules",
			data: readProtocolFixture(t, "request_rules_challenge.hex"),
			want: transport.Request{Type: transport.RulesRequest, Challenge: challenge, HasChallenge: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := transport.DecodeRequest(test.data)
			if err != nil {
				t.Fatalf("transport.DecodeRequest() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("decoded request = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestDecodeRequestRejectsMalformedDatagrams(t *testing.T) {
	validInfo := readProtocolFixture(t, "request_info.hex")
	validPlayers := readProtocolFixture(t, "request_players_challenge.hex")
	validPing, err := transport.AppendRequest(nil, transport.Request{Type: transport.PingRequest})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "short header", data: validInfo[:4], want: transport.ErrRequestHeader},
		{name: "wrong marker", data: append([]byte{0, 0, 0, 0}, validInfo[4:]...), want: transport.ErrRequestHeader},
		{name: "truncated info", data: validInfo[:len(validInfo)-1], want: transport.ErrRequestPayload},
		{name: "extra info payload", data: append(append([]byte(nil), validInfo...), 0), want: transport.ErrRequestPayload},
		{name: "truncated players", data: validPlayers[:len(validPlayers)-1], want: transport.ErrRequestPayload},
		{name: "extra ping payload", data: append(append([]byte(nil), validPing...), 0), want: transport.ErrRequestPayload},
		{name: "unknown type", data: append(append([]byte(nil), validPing[:4]...), 0xFF), want: transport.ErrWrongRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := transport.DecodeRequest(test.data)
			if !errors.Is(err, test.want) {
				t.Fatalf("transport.DecodeRequest() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestAppendRequestRejectsInvalidModels(t *testing.T) {
	tests := []struct {
		name    string
		request transport.Request
		want    error
	}{
		{
			name:    "players without challenge",
			request: transport.Request{Type: transport.PlayerRequest},
			want:    transport.ErrRequestPayload,
		},
		{
			name:    "rules without challenge",
			request: transport.Request{Type: transport.RulesRequest},
			want:    transport.ErrRequestPayload,
		},
		{
			name:    "ping with challenge",
			request: transport.Request{Type: transport.PingRequest, HasChallenge: true},
			want:    transport.ErrRequestPayload,
		},
		{
			name:    "unknown request",
			request: transport.Request{Type: 0xFF},
			want:    transport.ErrWrongRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := transport.AppendRequest(nil, test.request); !errors.Is(err, test.want) {
				t.Fatalf("transport.AppendRequest() error = %v, want %v", err, test.want)
			}
		})
	}
}
