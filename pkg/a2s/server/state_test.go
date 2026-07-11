package server

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func TestStateStoresDeepCopiedPreEncodedSnapshot(t *testing.T) {
	gameID := uint64(123456789)
	info := &a2s.Info{
		Format:   a2s.InfoFormat(a2s.ResponseInfo),
		Protocol: 17,
		Name:     "before",
		Map:      "map_before",
		Folder:   "game",
		Game:     "Game",
		Version:  "1.0",
		AppID:    10,
		Keywords: []string{"tag_before"},
		GameID:   &gameID,
	}
	players := []a2s.Player{{Name: "player_before", Score: 7}}
	rules := a2s.Rules{{Name: "rule", Value: "value_before"}}

	state := NewState()
	if err := state.Store(Snapshot{Info: info, Players: players, Rules: rules}); err != nil {
		t.Fatalf("Store() error = %v", err)
	}

	info.Name = "after"
	info.Keywords[0] = "tag_after"
	gameID = 987654321
	players[0].Name = "player_after"
	rules[0].Value = "value_after"

	infoResponse := stateResponse(t, state, a2s.InfoRequest)
	decodedInfo, err := a2s.DecodeInfo(infoResponse.Packet)
	if err != nil {
		t.Fatalf("DecodeInfo() error = %v", err)
	}
	if decodedInfo.Name != "before" || decodedInfo.Keywords[0] != "tag_before" || decodedInfo.EffectiveID() != 123456789 {
		t.Fatalf("stored info changed after Store: %#v", decodedInfo)
	}

	playersResponse := stateResponse(t, state, a2s.PlayerRequest)
	decodedPlayers, err := a2s.DecodePlayers(playersResponse.Packet)
	if err != nil {
		t.Fatalf("DecodePlayers() error = %v", err)
	}
	if len(decodedPlayers) != 1 || decodedPlayers[0].Name != "player_before" {
		t.Fatalf("stored players changed after Store: %#v", decodedPlayers)
	}

	rulesResponse := stateResponse(t, state, a2s.RulesRequest)
	decodedRules, err := a2s.DecodeRules(rulesResponse.Packet)
	if err != nil {
		t.Fatalf("DecodeRules() error = %v", err)
	}
	if len(decodedRules) != 1 || decodedRules[0].Value != "value_before" {
		t.Fatalf("stored rules changed after Store: %#v", decodedRules)
	}

	infoResponse.Packet.Payload[0] ^= 0xFF
	decodedAgain := stateResponse(t, state, a2s.InfoRequest)
	if _, err := a2s.DecodeInfo(decodedAgain.Packet); err != nil {
		t.Fatalf("internal packet was modified through response: %v", err)
	}
}

func TestStateDefinesEmptyAndMissingResponses(t *testing.T) {
	state := NewState()
	request := &Request{Query: a2s.Request{Type: a2s.InfoRequest}}
	if _, err := state.Handle(context.Background(), request); !errors.Is(err, ErrStateUnavailable) {
		t.Fatalf("Handle(empty INFO) error = %v, want ErrStateUnavailable", err)
	}

	if err := state.Store(Snapshot{}); err != nil {
		t.Fatalf("Store(empty) error = %v", err)
	}
	if _, err := state.Handle(context.Background(), request); !errors.Is(err, ErrStateUnavailable) {
		t.Fatalf("Handle(missing INFO) error = %v, want ErrStateUnavailable", err)
	}

	players := stateResponse(t, state, a2s.PlayerRequest)
	decodedPlayers, err := a2s.DecodePlayers(players.Packet)
	if err != nil {
		t.Fatalf("DecodePlayers(empty) error = %v", err)
	}
	if len(decodedPlayers) != 0 {
		t.Fatalf("empty players = %#v, want empty", decodedPlayers)
	}

	rules := stateResponse(t, state, a2s.RulesRequest)
	decodedRules, err := a2s.DecodeRules(rules.Packet)
	if err != nil {
		t.Fatalf("DecodeRules(empty) error = %v", err)
	}
	if len(decodedRules) != 0 {
		t.Fatalf("empty rules = %#v, want empty", decodedRules)
	}
}

func TestStateStoreFailureKeepsPreviousSnapshot(t *testing.T) {
	state := NewState()
	if err := state.Store(Snapshot{Rules: a2s.Rules{{Name: "key", Value: "before"}}}); err != nil {
		t.Fatalf("Store(initial) error = %v", err)
	}

	invalid := &a2s.Info{
		Format:  a2s.InfoFormat(a2s.ResponseInfo),
		Name:    "invalid\x00name",
		Version: "1.0",
	}
	if err := state.Store(Snapshot{Info: invalid}); !errors.Is(err, ErrState) {
		t.Fatalf("Store(invalid) error = %v, want ErrState", err)
	}

	rules := stateResponse(t, state, a2s.RulesRequest)
	decodedRules, err := a2s.DecodeRules(rules.Packet)
	if err != nil {
		t.Fatalf("DecodeRules() error = %v", err)
	}
	if len(decodedRules) != 1 || decodedRules[0].Value != "before" {
		t.Fatalf("previous snapshot was replaced: %#v", decodedRules)
	}
}

func TestStateConcurrentStoreAndHandle(t *testing.T) {
	state := NewState()
	snapshots := []Snapshot{
		{Rules: a2s.Rules{{Name: "key", Value: "one"}}},
		{Rules: a2s.Rules{{Name: "key", Value: "two"}}},
	}
	if err := state.Store(snapshots[0]); err != nil {
		t.Fatalf("Store(initial) error = %v", err)
	}

	var group sync.WaitGroup
	group.Add(3)
	go func() {
		defer group.Done()
		for i := 0; i < 100; i++ {
			if err := state.Store(snapshots[i%len(snapshots)]); err != nil {
				t.Errorf("Store() error = %v", err)
				return
			}
		}
	}()
	for range 2 {
		go func() {
			defer group.Done()
			request := &Request{Query: a2s.Request{Type: a2s.RulesRequest}}
			for i := 0; i < 100; i++ {
				response, err := state.Handle(context.Background(), request)
				if err != nil {
					t.Errorf("Handle() error = %v", err)
					return
				}
				if _, ok := response.(PacketResponse); !ok {
					t.Errorf("Handle() response = %T, want PacketResponse", response)
					return
				}
			}
		}()
	}
	group.Wait()
}

func stateResponse(t *testing.T, state *State, query a2s.QueryType) PacketResponse {
	t.Helper()

	response, err := state.Handle(context.Background(), &Request{Query: a2s.Request{Type: query}})
	if err != nil {
		t.Fatalf("Handle(0x%X) error = %v", query, err)
	}

	packet, ok := response.(PacketResponse)
	if !ok {
		t.Fatalf("Handle(0x%X) response = %T, want PacketResponse", query, response)
	}

	return packet
}
