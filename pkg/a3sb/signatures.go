package a3sb

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/bread"
)

// Read signatures from Arma 3 server browser proto
func (r *Rules) readSignatures(reader *bread.Reader) error {
	signCount, err := reader.Byte()
	if err != nil {
		return err
	}

	if signCount == 0 {
		return nil
	}

	for i := 0; i < int(signCount); i++ {
		signLen, err := reader.Byte()
		if err != nil {
			return fmt.Errorf("%d length: %w", i, err)
		}
		if signLen == 0 {
			continue
		}

		signature, err := reader.StringLen(int(signLen))
		if err != nil {
			return fmt.Errorf("%d name: %w", i, err)
		}

		r.Signatures = append(r.Signatures, signature)
	}

	return nil
}
