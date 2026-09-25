// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package pool

import (
	"fmt"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

const (
	// DefaultWorkers is the default number of concurrent UDP transactions.
	DefaultWorkers = 16
	// DefaultQueueSize is the default number of queued transactions.
	DefaultQueueSize = 64
	// DefaultTimeout is the default total budget for one pool query.
	DefaultTimeout = a2s.DefaultDeadlineTimeout
	// DefaultBufferSize is the default UDP receive buffer size.
	DefaultBufferSize uint16 = transport.DefaultBufferSize
)

// Option configures a Pool during construction.
type Option func(*config) error

// config contains resolved construction options.
type config struct {
	workers    int
	queueSize  int
	timeout    time.Duration
	bufferSize uint16
}

// WithWorkers sets the maximum number of simultaneous UDP transactions.
func WithWorkers(workers int) Option {
	return func(config *config) error {
		if workers < 1 {
			return fmt.Errorf("%w: workers must be positive", ErrInvalidConfig)
		}

		config.workers = workers
		return nil
	}
}

// WithQueueSize sets the maximum number of pending transactions.
func WithQueueSize(size int) Option {
	return func(config *config) error {
		if size < 1 {
			return fmt.Errorf("%w: queue size must be positive", ErrInvalidConfig)
		}

		config.queueSize = size
		return nil
	}
}

// WithTimeout sets the total query budget beginning at enqueue time.
// Zero disables the pool timeout; the caller context still applies.
func WithTimeout(timeout time.Duration) Option {
	return func(config *config) error {
		if timeout < 0 {
			return fmt.Errorf("%w: timeout must not be negative", ErrInvalidConfig)
		}

		config.timeout = timeout
		return nil
	}
}

// WithBufferSize sets the maximum UDP datagram size read by each worker.
func WithBufferSize(size uint16) Option {
	return func(config *config) error {
		if size == 0 {
			return fmt.Errorf("%w: buffer size must be positive", ErrInvalidConfig)
		}

		config.bufferSize = size
		return nil
	}
}
