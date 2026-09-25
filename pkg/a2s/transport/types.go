// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

import (
	"context"
	"time"
)

// QueryType identifies an A2S request type byte.
type QueryType byte

// ResponseType is a raw A2S response type byte.
type ResponseType byte

// Challenge is the opaque four-byte challenge token used by A2S.
type Challenge [4]byte

// Packet is a decoded logical A2S response.
type Packet struct {
	// Payload contains the logical response payload without packet framing.
	Payload []byte
	// Type is the response discriminator byte.
	Type ResponseType
}

// QueryMeta contains transport metadata for one completed logical query.
type QueryMeta struct {
	// Duration covers the complete logical transaction.
	Duration time.Duration
	// UsedChallenge reports whether at least one challenge exchange occurred.
	UsedChallenge bool
}

// DatagramIO is the socket operation required by the transaction engine.
// Implementations must honor ctx while reading and writing.
type DatagramIO interface {
	Write(context.Context, []byte) error
	Read(context.Context, []byte) (int, error)
}

// initialChallenge is the reserved wire token used for the first query attempt.
var initialChallenge = Challenge{0xFF, 0xFF, 0xFF, 0xFF}

// InitialChallenge returns the reserved token used to start a challenge exchange.
// The returned array is a copy and may be modified by the caller.
func InitialChallenge() Challenge {
	return initialChallenge
}
