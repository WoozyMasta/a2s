// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a3sb

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/woozymasta/a2s/pkg/a2s"
)

const (
	// DefaultPageSize is the current documented A3SB page value size.
	// A non-zero page size can be supplied to EncodePages
	// when a different server-compatible limit is required.
	DefaultPageSize = 124

	// maxPageCount is the number of pages representable by the A3SB key format.
	maxPageCount = 255
)

// EncodePages splits escaped A3SB binary data into ordered A2S_RULES pages.
//
// A page size of zero uses DefaultPageSize.
// The input must already be escaped with AppendEscapeSequences;
// page generation does not perform escaping.
func EncodePages(data []byte, pageSize int) (a2s.Rules, error) {
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if pageSize < 1 {
		return nil, errors.Join(ErrPageEncode, ErrPageSize)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: empty data", ErrPageEncode)
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return nil, fmt.Errorf("%w: input contains unescaped NUL byte", ErrPageEncode)
	}

	pageCount := len(data) / pageSize
	if len(data)%pageSize != 0 {
		pageCount++
	}
	if pageCount > maxPageCount {
		return nil, fmt.Errorf("%w: %d pages exceed %d", ErrPageCount, pageCount, maxPageCount)
	}

	rules := make(a2s.Rules, 0, pageCount)
	for pageIndex := 0; pageIndex < pageCount; pageIndex++ {
		start := pageIndex * pageSize
		end := start + pageSize
		if end > len(data) {
			end = len(data)
		}

		pageNumber := byte(pageIndex + 1) // #nosec G115 -- page count is bounded to 255.
		pageTotal := byte(pageCount)      // #nosec G115 -- page count is bounded to 255.

		rules = append(rules, a2s.Rule{
			Name:  string([]byte{pageNumber, pageTotal}),
			Value: string(data[start:end]),
		})
	}

	return rules, nil
}
