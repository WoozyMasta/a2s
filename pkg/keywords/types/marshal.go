// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

//go:build !a2s_no_marshal

package types //nolint:list

import (
	"encoding/json"
)

// MarshalJSON encodes GameType using its human-readable string.
func (gt GameType) MarshalJSON() ([]byte, error) {
	return json.Marshal(gt.String())
}

// MarshalJSON encodes ServerLang using its human-readable string.
func (sl ServerLang) MarshalJSON() ([]byte, error) {
	return json.Marshal(sl.String())
}

// MarshalJSON encodes Platform using its human-readable string.
func (p Platform) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// MarshalJSON encodes ServerState using its human-readable string.
func (ss ServerState) MarshalJSON() ([]byte, error) {
	return json.Marshal(ss.String())
}
