package a3sb

import (
	"testing"

	"github.com/woozymasta/a2s/internal/testfixtures"
	"github.com/woozymasta/a2s/pkg/appid"
)

func TestProtocolFixtureCorpusA3SBPayloads(t *testing.T) {
	tests := []struct {
		name     string
		fixture  string
		game     uint64
		modCount int
	}{
		{name: "arma3", fixture: "a3sb_arma3_payload.hex", game: appid.Arma3, modCount: 2},
		{name: "dayz", fixture: "a3sb_dayz_payload.hex", game: appid.DayZ, modCount: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := testfixtures.Read(test.fixture)
			if err != nil {
				t.Fatal(err)
			}

			rules := &Rules{id: test.game}
			if err := rules.readA3SB(data); err != nil {
				t.Fatalf("readA3SB() error = %v", err)
			}
			if rules.Version == 0 || len(rules.Mods) != test.modCount {
				t.Fatalf("parsed rules = %+v", rules)
			}
		})
	}
}
