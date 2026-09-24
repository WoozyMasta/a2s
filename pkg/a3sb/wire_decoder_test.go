// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a3sb

import (
	"encoding/binary"
	"testing"

	"github.com/woozymasta/a2s/internal/wire"
	"github.com/woozymasta/a2s/pkg/appid"
)

func TestReadModsTruncatedLengthPrefixedFields(t *testing.T) {
	data := []byte{1}
	data = binary.LittleEndian.AppendUint32(data, 0x12345678)
	data = append(data, 4)
	data = binary.LittleEndian.AppendUint32(data, 1234)
	data = append(data, 3)
	data = append(data, "mod"...)

	for end := 0; end < len(data); end++ {
		t.Run("truncated", func(t *testing.T) {
			decoder := wire.NewDecoder(data[:end])
			rules := &Rules{}
			if err := rules.readMods(&decoder); err == nil {
				t.Fatalf("readMods() returned nil for length %d", end)
			}
		})
	}
}

func TestReadModsKeepsCreatorDLCSliceContract(t *testing.T) {
	data := []byte{1}
	data = binary.LittleEndian.AppendUint32(data, 0x12345678)
	data = append(data, 19)
	data = binary.LittleEndian.AppendUint32(data, 1042220)

	decoder := wire.NewDecoder(data)
	rules := &Rules{}
	if err := rules.readMods(&decoder); err != nil {
		t.Fatalf("readMods() error = %v", err)
	}

	if rules.CreatorDLC == nil {
		t.Fatal("CreatorDLC is nil, want non-nil empty-or-populated slice")
	}
	if len(rules.CreatorDLC) != 1 {
		t.Fatalf("len(CreatorDLC) = %d, want 1", len(rules.CreatorDLC))
	}
	if got := rules.CreatorDLC[0].ID; got != 1042220 {
		t.Fatalf("CreatorDLC[0].ID = %d, want 1042220", got)
	}
}

func TestReadModsWithoutCreatorDLCKeepsNonNilSlice(t *testing.T) {
	data := []byte{1}
	data = binary.LittleEndian.AppendUint32(data, 0x12345678)
	data = append(data, 4)
	data = binary.LittleEndian.AppendUint32(data, 1234)
	data = append(data, 0)

	decoder := wire.NewDecoder(data)
	rules := &Rules{}
	if err := rules.readMods(&decoder); err != nil {
		t.Fatalf("readMods() error = %v", err)
	}
	if rules.CreatorDLC == nil {
		t.Fatal("CreatorDLC is nil, want a non-nil empty slice")
	}
	if len(rules.CreatorDLC) != 0 {
		t.Fatalf("len(CreatorDLC) = %d, want 0", len(rules.CreatorDLC))
	}
}

func TestReadSignaturesTruncatedLengthPrefixedFields(t *testing.T) {
	data := []byte{1, 3, 'a', 'b', 'c'}

	for end := 0; end < len(data); end++ {
		t.Run("truncated", func(t *testing.T) {
			decoder := wire.NewDecoder(data[:end])
			rules := &Rules{}
			if err := rules.readSignatures(&decoder); err == nil {
				t.Fatalf("readSignatures() returned nil for length %d", end)
			}
		})
	}
}

func TestReadA3SBDayZDescriptionTruncation(t *testing.T) {
	const description = "dayz"

	data := []byte{2, 0, 0, 0, 0, 0, byte(len(description))}
	data = append(data, description...)
	baseLength := len(data) - len(description)

	for end := baseLength; end < len(data); end++ {
		t.Run("truncated description", func(t *testing.T) {
			rules := &Rules{Layout: LayoutDayZ, appID: appid.DayZ}
			if err := rules.readA3SB(data[:end]); err == nil {
				t.Fatalf("readA3SB() returned nil for length %d", end)
			}
		})
	}
}

func TestReadA3SBShortBuffersDoNotPanic(t *testing.T) {
	for length := 0; length < 32; length++ {
		t.Run("short buffer", func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("readA3SB() panicked for length %d: %v", length, recovered)
				}
			}()

			rules := &Rules{Layout: LayoutArma3, appID: appid.Arma3}
			_ = rules.readA3SB(make([]byte, length))
		})
	}
}
