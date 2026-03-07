package a2s

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/woozymasta/a2s/internal/bread"
)

func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hex decode failed: %v", err)
	}
	return b
}

func TestReadGoldSourceInfo_MissingBotsByte(t *testing.T) {
	// Real-world response sample where VAC exists but bots byte is missing.
	data := mustDecodeHex(t, "302e302e302e303a3237303135004e4f5244204c4c472023204e45572049503a203133352e3132352e3231322e32393a32373031350064655f64757374320d0063737472696b65005061696e7462616c6c204d6f64001f202f646c000001")

	info := &Info{}
	if err := info.readGoldSourceInfo(bread.NewReader(data)); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if info.VAC != true {
		t.Fatalf("expected VAC=true, got %v", info.VAC)
	}
	if info.Bots != 0 {
		t.Fatalf("expected missing bots to default to 0, got %d", info.Bots)
	}
}

func TestReadGoldSourceInfo_MissingVacAndBotsBytes(t *testing.T) {
	data := []byte{
		'1', '.', '2', '.', '3', '.', '4', ':', '2', '7', '0', '1', '5', 0,
		'N', 0,
		'm', 0,
		'f', 0,
		'g', 0,
		1,    // players
		2,    // max players
		47,   // protocol
		'd',  // server type
		'l',  // environment
		0x00, // visibility
		0x00, // modded
		// VAC and bots are intentionally absent
	}

	info := &Info{}
	if err := info.readGoldSourceInfo(bread.NewReader(data)); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if info.VAC {
		t.Fatalf("expected missing VAC to default to false, got %v", info.VAC)
	}
	if info.Bots != 0 {
		t.Fatalf("expected missing bots to default to 0, got %d", info.Bots)
	}
}

func TestReadGoldSourceInfo_StillFailsOnMandatoryFields(t *testing.T) {
	data := []byte{
		'1', '.', '2', '.', '3', '.', '4', ':', '2', '7', '0', '1', '5', 0,
		'N', 0,
		'm', 0,
		'f', 0,
		'g', 0,
		1, // players
		// max players is missing
	}

	info := &Info{}
	err := info.readGoldSourceInfo(bread.NewReader(data))
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if !errors.Is(err, ErrInfoMaxPlayerCount) {
		t.Fatalf("expected ErrInfoMaxPlayerCount, got %v", err)
	}
}
