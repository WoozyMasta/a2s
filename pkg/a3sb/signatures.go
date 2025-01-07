package a3sb

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/pkg/bread"
)

// Read signatures from Arma 3 server browser proto
func (r *Rules) readSignatures(buf *bytes.Buffer) error {
	signCount, err := bread.Byte(buf)
	if err != nil {
		return err
	}

	if signCount == 0 {
		return nil
	}

	for i := 0; i < int(signCount); i++ {
		signLen, err := bread.Byte(buf)
		if err != nil {
			return fmt.Errorf("%d length: %w", i, err)
		}
		if signLen == 0 {
			continue
		}

		signature, err := bread.StringLen(buf, int(signLen))
		if err != nil {
			return fmt.Errorf("%d name: %w", i, err)
		}

		r.Signatures = append(r.Signatures, signature)
	}

	return nil
}
