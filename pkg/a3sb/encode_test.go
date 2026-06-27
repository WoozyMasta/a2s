package a3sb

import (
	"bytes"
	"errors"
	"testing"

	"github.com/woozymasta/a2s/internal/testfixtures"
	"github.com/woozymasta/a2s/pkg/appid"
)

func TestAppendBinaryMatchesFixtures(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		layout  Layout
		appID   uint64
	}{
		{name: "Arma3", fixture: "a3sb_arma3_payload.hex", layout: LayoutArma3, appID: appid.Arma3},
		{name: "DayZ", fixture: "a3sb_dayz_payload.hex", layout: LayoutDayZ, appID: appid.DayZ},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			want, err := testfixtures.Read(test.fixture)
			if err != nil {
				t.Fatal(err)
			}

			rules := &Rules{Layout: test.layout, appID: test.appID}
			if err := rules.readA3SB(want); err != nil {
				t.Fatalf("readA3SB() error = %v", err)
			}

			got, err := AppendBinary(nil, *rules)
			if err != nil {
				t.Fatalf("AppendBinary() error = %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("encoded binary = %X, want %X", got, want)
			}
		})
	}
}

func TestAppendBinaryUsesDLCFlagsInWireOrder(t *testing.T) {
	rules := Rules{
		Layout:  LayoutArma3,
		Version: 3,
		DLC: []DLCInfo{
			{Flag: 0x0800, Hash: 0x22222222},
			{Flag: 0x0400, Hash: 0x11111111},
		},
	}

	data, err := AppendBinary(nil, rules)
	if err != nil {
		t.Fatalf("AppendBinary() error = %v", err)
	}

	want := []byte{
		3, 0, 0x00, 0x0C,
		0, 0,
		0x11, 0x11, 0x11, 0x11,
		0x22, 0x22, 0x22, 0x22,
		0, 0,
	}
	if !bytes.Equal(data, want) {
		t.Fatalf("encoded binary = %X, want %X", data, want)
	}
}

func TestAppendBinaryRejectsCreatorDLC(t *testing.T) {
	_, err := AppendBinary(nil, Rules{
		Layout:     LayoutArma3,
		CreatorDLC: []DLCInfo{{ID: 1042220}},
	})
	if !errors.Is(err, ErrEncodeCreatorDLC) {
		t.Fatalf("AppendBinary() error = %v, want ErrEncodeCreatorDLC", err)
	}
}

func TestAppendBinaryEmitsFixedArma3DifficultyWidth(t *testing.T) {
	data, err := AppendBinary(nil, Rules{Layout: LayoutArma3})
	if err != nil {
		t.Fatalf("AppendBinary() error = %v", err)
	}

	if want := []byte{3, 0, 0, 0, 0, 0, 0, 0}; !bytes.Equal(data, want) {
		t.Fatalf("encoded binary = %X, want %X", data, want)
	}
}

func TestAppendBinaryRejectsDayZDifficulty(t *testing.T) {
	_, err := AppendBinary(nil, Rules{
		Layout:     LayoutDayZ,
		Difficulty: &Difficulty{},
	})
	if !errors.Is(err, ErrEncode) {
		t.Fatalf("AppendBinary() error = %v, want ErrEncode", err)
	}
}
