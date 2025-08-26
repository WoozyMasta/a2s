package a2s

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/woozymasta/a2s/internal/bread"
)

// GetRules return A2S_RULES
// https://developer.valvesoftware.com/wiki/Server_queries#Response_Format_3
func (c *Client) GetRules() (map[string]string, error) {
	data, _, _, err := c.Get(RulesRequest)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)

	count, err := bread.Uint16(buf)
	if err != nil {
		return nil, fmt.Errorf("%w rules count: 0x%X", ErrRuleRead, data[:4])
	}

	if count == 0 {
		return nil, nil
	}

	rules := make(map[string]string)

	for i := 0; i < int(count); i++ {
		if buf.Len() < 4 {
			return nil, fmt.Errorf("%w in rule %d", ErrInsufficientData, i)
		}

		key, err := bread.String(buf)
		if err != nil {
			return nil, fmt.Errorf("%w key: %w", ErrRuleRead, err)
		}

		value, err := bread.String(buf)
		if err != nil {
			return nil, fmt.Errorf("%w value for key '%s': %w", ErrRuleRead, key, err)
		}

		rules[key] = value
	}

	return rules, nil
}

// GetParsedRules return A2S_RULES and try parse values to int64, float64, boolean and decode base64, save strings by default
func (c *Client) GetParsedRules() (map[string]any, error) {
	data, err := c.GetRules()
	if err != nil {
		return nil, err
	}

	rules := make(map[string]any)

	for k, v := range data {
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			rules[k] = num // Try parse integer
		} else if num, err := strconv.ParseFloat(strings.TrimSuffix(v, "f"), 64); err == nil {
			rules[k] = num // Try parse float
		} else if boolean, err := strconv.ParseBool(v); err == nil {
			rules[k] = boolean // try parse boolean
		} else if decoded, err := base64.StdEncoding.DecodeString(v); err == nil && utf8.Valid(decoded) {
			rules[k] = string(decoded) // try parse base64
		} else {
			rules[k] = v // Save as is
		}
	}

	return rules, nil
}
