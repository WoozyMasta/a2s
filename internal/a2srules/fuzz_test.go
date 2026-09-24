// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2srules

import "testing"

func FuzzParse(f *testing.F) {
	f.Add([]byte{0, 0})
	f.Add([]byte{1, 0, 'm', 'o', 'd', 'e', 0, 'c', 'o', 'o', 'p', 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		result, _ := Parse(data)
		_ = Map(result.Entries)
	})
}
