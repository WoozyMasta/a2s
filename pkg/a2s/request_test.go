package a2s

import (
	"errors"
	"testing"
)

func TestRequestCodecRoundTrip(t *testing.T) {
	challenge := Challenge{0x78, 0x56, 0x34, 0x12}
	tests := []struct {
		name    string
		request Request
	}{
		{name: "info", request: Request{Type: InfoRequest}},
		{name: "info with challenge", request: Request{
			Type:         InfoRequest,
			Challenge:    challenge,
			HasChallenge: true,
		}},
		{name: "players", request: Request{
			Type:         PlayerRequest,
			Challenge:    challenge,
			HasChallenge: true,
		}},
		{name: "rules", request: Request{
			Type:         RulesRequest,
			Challenge:    challenge,
			HasChallenge: true,
		}},
		{name: "challenge", request: Request{Type: ChallengeRequest}},
		{name: "ping", request: Request{Type: PingRequest}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := AppendRequest(nil, test.request)
			if err != nil {
				t.Fatalf("AppendRequest() error = %v", err)
			}

			decoded, err := DecodeRequest(encoded)
			if err != nil {
				t.Fatalf("DecodeRequest() error = %v", err)
			}
			if decoded != test.request {
				t.Fatalf("decoded request = %+v, want %+v", decoded, test.request)
			}
		})
	}
}

func TestDecodeRequestFixtures(t *testing.T) {
	challenge := Challenge{0x78, 0x56, 0x34, 0x12}
	tests := []struct {
		name string
		data []byte
		want Request
	}{
		{
			name: "info",
			data: readProtocolFixture(t, "request_info.hex"),
			want: Request{Type: InfoRequest},
		},
		{
			name: "challenge",
			data: readProtocolFixture(t, "request_challenge.hex"),
			want: Request{Type: ChallengeRequest},
		},
		{
			name: "players",
			data: readProtocolFixture(t, "request_players_challenge.hex"),
			want: Request{Type: PlayerRequest, Challenge: challenge, HasChallenge: true},
		},
		{
			name: "rules",
			data: readProtocolFixture(t, "request_rules_challenge.hex"),
			want: Request{Type: RulesRequest, Challenge: challenge, HasChallenge: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := DecodeRequest(test.data)
			if err != nil {
				t.Fatalf("DecodeRequest() error = %v", err)
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
	validPing, err := AppendRequest(nil, Request{Type: PingRequest})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "short header", data: validInfo[:4], want: ErrRequestHeader},
		{name: "wrong marker", data: append([]byte{0, 0, 0, 0}, validInfo[4:]...), want: ErrRequestHeader},
		{name: "truncated info", data: validInfo[:len(validInfo)-1], want: ErrRequestPayload},
		{name: "extra info payload", data: append(append([]byte(nil), validInfo...), 0), want: ErrRequestPayload},
		{name: "truncated players", data: validPlayers[:len(validPlayers)-1], want: ErrRequestPayload},
		{name: "extra ping payload", data: append(append([]byte(nil), validPing...), 0), want: ErrRequestPayload},
		{name: "unknown type", data: append(append([]byte(nil), validPing[:4]...), 0xFF), want: ErrWrongRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := DecodeRequest(test.data)
			if !errors.Is(err, test.want) {
				t.Fatalf("DecodeRequest() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestAppendRequestRejectsInvalidModels(t *testing.T) {
	tests := []struct {
		name    string
		request Request
		want    error
	}{
		{
			name:    "players without challenge",
			request: Request{Type: PlayerRequest},
			want:    ErrRequestPayload,
		},
		{
			name:    "rules without challenge",
			request: Request{Type: RulesRequest},
			want:    ErrRequestPayload,
		},
		{
			name:    "ping with challenge",
			request: Request{Type: PingRequest, HasChallenge: true},
			want:    ErrRequestPayload,
		},
		{
			name:    "unknown request",
			request: Request{Type: 0xFF},
			want:    ErrWrongRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := AppendRequest(nil, test.request); !errors.Is(err, test.want) {
				t.Fatalf("AppendRequest() error = %v, want %v", err, test.want)
			}
		})
	}
}
