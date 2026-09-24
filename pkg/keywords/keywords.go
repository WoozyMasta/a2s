// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

// Package keywords provides parsers for sv_tag values from A2S_INFO responses.
package keywords

import (
	"fmt"
	"strconv"

	"github.com/woozymasta/a2s/pkg/appid"
)

// Parse selects the keyword parser for the supplied Steam application ID.
func Parse(id uint64, keywords []string) (any, error) {
	switch id {
	case appid.Arma3:
		data := &Arma3{}
		data.Parse(keywords)
		return data, nil

	case appid.DayZ, appid.DayZExperimental:
		data := &DayZ{}
		data.Parse(keywords)
		return data, nil

	default:
		return nil, fmt.Errorf("unsupported application ID %d", id)
	}
}

// parseBool returns true if the value is "t", false otherwise.
func parseBool(val string) bool {
	return val == "t"
}

// ParseUint8 parses a string into a uint8 with overflow checking.
func ParseUint8(val string) uint8 {
	num, err := strconv.ParseUint(val, 10, 8)
	if err != nil || num > 255 {
		return 0
	}

	return uint8(num) // #nosec G115
}

// ParseUint16 parses a string into a uint16 with overflow checking.
func ParseUint16(val string) uint16 {
	num, err := strconv.ParseUint(val, 10, 16)
	if err != nil || num > 65535 {
		return 0
	}

	return uint16(num) // #nosec G115
}

// parseUint32 parses a string into a uint32 with overflow checking.
func parseUint32(val string) uint32 {
	num, err := strconv.ParseUint(val, 10, 32)
	if err != nil || num > 4294967295 {
		return 0
	}

	return uint32(num) // #nosec G115
}

// parseFloat64 parses a string into float64.
func parseFloat64(val string) float64 {
	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.0
	}

	return num
}

// parseCoordinates parses a coordinate string formatted as "lon-lat",
// where lon and lat can be negative.
//
// Examples: "-21--52", "11--22", "-15-32", "7-32"
//
// Returns:
//   - longitude as int32
//   - latitude as int32
func parseCoordinates(val string) (int32, int32) {
	dashIdx := -1
	for i := 1; i < len(val); i++ {
		if val[i] == '-' {
			if val[i-1] >= '0' && val[i-1] <= '9' {
				dashIdx = i
				break
			}
		}
	}

	if dashIdx <= 0 || dashIdx >= len(val)-1 {
		return 0, 0
	}

	lonStr := val[:dashIdx]
	latStr := val[dashIdx+1:]

	lon, err1 := parseInt32(lonStr)
	lat, err2 := parseInt32(latStr)

	if err1 != nil || err2 != nil {
		return 0, 0
	}

	return lon, lat
}

// parseInt32 parses a string into int32.
func parseInt32(s string) (int32, error) {
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(n), nil // #nosec G115 -- ParseInt constrains the value to int32.
}
