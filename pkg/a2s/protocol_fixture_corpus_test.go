package a2s

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/woozymasta/a2s/internal/testfixtures"
)

func TestProtocolFixtureCorpusRequests(t *testing.T) {
	tests := []struct {
		name      string
		fixture   string
		request   QueryType
		challenge Challenge
	}{
		{
			name:      "info",
			fixture:   "request_info.hex",
			request:   InfoRequest,
			challenge: InitialChallenge,
		},
		{
			name:      "challenge",
			fixture:   "request_challenge.hex",
			request:   ChallengeRequest,
			challenge: InitialChallenge,
		},
		{
			name:      "players",
			fixture:   "request_players_challenge.hex",
			request:   PlayerRequest,
			challenge: Challenge{0x78, 0x56, 0x34, 0x12},
		},
		{
			name:      "rules",
			fixture:   "request_rules_challenge.hex",
			request:   RulesRequest,
			challenge: Challenge{0x78, 0x56, 0x34, 0x12},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			want := readProtocolFixture(t, test.fixture)
			got, err := createHeader(test.request, test.challenge)
			if err != nil {
				t.Fatalf("createHeader() error = %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("request = %X, want %X", got, want)
			}
		})
	}
}

func TestProtocolFixtureCorpusA2SPayloads(t *testing.T) {
	t.Run("source info", func(t *testing.T) {
		info, err := DecodeInfo(Packet{
			Type:    ResponseInfo,
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
		info, err := DecodeInfo(Packet{
			Type:    ResponseInfoGoldSource,
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
		players, err := parsePlayers(readProtocolFixture(t, "players_payload.hex"))
		if err != nil {
			t.Fatalf("parsePlayers() error = %v", err)
		}
		if len(players) != 2 || players[0].Score != -1 || players[1].Score != int32(^uint32(0)>>1) {
			t.Fatalf("parsed players = %+v", players)
		}
	})

	t.Run("rules", func(t *testing.T) {
		rules, err := parseRules(readProtocolFixture(t, "rules_payload.hex"))
		if err != nil {
			t.Fatalf("parseRules() error = %v", err)
		}
		mode, modeOK := rules.Get("mode")
		encoded, encodedOK := rules.Get("encoded")
		if !modeOK || !encodedOK || mode != "coop" || encoded != "c2VydmVy" {
			t.Fatalf("parsed rules = %#v", rules)
		}
	})

	t.Run("challenge", func(t *testing.T) {
		packet := readProtocolFixture(t, "response_challenge.hex")
		challenge, err := parseChallenge(packet[5:])
		if err != nil {
			t.Fatalf("parseChallenge() error = %v", err)
		}
		want := Challenge{0x78, 0x56, 0x34, 0x12}
		if challenge != want {
			t.Fatalf("challenge = %X, want %X", challenge, want)
		}
	})
}

func TestProtocolFixtureCorpusSourceSplitHeaders(t *testing.T) {
	for index := 0; index < 5; index++ {
		packet := readProtocolFixture(t, "source_split_packet_"+strconv.Itoa(index)+".hex")
		header, err := parseSplitHeader(packet)
		if err != nil {
			t.Fatalf("parseSplitHeader(packet %d) error = %v", index, err)
		}
		if header.id != 0x12345678 || header.count != 5 || header.index != index {
			t.Fatalf("header %d = %+v", index, header)
		}
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
