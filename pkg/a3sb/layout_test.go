// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a3sb

import (
	"testing"

	"github.com/woozymasta/a2s/internal/wire"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/appid"
)

func TestLayoutSelectionIsIndependentFromAppID(t *testing.T) {
	rules := &Rules{Layout: LayoutDayZ, appID: appid.DayZExperimental}
	if err := rules.readA3SB([]byte{2, 0, 0, 0, 0, 0}); err != nil {
		t.Fatalf("readA3SB() error = %v", err)
	}
	if rules.Layout != LayoutDayZ {
		t.Fatalf("layout = %v, want %v", rules.Layout, LayoutDayZ)
	}
	if rules.GetAppID() != appid.DayZExperimental {
		t.Fatalf("AppID = %d, want %d", rules.GetAppID(), appid.DayZExperimental)
	}
}

func TestFlagsPreserveUnknownBits(t *testing.T) {
	rules := &Rules{}
	decoder := wire.NewDecoder([]byte{0xA5})
	if err := rules.readFlags(&decoder); err != nil {
		t.Fatalf("readFlags() error = %v", err)
	}
	if rules.Flags == nil {
		t.Fatal("Flags is nil")
	}
	if got, want := *rules.Flags, Flags(0xA5); got != want {
		t.Fatalf("flags = 0x%X, want 0x%X", got, want)
	}
	for _, bit := range []uint8{0, 2, 5, 7} {
		if !rules.Flags.Has(bit) {
			t.Fatalf("Has(%d) = false, want true", bit)
		}
	}
	for _, bit := range []uint8{1, 3, 4, 6, 8} {
		if rules.Flags.Has(bit) {
			t.Fatalf("Has(%d) = true, want false", bit)
		}
	}
}

func TestDayZPlatformKeepsRawValue(t *testing.T) {
	rules := &Rules{}
	if err := rules.parseRulesDayZ(a2sRules("platform", "?")); err != nil {
		t.Fatalf("parseRulesDayZ() error = %v", err)
	}
	if rules.Platform != "Linux" {
		t.Fatalf("normalized platform = %q, want Linux", rules.Platform)
	}
	if rules.PlatformRaw != "?" {
		t.Fatalf("raw platform = %q, want ?", rules.PlatformRaw)
	}
}

func a2sRules(name, value string) a2s.Rules {
	return a2s.Rules{{Name: name, Value: value}}
}
