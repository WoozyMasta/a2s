// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a3sb

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/wire"
)

// Mod contains mod information from an A3SB response.
type Mod struct {
	Name     string `json:"name,omitempty"`      // Mod name from response.
	ID       uint64 `json:"id,omitempty"`        // Mod ID in SteamWorkshop.
	Hash     uint32 `json:"hash,omitempty"`      // Mod short hash.
	IDLength byte   `json:"id_length,omitempty"` // Wire idLen value.
}

// arma3CreatorDLC maps creator DLC AppIDs found in the mods block to names.
var arma3CreatorDLC = map[uint64]string{
	1042220: "Creator DLC: Global Mobilization - Cold War Germany",
	1175380: "Creator DLC: Spearhead 1944",
	1227700: "Creator DLC: S.O.G. Prairie Fire",
	1294440: "Creator DLC: CSLA Iron Curtain",
	1681170: "Creator DLC: Western Sahara",
	2647760: "Creator DLC: Reaction Forces",
	2647830: "Creator DLC: Expeditionary Forces",
}

// readMods parses mods and creator DLC from an A3SB response.
func (r *Rules) readMods(reader *wire.Decoder) error {
	modCount, err := reader.Byte()
	if err != nil {
		return fmt.Errorf("mod count: %w", err)
	}
	if modCount == 0 {
		return nil
	}

	r.Mods = make([]Mod, 0, int(modCount))
	// Preserve the existing non-nil empty slice
	// without allocating backing storage until a Creator DLC entry is actually present.
	r.CreatorDLC = []DLCInfo{}

	for i := 0; i < int(modCount); i++ {
		var mod Mod
		var creatorDLC DLCInfo

		if mod.Hash, err = reader.Uint32(); err != nil {
			return fmt.Errorf("mod %d hash: %w", i, err)
		}

		idLen, err := reader.Byte()
		if err != nil {
			return fmt.Errorf("mod %d id length: %w", i, err)
		}

		switch idLen {
		case 1:
			// Theoretical short ID form;
			// not observed in available server responses.
			id, err := reader.Byte()
			if err != nil {
				return fmt.Errorf("mod %d id length: %w", i, err)
			}
			mod.ID = uint64(id)

		case 4:
			// Observed standard form: Workshop IDs are encoded as uint32 values.
			// ID 0 is intended for private/local mods.
			id, err := reader.Uint32()
			if err != nil {
				return fmt.Errorf("mod %d id length: %w", i, err)
			}
			mod.ID = uint64(id)

		case 8:
			// Theoretical extended Steam ID form;
			// not observed in available server responses.
			id, err := reader.Uint64()
			if err != nil {
				return fmt.Errorf("mod %d id length: %w", i, err)
			}
			mod.ID = id

		case 19:
			// 0x13 marks a Creator DLC entry, not a 19-byte ID.
			// It is followed by a uint32 Steam AppID and no mod name;
			// the next byte starts the next mod record.
			if cap(r.CreatorDLC) == 0 {
				r.CreatorDLC = make([]DLCInfo, 0, 4)
			}
			id, err := reader.Uint32()
			if err != nil {
				return fmt.Errorf("mod %d id length: %w", i, err)
			}

			creatorDLC.ID = uint64(id)
			creatorDLC.Name = arma3CreatorDLC[creatorDLC.ID]
			r.CreatorDLC = append(r.CreatorDLC, creatorDLC)
			continue

		default:
			// The 2-byte form is mentioned in the protocol notes,
			// but its wire layout is not confirmed by a packet fixture.
			// Keep it unsupported until a real response justifies a parser change.
			return fmt.Errorf("mod %d id length (%d) unknown", i, idLen)
		}

		mod.IDLength = idLen

		nameLen, err := reader.Byte()
		if err != nil {
			return fmt.Errorf("mod %d name length: %w", i, err)
		}

		if nameLen != 0 {
			if mod.Name, err = reader.FixedString(int(nameLen)); err != nil {
				return fmt.Errorf("mod %d hash: %w", i, err)
			}
		}

		r.Mods = append(r.Mods, mod)
	}

	return nil
}
