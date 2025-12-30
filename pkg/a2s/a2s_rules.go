package a2s

import (
	"encoding/base64"
	"errors"
	"strconv"
	"unicode/utf8"

	"github.com/woozymasta/a2s/internal/bread"
)

// GetRules queries server rules (A2S_RULES).
// See https://developer.valvesoftware.com/wiki/Server_queries#Response_Format_3
func (c *Client) GetRules() (map[string]string, error) {
	data, _, _, err := c.Get(RulesRequest)
	if err != nil {
		return nil, err
	}

	if cap(c.parseData) < len(data) {
		c.parseData = make([]byte, len(data)+64)
	}
	c.parseData = c.parseData[:len(data)]
	copy(c.parseData, data)

	reader := bread.NewReader(c.parseData)
	count, err := reader.Uint16()
	if err != nil {
		return nil, errors.Join(ErrRuleCount, err)
	}

	if count == 0 {
		return nil, nil
	}

	rules := make(map[string]string, int(count))

	for i := 0; i < int(count); i++ {
		if reader.Len() < 4 {
			return nil, ErrInsufficientData
		}

		key, err := reader.String()
		if err != nil {
			return nil, errors.Join(ErrRuleKey, err)
		}

		value, err := reader.String()
		if err != nil {
			return nil, errors.Join(ErrRuleValue, err)
		}

		rules[key] = value
	}

	return rules, nil
}

// GetParsedRules queries server rules and parses values into appropriate types.
// Attempts to parse as int64, float64, bool, or base64-encoded string. Falls back to string if parsing fails.
func (c *Client) GetParsedRules() (map[string]any, error) {
	data, _, _, err := c.Get(RulesRequest)
	if err != nil {
		return nil, err
	}

	if cap(c.parseData) < len(data) {
		c.parseData = make([]byte, len(data)+64)
	}
	c.parseData = c.parseData[:len(data)]
	copy(c.parseData, data)

	reader := bread.NewReader(c.parseData)
	count, err := reader.Uint16()
	if err != nil {
		return nil, errors.Join(ErrRuleCount, err)
	}

	if count == 0 {
		return nil, nil
	}

	rules := make(map[string]any, int(count))

	var base64Buf []byte

	for i := 0; i < int(count); i++ {
		if reader.Len() < 4 {
			return nil, ErrInsufficientData
		}

		key, err := reader.String()
		if err != nil {
			return nil, errors.Join(ErrRuleKey, err)
		}

		value, err := reader.String()
		if err != nil {
			return nil, errors.Join(ErrRuleValue, err)
		}

		parsed := parseRuleValue(value, &base64Buf)
		rules[key] = parsed
	}

	return rules, nil
}

// parseRuleValue attempts to parse value string into int64, float64, bool, or base64-decoded string.
// Uses reusable base64Buf to minimize allocations.
func parseRuleValue(v string, base64Buf *[]byte) any {
	vLen := len(v)
	if vLen == 0 {
		return v
	}

	switch vLen {
	case 4:
		if v[0] == 't' && v[1] == 'r' && v[2] == 'u' && v[3] == 'e' {
			return true
		}
	case 5:
		if v[0] == 'f' && v[1] == 'a' && v[2] == 'l' && v[3] == 's' && v[4] == 'e' {
			return false
		}
	}

	first := v[0]
	if (first >= '0' && first <= '9') || (first == '-' && vLen > 1 && v[1] >= '0' && v[1] <= '9') {
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			return num
		}
	}

	var floatStr string
	if vLen > 1 && v[vLen-1] == 'f' {
		floatStr = v[:vLen-1]
	} else {
		floatStr = v
	}
	if num, err := strconv.ParseFloat(floatStr, 64); err == nil {
		return num
	}

	if vLen%4 == 0 && vLen > 8 {
		decodedLen := base64.StdEncoding.DecodedLen(vLen)
		if cap(*base64Buf) < decodedLen {
			*base64Buf = make([]byte, decodedLen)
		}
		buf := (*base64Buf)[:decodedLen]

		n, err := base64.StdEncoding.Decode(buf, []byte(v))
		if err == nil && n > 0 {
			decoded := buf[:n]
			if utf8.Valid(decoded) {
				return string(decoded)
			}
		}
	}

	return v
}
