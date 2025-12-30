// Package keywords provide additional parsers for tags (sv_tag) from the A2S_INFO response
package keywords

import (
	"fmt"
	"strconv"

	"github.com/woozymasta/steam/utils/appid"
)

// Parse universal function for outputting result depending on application ID,
// if parser exists it will return updated structure, otherwise it will return error
func Parse(id uint64, keywords []string) (any, error) {
	switch id {
	case appid.Arma3.Uint64():
		data := &Arma3{}
		data.Parse(keywords)
		return data, nil

	case appid.DayZ.Uint64(), appid.DayZExp.Uint64():
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

// parseUint32 parses a string into a uint16 with overflow checking.
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
// where lon and lat can be negative. Examples:
// "-21--52", "11--22", "-15-32", "7-32"
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
	if len(s) == 0 {
		return 0, strconv.ErrSyntax
	}

	neg := false
	start := 0
	if s[0] == '-' {
		neg = true
		start = 1
		if len(s) == 1 {
			return 0, strconv.ErrSyntax
		}
	}

	var n int32
	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			break // Stop at first non-digit (for cases like "2-3")
		}

		digit := int32(s[i] - '0')
		if neg && n == 214748364 && digit == 8 && i == len(s)-1 {
			return -2147483648, nil
		}

		if n > (2147483647-digit)/10 {
			return 0, strconv.ErrRange
		}

		n = n*10 + digit
	}

	if neg {
		n = -n
	}

	return n, nil
}
