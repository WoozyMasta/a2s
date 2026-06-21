package a2s

import (
	"encoding/binary"
	"fmt"
	"strings"
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

	dst = appendPacketHeader(dst, ResponseRules)
	// #nosec G115 -- count is validated to fit uint16.
	for _, rule := range rules {
		dst = binary.LittleEndian.AppendUint16(dst, uint16(len(rules)))
		dst = append(dst, rule.Name...)
		dst = append(dst, 0)
		dst = append(dst, rule.Value...)
		dst = append(dst, 0)
	}

	return dst, nil
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
