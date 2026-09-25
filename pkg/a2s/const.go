// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"time"
)

const (
	// DefaultDeadlineTimeout is the default UDP read deadline.
	DefaultDeadlineTimeout time.Duration = 5 * time.Second
	// edfPort indicates that the game port follows the base response.
	edfPort EDF = 0x80
	// edfSteamID indicates that the server SteamID follows the base response.
	edfSteamID EDF = 0x10
	// edfSourceTV indicates that SourceTV port and name follow the base response.
	edfSourceTV EDF = 0x40
	// edfKeywords indicates that server keywords follow the base response.
	edfKeywords EDF = 0x20
	// edfGameID indicates that the full 64-bit game ID follows the base response.
	edfGameID EDF = 0x01
)
