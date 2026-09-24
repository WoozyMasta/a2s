// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a3sb

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sort"
)

// AppendBinary appends one unescaped A3SB binary message to dst.
//
// The result is the inner binary payload only. A3SB escape encoding,
// A2S_RULES page generation, and UDP packet framing are separate layers.
func AppendBinary(dst []byte, rules Rules) ([]byte, error) {
	version, err := binaryVersion(rules)
	if err != nil {
		return dst, err
	}

	dlcs, dlcMask, err := orderedDLC(rules.DLC)
	if err != nil {
		return dst, err
	}
	if err := validateBinaryModel(rules, dlcs); err != nil {
		return dst, err
	}

	dst = append(dst, version)
	if rules.Flags != nil {
		dst = append(dst, byte(*rules.Flags))
	} else {
		dst = append(dst, 0)
	}
	dst = binary.LittleEndian.AppendUint16(dst, dlcMask)

	if rules.Layout == LayoutArma3 {
		dst = appendDifficulty(dst, rules.Difficulty)
	}

	for _, dlc := range dlcs {
		dst = binary.LittleEndian.AppendUint32(dst, dlc.Hash)
	}

	dst = appendMods(dst, rules.Mods)
	dst = appendSignatures(dst, rules.Signatures)

	if rules.Layout == LayoutDayZ {
		dst = appendLengthPrefixedString(dst, rules.Description)
	}

	return dst, nil
}

// binaryVersion validates or supplies the protocol version for the selected layout.
func binaryVersion(rules Rules) (byte, error) {
	version := rules.Version
	if version == 0 {
		switch rules.Layout {
		case LayoutArma3:
			version = 3

		case LayoutDayZ:
			version = 2
		}
	}

	switch rules.Layout {
	case LayoutArma3:
		if version != 2 && version != 3 {
			return 0, fmt.Errorf("%w: Arma 3 version %d is not supported", ErrEncode, version)
		}

	case LayoutDayZ:
		if version != 2 {
			return 0, fmt.Errorf("%w: DayZ version %d is not supported", ErrEncode, version)
		}

	default:
		return 0, fmt.Errorf("%w: layout %s is not encodable", ErrEncode, rules.Layout)
	}

	return version, nil
}

// orderedDLC validates DLC flags and returns entries in wire mask order.
func orderedDLC(dlcs []DLCInfo) ([]DLCInfo, uint16, error) {
	ordered := append([]DLCInfo(nil), dlcs...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Flag < ordered[j].Flag
	})

	var mask uint16
	for _, dlc := range ordered {
		flag := uint16(dlc.Flag)
		if flag == 0 || flag&(flag-1) != 0 {
			return nil, 0, fmt.Errorf("%w: DLC flag 0x%X is not one bit", ErrEncode, flag)
		}
		if mask&flag != 0 {
			return nil, 0, fmt.Errorf("%w: duplicate DLC flag 0x%X", ErrEncode, flag)
		}
		mask |= flag
	}

	return ordered, mask, nil
}

// validateBinaryModel rejects model values that have no proven binary representation.
func validateBinaryModel(rules Rules, dlcs []DLCInfo) error {
	if len(rules.CreatorDLC) != 0 {
		return errors.Join(ErrEncode, ErrEncodeCreatorDLC)
	}
	if rules.Layout == LayoutArma3 && rules.Description != "" {
		return fmt.Errorf("%w: Arma 3 has no description field", ErrEncode)
	}
	if rules.Layout == LayoutDayZ && rules.Difficulty != nil {
		return fmt.Errorf("%w: DayZ has no difficulty field", ErrEncode)
	}
	if rules.Difficulty != nil && (rules.Difficulty.Level > 7 || rules.Difficulty.AILevel > 7) {
		return fmt.Errorf("%w: difficulty level exceeds three bits", ErrEncode)
	}
	if len(dlcs) > 16 {
		return fmt.Errorf("%w: too many DLC entries", ErrEncode)
	}
	if len(rules.Mods) > math.MaxUint8 {
		return fmt.Errorf("%w: too many mods", ErrEncode)
	}

	for index, mod := range rules.Mods {
		if mod.IDLength != 0 && mod.IDLength != 1 && mod.IDLength != 4 && mod.IDLength != 8 {
			return fmt.Errorf("%w: mod %d ID length %d is unsupported", ErrEncode, index, mod.IDLength)
		}
		idLength := mod.IDLength
		if idLength == 0 {
			idLength = 4
		}
		if idLength == 1 && mod.ID > math.MaxUint8 {
			return fmt.Errorf("%w: mod %d ID does not fit one byte", ErrEncode, index)
		}
		if idLength == 4 && mod.ID > math.MaxUint32 {
			return fmt.Errorf("%w: mod %d ID does not fit four bytes", ErrEncode, index)
		}
		if len([]byte(mod.Name)) > math.MaxUint8 {
			return fmt.Errorf("%w: mod %d name is too long", ErrEncode, index)
		}
	}

	for index, signature := range rules.Signatures {
		if len([]byte(signature)) > math.MaxUint8 {
			return fmt.Errorf("%w: signature %d is too long", ErrEncode, index)
		}
	}

	if len(rules.Signatures) > math.MaxUint8 {
		return fmt.Errorf("%w: too many signatures", ErrEncode)
	}
	if len([]byte(rules.Description)) > math.MaxUint8 {
		return fmt.Errorf("%w: description is too long", ErrEncode)
	}

	return nil
}

// appendDifficulty appends the fixed-width Arma 3 difficulty section.
func appendDifficulty(dst []byte, difficulty *Difficulty) []byte {
	if difficulty == nil {
		return append(dst, 0, 0)
	}

	value := difficulty.Level & 0b00000111
	value |= (difficulty.AILevel & 0b00000111) << 3
	if !difficulty.AdvanceFlight {
		value |= 1 << 6
	}
	if difficulty.ThirdPerson {
		value |= 1 << 7
	}

	var crosshair byte
	if difficulty.Crosshair {
		crosshair = 1
	}

	return append(dst, value, crosshair)
}

// appendMods appends the regular mod records in their existing order.
func appendMods(dst []byte, mods []Mod) []byte {
	dst = append(dst, byte(len(mods))) // #nosec G115 -- mod count is validated to fit uint8.
	for _, mod := range mods {
		idLength := mod.IDLength
		if idLength == 0 {
			idLength = 4
		}

		dst = binary.LittleEndian.AppendUint32(dst, mod.Hash)
		dst = append(dst, idLength)

		switch idLength {
		case 1:
			// #nosec G115 -- ID is validated to fit one byte.
			dst = append(dst, byte(mod.ID))

		case 4:
			// #nosec G115 -- validated before encoding.
			dst = binary.LittleEndian.AppendUint32(dst, uint32(mod.ID))

		case 8:
			dst = binary.LittleEndian.AppendUint64(dst, mod.ID)
		}

		dst = appendLengthPrefixedString(dst, mod.Name)
	}

	return dst
}

// appendSignatures appends length-prefixed signature names.
func appendSignatures(dst []byte, signatures []string) []byte {
	// #nosec G115 -- signature count is validated to fit uint8.
	dst = append(dst, byte(len(signatures)))

	for _, signature := range signatures {
		dst = appendLengthPrefixedString(dst, signature)
	}

	return dst
}

// appendLengthPrefixedString appends one uint8-length-prefixed string.
func appendLengthPrefixedString(dst []byte, value string) []byte {
	data := []byte(value)

	// #nosec G115 -- callers validate the uint8 length prefix.
	dst = append(dst, byte(len(data)))

	return append(dst, data...)
}
