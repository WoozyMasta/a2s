// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a3sb

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/wire"
)

// Flags represents the currently undocumented bit flags from an A3SB response.
// Unknown bits are retained so a response can be encoded without loss.
type Flags byte

// Has reports whether the zero-based bit index is set.
func (f Flags) Has(bit uint8) bool {
	if bit >= 8 {
		return false
	}

	return f&(1<<bit) != 0
}

// readFlags reads the optional flags byte and keeps zero flags as nil.
func (r *Rules) readFlags(reader *wire.Decoder) error {
	value, err := reader.Byte()
	if err != nil {
		return fmt.Errorf("flags: %w", err)
	}
	if value == 0 {
		return nil
	}

	flags := Flags(value)
	r.Flags = &flags

	return nil
}
