package a3sb

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/pkg/bread"
)

// Structure for storing information about mods
type Mod struct {
	Name string `json:"name,omitempty"` // Mod name from response
	ID   uint64 `json:"id,omitempty"`   // Mod ID in SteamWorkshop
	Hash uint32 `json:"hash,omitempty"` // Mod short hash
}

// Arma 3 creator DLC stored in mods byte block
var arma3CreatorDLC = map[uint64]string{
	1042220: "Creator DLC: Global Mobilization - Cold War Germany",
	1175380: "Creator DLC: Spearhead 1944",
	1227700: "Creator DLC: S.O.G. Prairie Fire",
	1294440: "Creator DLC: CSLA Iron Curtain",
	1681170: "Creator DLC: Western Sahara",
	2647760: "Creator DLC: Reaction Forces",
	2647830: "Creator DLC: Expeditionary Forces",
}

// Read mods from Arma 3 server browser proto
func (r *Rules) readMods(buf *bytes.Buffer) error {
	modCount, err := bread.Byte(buf)
	if err != nil {
		return fmt.Errorf("mod count: %w", err)
	}
	if modCount == 0 {
		return nil
	}

	for i := 0; i < int(modCount); i++ {
		mod := Mod{}
		creatorDLC := DLCInfo{}

		if mod.Hash, err = bread.Uint32(buf); err != nil {
			return fmt.Errorf("mod %d hash: %w", i, err)
		}

		idLen, err := bread.Byte(buf)
		if err != nil {
			return fmt.Errorf("mod %d id length: %w", i, err)
		}

		switch idLen {
		case 1:
			id, err := bread.Byte(buf)
			if err != nil {
				return fmt.Errorf("mod %d id length: %w", i, err)
			}
			mod.ID = uint64(id)

		case 4:
			id, err := bread.Uint32(buf)
			if err != nil {
				return fmt.Errorf("mod %d id length: %w", i, err)
			}
			mod.ID = uint64(id)

		case 8:
			id, err := bread.Uint64(buf)
			if err != nil {
				return fmt.Errorf("mod %d id length: %w", i, err)
			}
			mod.ID = id

		case 19: // Arma Creators DLC, right way check 4 byte, but this works too, return 00010011
			id, err := bread.Uint32(buf)
			if err != nil {
				return fmt.Errorf("mod %d id length: %w", i, err)
			}
			creatorDLC.ID = uint64(id)
			creatorDLC.Name = arma3CreatorDLC[creatorDLC.ID]
			r.CreatorDLC = append(r.CreatorDLC, creatorDLC)
			continue

		default:
			return fmt.Errorf("mod %d id length (%d) unknown", i, idLen)
		}

		nameLen, err := bread.Byte(buf)
		if err != nil {
			return fmt.Errorf("mod %d name length: %w", i, err)
		}

		if nameLen != 0 {
			if mod.Name, err = bread.StringLen(buf, int(nameLen)); err != nil {
				return fmt.Errorf("mod %d hash: %w", i, err)
			}
		}

		r.Mods = append(r.Mods, mod)
	}

	return nil
}
