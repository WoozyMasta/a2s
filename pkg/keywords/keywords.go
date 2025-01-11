// Package keywords provide additional parsers for tags (sv_tag) from the A2S_INFO response
package keywords

import (
	"fmt"
	"strconv"

	"github.com/woozymasta/a2s/pkg/appid"
)

// Universal function for outputting result depending on application ID,
// if parser exists it will return updated structure, otherwise it will return error
func Parse(id uint64, keywords []string) (any, error) {
	switch id {
	case appid.Arma3:
		data := &Arma3{}
		data.Parse(keywords)
		return data, nil

	case appid.DayZ, appid.DayZExp:
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

// parseInt32 parses a string into a uint8 with overflow checking.
func parseInt32(val string) int32 {
	num, err := strconv.ParseInt(val, 10, 32)
	if err != nil || num >= 2147483648 || num <= -2147483648 {
		return 0
	}

	return int32(num)
}

// parseUint8 parses a string into a uint8 with overflow checking.
func ParseUint8(val string) uint8 {
	num, err := strconv.ParseUint(val, 10, 8)
	if err != nil || num > 255 {
		return 0
	}

	return uint8(num)
}

// parseUint16 parses a string into a uint16 with overflow checking.
func ParseUint16(val string) uint16 {
	num, err := strconv.ParseUint(val, 10, 16)
	if err != nil || num > 65535 {
		return 0
	}

	return uint16(num)
}

// parseUint32 parses a string into a uint16 with overflow checking.
func parseUint32(val string) uint32 {
	num, err := strconv.ParseUint(val, 10, 32)
	if err != nil || num > 4294967295 {
		return 0
	}

	return uint32(num)
}

// parseFloat64 parses a string into float64.
func parseFloat64(val string) float64 {
	num, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.0
	}

	return num
}
