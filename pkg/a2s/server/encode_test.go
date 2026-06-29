package server

import (
	"errors"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func TestNormalizeResponseTyped(t *testing.T) {
	tests := []struct {
		name     string
		query    a2s.QueryType
		response Response
		wantType a2s.ResponseType
	}{
		{
			name:     "info",
			query:    a2s.InfoRequest,
			response: InfoResponse{Info: a2s.Info{Format: a2s.InfoFormat(a2s.ResponseInfo)}},
			wantType: a2s.ResponseInfo,
		},
		{
			name:     "players",
			query:    a2s.PlayerRequest,
			response: PlayersResponse{Players: []a2s.Player{{Name: "player"}}},
			wantType: a2s.ResponsePlayers,
		},
		{
			name:     "rules",
			query:    a2s.RulesRequest,
			response: RulesResponse{Rules: a2s.Rules{{Name: "hostname", Value: "server"}}},
			wantType: a2s.ResponseRules,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			packet, err := NormalizeResponse(&Request{Query: a2s.Request{Type: test.query}}, test.response)
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
	want := a2s.Packet{Type: a2s.ResponseRules, Payload: payload}

	got, err := NormalizeResponse(
		&Request{Query: a2s.Request{Type: a2s.RulesRequest}},
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
		&Request{Query: a2s.Request{Type: a2s.InfoRequest}},
		PlayersResponse{},
	)
	if !errors.Is(err, ErrResponseQuery) {
		t.Fatalf("normalizeResponse() error = %v, want ErrResponseQuery", err)
	}
}

func TestNormalizeResponseWrapsEncoderError(t *testing.T) {
	_, err := NormalizeResponse(
		&Request{Query: a2s.Request{Type: a2s.RulesRequest}},
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
		query        a2s.QueryType
		responseType a2s.ResponseType
	}{
		{query: a2s.ChallengeRequest, responseType: a2s.ResponseChallenge},
		{query: a2s.PingRequest, responseType: a2s.ResponsePing},
	}

	for _, test := range tests {
		_, err := NormalizeResponse(
			&Request{Query: a2s.Request{Type: test.query}},
			PacketResponse{Packet: a2s.Packet{Type: test.responseType}},
		)
		if err != nil {
			t.Errorf("normalizeResponse(%#v) error = %v", test, err)
		}
	}
}
