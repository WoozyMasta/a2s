// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/woozymasta/a2s/internal/testfixtures"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

func TestProtocolFixtureCorpusRequests(t *testing.T) {
	tests := []struct {
		name      string
		fixture   string
		request   transport.QueryType
		challenge transport.Challenge
	}{
		{
			name:      "info",
			fixture:   "request_info.hex",
			request:   transport.InfoRequest,
			challenge: transport.InitialChallenge(),
		},
		{
			name:      "challenge",
			fixture:   "request_challenge.hex",
			request:   transport.ChallengeRequest,
			challenge: transport.InitialChallenge(),
		},
		{
			name:      "players",
			fixture:   "request_players_challenge.hex",
			request:   transport.PlayerRequest,
			challenge: transport.Challenge{0x78, 0x56, 0x34, 0x12},
		},
		{
			name:      "rules",
			fixture:   "request_rules_challenge.hex",
			request:   transport.RulesRequest,
			challenge: transport.Challenge{0x78, 0x56, 0x34, 0x12},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			want := readProtocolFixture(t, test.fixture)
			got, err := transport.CreateHeader(test.request, test.challenge)
			if err != nil {
				t.Fatalf("transport.CreateHeader() error = %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("request = %X, want %X", got, want)
			}
		})
	}
}

func TestProtocolFixtureCorpusA2SPayloads(t *testing.T) {
	t.Run("source info", func(t *testing.T) {
		info, err := DecodeInfo(transport.Packet{
			Type:    transport.ResponseInfo,
			Payload: readProtocolFixture(t, "source_info_payload.hex"),
		})
		if err != nil {
			t.Fatalf("DecodeInfo() error = %v", err)
		}
		if info.Name != "Test server" || info.Map != "test_map" {
			t.Fatalf("parsed info = %+v", info)
		}
	})

	t.Run("goldsource info", func(t *testing.T) {
		info, err := DecodeInfo(transport.Packet{
			Type:    transport.ResponseInfoGoldSource,
			Payload: readProtocolFixture(t, "goldsource_info_payload.hex"),
		})
		if err != nil {
			t.Fatalf("DecodeInfo() error = %v", err)
		}
		if info.Name != "Test server" || info.Map != "test_map" {
			t.Fatalf("parsed info = %+v", info)
		}
	})

	t.Run("players", func(t *testing.T) {
		players, err := DecodePlayers(transport.Packet{
			Type:    transport.ResponsePlayers,
			Payload: readProtocolFixture(t, "players_payload.hex"),
		})
		if err != nil {
			t.Fatalf("DecodePlayers() error = %v", err)
		}
		if len(players) != 2 || players[0].Score != -1 || players[1].Score != int32(^uint32(0)>>1) {
			t.Fatalf("parsed players = %+v", players)
		}
	})

	t.Run("rules", func(t *testing.T) {
		rules, err := DecodeRules(transport.Packet{
			Type:    transport.ResponseRules,
			Payload: readProtocolFixture(t, "rules_payload.hex"),
		})
		if err != nil {
			t.Fatalf("DecodeRules() error = %v", err)
		}
		mode, modeOK := rules.Get("mode")
		encoded, encodedOK := rules.Get("encoded")
		if !modeOK || !encodedOK || mode != "coop" || encoded != "c2VydmVy" {
			t.Fatalf("parsed rules = %#v", rules)
		}
	})

	t.Run("challenge", func(t *testing.T) {
		packet := readProtocolFixture(t, "response_challenge.hex")
		challenge, err := transport.ParseChallengePayload(packet[5:])
		if err != nil {
			t.Fatalf("ParseChallengePayload() error = %v", err)
		}
		want := transport.Challenge{0x78, 0x56, 0x34, 0x12}
		if challenge != want {
			t.Fatalf("challenge = %X, want %X", challenge, want)
		}
	})
}

func TestProtocolFixtureCorpusSourceSplitHeaders(t *testing.T) {
	for index := 0; index < 5; index++ {
		packet := readProtocolFixture(t, "source_split_packet_"+strconv.Itoa(index)+".hex")
		header, err := transport.ParseSplitHeader(packet)
		if err != nil {
			t.Fatalf("transport.ParseSplitHeader(packet %d) error = %v", index, err)
		}
		if header.ID != 0x12345678 || header.Count != 5 || header.Index != index {
			t.Fatalf("header %d = %+v", index, header)
		}
	}
}

func TestProtocolFixtureCorpusGoldSourceSplitHeader(t *testing.T) {
	// The fixture header is copied from the real datagram published in
	// https://github.com/ValveSoftware/halflife/issues/2406.
	// Its 0x12 control byte is documented there as fragment 2 of 2.
	packet := readProtocolFixture(t, "goldsource_split_packet_1.hex")
	header, err := transport.ParseSplitHeader(packet)
	if err != nil {
		t.Fatalf("transport.ParseSplitHeader() error = %v", err)
	}

	if !header.GoldSource {
		t.Fatal("GoldSource fixture classified as Source")
	}
	if header.Compressed {
		t.Fatal("GoldSource fixture incorrectly classified as compressed")
	}
	if header.ID != 0xD9D51BBC || header.Count != 2 || header.Index != 1 {
		t.Fatalf("header = %+v, want ID 0xD9D51BBC, count 2, index 1", header)
	}
}

func readProtocolFixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := testfixtures.Read(name)
	if err != nil {
		t.Fatal(err)
	}

	return data
}
