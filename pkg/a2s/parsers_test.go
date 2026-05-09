package a2s

import (
	"encoding/binary"
	"errors"
	"math"
	"testing"
	"time"
)

func TestParseChallenge(t *testing.T) {
	data := []byte{0x78, 0x56, 0x34, 0x12}

	got, err := parseChallenge(data)
	if err != nil {
		t.Fatalf("parseChallenge returned error: %v", err)
	}
	if got != 0x12345678 {
		t.Fatalf("challenge = 0x%X, want 0x12345678", got)
	}

	if _, err := parseChallenge(data[:3]); !errors.Is(err, ErrChallengeValue) {
		t.Fatalf("truncated challenge error = %v, want ErrChallengeValue", err)
	}
}

func TestParseInfo(t *testing.T) {
	body := []byte{17}
	for _, value := range []string{"Test server", "test_map", "test_folder", "Test game"} {
		body = append(body, value...)
		body = append(body, 0)
	}
	body = binary.LittleEndian.AppendUint16(body, 1234)
	body = append(body, 1, 16, 2, 'd', 'w', 1, 0)
	body = append(body, "1.0"...)
	body = append(body, 0, 0)

	info, err := parseInfo(body, infoResponseSource, 25*time.Millisecond)
	if err != nil {
		t.Fatalf("parseInfo returned error: %v", err)
	}
	if info.Name != "Test server" || info.ID != 1234 || info.Ping != 25*time.Millisecond {
		t.Fatalf("parsed info = %+v", info)
	}
}

func TestParsePlayers(t *testing.T) {
	data := []byte{2}
	for index, score := range []int32{-1, math.MaxInt32} {
		data = append(data, byte(index))
		data = append(data, "player"...)
		data = append(data, byte('0'+index), 0)
		data = binary.LittleEndian.AppendUint32(data, uint32(score))
		data = binary.LittleEndian.AppendUint32(data, math.Float32bits(1))
	}

	players, err := parsePlayers(data)
	if err != nil {
		t.Fatalf("parsePlayers returned error: %v", err)
	}
	if len(players) != 2 || players[0].Score != -1 || players[1].Score != math.MaxInt32 {
		t.Fatalf("parsed players = %+v", players)
	}
}

func TestParseTheShipPlayers(t *testing.T) {
	data := []byte{1, 0}
	data = append(data, "player"...)
	data = append(data, 0)
	score := int32(-1)
	data = binary.LittleEndian.AppendUint32(data, uint32(score))
	data = binary.LittleEndian.AppendUint32(data, math.Float32bits(1))
	data = binary.LittleEndian.AppendUint32(data, 2)
	data = binary.LittleEndian.AppendUint32(data, 3)

	players, err := parseTheShipPlayers(data)
	if err != nil {
		t.Fatalf("parseTheShipPlayers returned error: %v", err)
	}
	if len(players) != 1 || players[0].Score != -1 || players[0].Deaths != 2 || players[0].Money != 3 {
		t.Fatalf("parsed The Ship players = %+v", players)
	}
}

func TestParseRules(t *testing.T) {
	data := []byte{2, 0}
	for _, entry := range [][2]string{{"mode", "coop"}, {"encoded", "c2VydmVy"}} {
		data = append(data, entry[0]...)
		data = append(data, 0)
		data = append(data, entry[1]...)
		data = append(data, 0)
	}

	rules, err := parseRules(data)
	if err != nil {
		t.Fatalf("parseRules returned error: %v", err)
	}
	if rules["mode"] != "coop" || rules["encoded"] != "c2VydmVy" {
		t.Fatalf("parsed rules = %#v", rules)
	}
}

func BenchmarkParsePlayers(b *testing.B) {
	data := []byte{1, 0}
	data = append(data, "player"...)
	data = append(data, 0)
	data = binary.LittleEndian.AppendUint32(data, 1)
	data = binary.LittleEndian.AppendUint32(data, math.Float32bits(1))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := parsePlayers(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseRules(b *testing.B) {
	data := []byte{1, 0, 'm', 'o', 'd', 'e', 0, 'c', 'o', 'o', 'p', 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := parseRules(data); err != nil {
			b.Fatal(err)
		}
	}
}
