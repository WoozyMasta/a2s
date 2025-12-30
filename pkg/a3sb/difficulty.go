package a3sb

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/bread"
	"github.com/woozymasta/steam/utils/appid"
)

// Difficulty represents Arma 3 server difficulty settings as bits:
//   - 0 - newbie
//   - 1 - normal
//   - 2 - expert
//   - 3 - custom
type Difficulty struct {
	Level         byte `json:"level"`          // First 3 bits
	AILevel       byte `json:"level_ai"`       // Bits 3-5
	AdvanceFlight bool `json:"advance_flight"` // Bit 6
	ThirdPerson   bool `json:"third_person"`   // Bit 7
	Crosshair     bool `json:"crosshair"`      // Second byte, bit 0
}

// readDifficulty parses difficulty settings (Arma 3 only).
func (r *Rules) readDifficulty(reader *bread.Reader) error {
	if r.id != appid.Arma3.Uint64() {
		return nil
	}

	value, err := reader.Byte()
	if err != nil {
		return fmt.Errorf("first byte: %w", err)
	}
	if value == 0 {
		return nil
	}

	r.Difficulty = &Difficulty{
		Level:         value & 0b00000111,        // Mask for the first 3 bits
		AILevel:       (value >> 3) & 0b00000111, // Shift 3 bits right, then mask for next 3 bits
		AdvanceFlight: value&(1<<6) == 0,         // Checking bit 6
		ThirdPerson:   value&(1<<7) != 0,         // Checking bit 7
	}

	crosshair, err := reader.Byte()
	if err != nil {
		return fmt.Errorf("second byte: %w", err)
	}
	r.Difficulty.Crosshair = (crosshair & 0x01) != 0

	return nil
}
