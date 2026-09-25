// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"encoding/binary"
	"fmt"
	"slices"
	"strings"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

// AppendRules appends one complete logical A2S_RULES response to dst.
//
// Rules are encoded in their existing slice order, including duplicate names.
func AppendRules(dst []byte, rules Rules) ([]byte, error) {
	if len(rules) > 65535 {
		return dst, fmt.Errorf("%w: %d rules exceed uint16 count", ErrRuleEncode, len(rules))
	}
	if err := validateRulesForEncoding(rules); err != nil {
		return dst, err
	}

	dst = slices.Grow(dst, rulesEncodedSize(rules))
	dst = transport.AppendPacketHeader(dst, transport.ResponseRules)
	// #nosec G115 -- count is validated to fit uint16.
	dst = binary.LittleEndian.AppendUint16(dst, uint16(len(rules)))
	for _, rule := range rules {
		dst = append(dst, rule.Name...)
		dst = append(dst, 0)
		dst = append(dst, rule.Value...)
		dst = append(dst, 0)
	}

	return dst, nil
}

// rulesEncodedSize returns the number of bytes appended by AppendRules.
func rulesEncodedSize(rules Rules) int {
	size := transport.PacketHeaderSize + 2
	for _, rule := range rules {
		size += len(rule.Name) + len(rule.Value) + 2
	}

	return size
}

// validateRulesForEncoding rejects rule strings that cannot be represented as C strings.
func validateRulesForEncoding(rules Rules) error {
	for _, rule := range rules {
		if strings.IndexByte(rule.Name, 0) >= 0 {
			return fmt.Errorf("%w: rule name contains NUL", ErrRuleEncode)
		}
		if strings.IndexByte(rule.Value, 0) >= 0 {
			return fmt.Errorf("%w: rule value contains NUL", ErrRuleEncode)
		}
	}

	return nil
}
