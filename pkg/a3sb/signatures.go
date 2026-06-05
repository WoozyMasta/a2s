package a3sb

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/wire"
)

// readSignatures parses the signature list from an A3SB response.
func (r *Rules) readSignatures(reader *wire.Decoder) error {
	signCount, err := reader.Byte()
	if err != nil {
		return err
	}

	if signCount == 0 {
		return nil
	}

	r.Signatures = make([]string, 0, int(signCount))

	for i := 0; i < int(signCount); i++ {
		signLen, err := reader.Byte()
		if err != nil {
			return fmt.Errorf("%d length: %w", i, err)
		}
		if signLen == 0 {
			continue
		}

		signature, err := reader.FixedString(int(signLen))
		if err != nil {
			return fmt.Errorf("%d name: %w", i, err)
		}

		r.Signatures = append(r.Signatures, signature)
	}

	return nil
}
