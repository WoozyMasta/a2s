package a2s

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestAppendPlayersMatchesFixture(t *testing.T) {
	payload := readProtocolFixture(t, "players_payload.hex")
	players, err := DecodePlayers(Packet{Type: ResponsePlayers, Payload: payload})
	if err != nil {
		t.Fatalf("DecodePlayers() error = %v", err)
	}

	encoded, err := AppendPlayers(nil, players)
	if err != nil {
		t.Fatalf("AppendPlayers() error = %v", err)
	}

	want := singlePacketFixture(ResponsePlayers, payload)
	if !bytes.Equal(encoded, want) {
		t.Fatalf("encoded players response = %X, want %X", encoded, want)
	}
}

func TestAppendPlayersRoundTrip(t *testing.T) {
	want := []Player{
		{Name: "negative", Index: 2, Score: -123, Duration: 1250 * time.Millisecond},
		{Name: "fractional", Index: 7, Score: 456, Duration: 1*time.Second + 250*time.Millisecond},
	}

	encoded, err := AppendPlayers(nil, want)
	if err != nil {
		t.Fatalf("AppendPlayers() error = %v", err)
	}
	packet, err := DecodePacket(encoded)
	if err != nil {
		t.Fatalf("DecodePacket() error = %v", err)
	}
	got, err := DecodePlayers(packet)
	if err != nil {
		t.Fatalf("DecodePlayers() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip players = %+v, want %+v", got, want)
	}
}

func TestAppendPlayersRejectsInvalidModels(t *testing.T) {
	tests := []struct {
		name    string
		players []Player
	}{
		{name: "too many players", players: make([]Player, 256)},
		{name: "embedded NUL", players: []Player{{Name: "bad\x00name"}}},
		{name: "negative duration", players: []Player{{Duration: -time.Nanosecond}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := AppendPlayers(nil, test.players); !errors.Is(err, ErrPlayerEncode) {
				t.Fatalf("AppendPlayers() error = %v, want ErrPlayerEncode", err)
			}
		})
	}
}

func TestDecodePlayersRejectsWrongResponseType(t *testing.T) {
	_, err := DecodePlayers(Packet{Type: ResponseRules})
	if !errors.Is(err, ErrPlayerRead) {
		t.Fatalf("DecodePlayers() error = %v, want ErrPlayerRead", err)
	}
}
