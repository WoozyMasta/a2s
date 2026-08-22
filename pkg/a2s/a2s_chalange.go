// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"context"
	"errors"

	"github.com/woozymasta/a2s/internal/wire"
)

// GetChallenge queries an opaque challenge token
// (A2S_SERVERQUERY_GETCHALLENGE).
//
// The A2S_SERVERQUERY_GETCHALLENGE wire request is obsolete for ordinary queries
// because Query handles challenge exchanges automatically,
// but this method remains supported for explicit protocol access.
func (c *Client) GetChallenge(ctx context.Context) (Challenge, error) {
	packet, _, err := c.Query(ctx, ChallengeRequest)
	if err != nil {
		return Challenge{}, err
	}

	return parseChallenge(packet.Payload)
}

// parseChallenge parses the four-byte challenge payload without reordering it.
func parseChallenge(data []byte) (Challenge, error) {
	decoder := wire.NewDecoder(data)
	value, err := decoder.Bytes(len(Challenge{}))
	if err != nil {
		return Challenge{}, errors.Join(ErrChallengeValue, err)
	}

	var challenge Challenge
	copy(challenge[:], value)

	return challenge, nil
}
