// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

// Package testfixtures provides deterministic protocol bytes for offline tests.
package testfixtures

import (
	"embed"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"
)

//go:embed data/*.hex
var fixtureFiles embed.FS

// Read returns a fresh decoded copy of the named hexadecimal fixture.
// Fixture names are relative to the embedded data directory.
func Read(name string) ([]byte, error) {
	raw, err := fixtureFiles.ReadFile("data/" + name)
	if err != nil {
		return nil, fmt.Errorf("read fixture %q: %w", name, err)
	}

	compact := strings.Map(func(value rune) rune {
		if unicode.IsSpace(value) {
			return -1
		}

		return value
	}, string(raw))

	data, err := hex.DecodeString(compact)
	if err != nil {
		return nil, fmt.Errorf("decode fixture %q: %w", name, err)
	}

	return data, nil
}
