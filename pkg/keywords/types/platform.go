// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package types

// Platform represents a game-server operating system keyword.
type Platform string

const (
	OSWLinux  Platform = "l" // Linux
	OSMac     Platform = "m" // MacOS
	OSOther   Platform = "o" // Other
	OSWindows Platform = "w" // Windows
)

// String returns the human-readable operating system name.
func (p Platform) String() string {
	switch p {
	case OSWLinux:
		return "Linux"
	case OSMac:
		return "MacOS"
	case OSOther:
		return "Other"
	case OSWindows:
		return "Windows"
	default:
		return "Undefined"
	}
}
