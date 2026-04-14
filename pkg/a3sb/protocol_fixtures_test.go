package a3sb

import (
	"testing"

	"github.com/woozymasta/steam/utils/appid"
)

func TestMinimalDayZProtocolFixture(t *testing.T) {
	// v2, zero flags, no DLC, no mods, and no signatures.
	data := []byte{2, 0, 0, 0, 0, 0}
	rules := &Rules{id: appid.DayZ.Uint64()}

	if err := rules.readA3SB(data); err != nil {
		t.Fatalf("readA3SB returned error: %v", err)
	}
	if rules.Version != 2 {
		t.Fatalf("protocol version = %d, want 2", rules.Version)
	}
}
