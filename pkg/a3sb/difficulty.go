package a3sb

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/internal/bread"
	"github.com/woozymasta/steam/utils/appid"
)

// Difficulty Arma 3 structure, bytes for level are represented as:
//   - 0 - newbie
//   - 1 - normal
//   - 2 - expert
//   - 3 - custom
type Difficulty struct {
	// First difficulty byte

	Level         byte `json:"level"`          // First 3 bits (0, 1, 2)
	AILevel       byte `json:"level_ai"`       // Second 3 bits (3, 4, 5)
	AdvanceFlight bool `json:"advance_flight"` // 6 bit
	ThirdPerson   bool `json:"third_person"`   // 7 bit

	// Second difficulty byte

	Crosshair bool `json:"crosshair"` // First bit
}

// Read difficulty from Arma 3 server browser proto
func (r *Rules) readDifficulty(buf *bytes.Buffer) error {
	if r.id != appid.Arma3.Uint64() {
		return nil
	}

	value, err := bread.Byte(buf)
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

	crosshair, err := bread.Byte(buf)
	if err != nil {
		return fmt.Errorf("second byte: %w", err)
	}
	r.Difficulty.Crosshair = (crosshair&0x01 != 0) // Isolate first bit

	return nil
}
