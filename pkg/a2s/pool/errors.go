// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import "errors"

var (
	// ErrClosed indicates that the pool is closed or is shutting down.
	ErrClosed = errors.New("a2s pool is closed")
	// ErrInvalidTarget indicates that a target address cannot be queried.
	ErrInvalidTarget = errors.New("a2s pool target is invalid")
	// ErrInvalidQuery indicates that a query type is not supported by the pool.
	ErrInvalidQuery = errors.New("a2s pool query type is invalid")
	// ErrInvalidConfig indicates that a pool option is invalid.
	ErrInvalidConfig = errors.New("a2s pool configuration is invalid")
	// ErrNilContext indicates that a nil context was passed to a pool operation.
	ErrNilContext = errors.New("a2s pool context is nil")
)
