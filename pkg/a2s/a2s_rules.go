// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/woozymasta/a2s/internal/a2srules"
)

// Rule is one ordered A2S_RULES key/value pair.
type Rule struct {
	// Name is the rule key as received on the wire.
	Name string `json:"name" yaml:"name"`

	// Value is the rule value as received on the wire.
	Value string `json:"value" yaml:"value"`
}

// Rules preserves A2S_RULES entry order and duplicate names.
type Rules []Rule

// Get returns the last value for name, matching the old map representation.
func (r Rules) Get(name string) (string, bool) {
	for i := len(r) - 1; i >= 0; i-- {
		if r[i].Name == name {
			return r[i].Value, true
		}
	}

	return "", false
}

// Values returns all values for name in wire order.
func (r Rules) Values(name string) []string {
	var values []string
	for _, rule := range r {
		if rule.Name == name {
			values = append(values, rule.Value)
		}
	}

	return values
}

// Map converts rules to a map using the last value for duplicate names.
// The conversion loses entry order and duplicate values.
func (r Rules) Map() map[string]string {
	if len(r) == 0 {
		return nil
	}

	rules := make(map[string]string, len(r))
	for _, rule := range r {
		rules[rule.Name] = rule.Value
	}

	return rules
}

// GetRules queries server rules (A2S_RULES).
// See https://developer.valvesoftware.com/wiki/Server_queries#Response_Format_3
func (c *Client) GetRules(ctx context.Context) (Rules, error) {
	packet, _, err := c.Query(ctx, RulesRequest)
	if err != nil {
		return nil, err
	}

	return DecodeRules(packet)
}

// DecodeRules parses a logical A2S_RULES response packet.
//
// The response type must be ResponseRules.
// Rule order and duplicate names are preserved in the returned slice.
func DecodeRules(packet Packet) (Rules, error) {
	if packet.Type != ResponseRules {
		return nil, errors.Join(ErrRuleRead, fmt.Errorf("unexpected response type 0x%X", packet.Type))
	}

	result, err := a2srules.Parse(packet.Payload)
	if err != nil {
		switch {
		case errors.Is(err, a2srules.ErrCount):
			return nil, errors.Join(ErrRuleCount, err)

		case errors.Is(err, a2srules.ErrInsufficientData):
			return nil, errors.Join(ErrInsufficientData, err)

		case errors.Is(err, a2srules.ErrKey):
			return nil, errors.Join(ErrRuleKey, err)

		case errors.Is(err, a2srules.ErrValue):
			return nil, errors.Join(ErrRuleValue, err)

		default:
			return nil, err
		}
	}

	rules := make(Rules, 0, len(result.Entries))
	for _, entry := range result.Entries {
		rules = append(rules, Rule{
			Name:  string(entry.Key),
			Value: string(entry.Value),
		})
	}

	return rules, nil
}

// GetParsedRules queries server rules and converts values into convenient Go types.
// Numeric, boolean, and valid UTF-8 Base64 values are converted heuristically.
// Duplicate names use the last value, like Rules.Get and Rules.Map.
func (c *Client) GetParsedRules(ctx context.Context) (map[string]any, error) {
	data, err := c.GetRules(ctx)
	if err != nil {
		return nil, err
	}

	return ParseRuleValues(data), nil
}

// ParseRuleValues converts already fetched A2S rules
// into convenient Go values without issuing another network request.
// Numeric, boolean, and valid UTF-8 Base64 values are converted heuristically.
// The input rules are not modified.
func ParseRuleValues(data Rules) map[string]any {
	if data == nil {
		return nil
	}

	rules := make(map[string]any, len(data))
	var base64Buf []byte
	for _, rule := range data {
		rules[rule.Name] = parseRuleValue(rule.Value, &base64Buf)
	}

	return rules
}

// parseRuleValue attempts to parse a rule value
// as an integer, float, boolean, or UTF-8 Base64 string.
// It returns the original value when no conversion fits.
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

	floatStr := v
	if vLen > 1 && v[vLen-1] == 'f' {
		floatStr = v[:vLen-1]
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
