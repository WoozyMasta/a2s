// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"context"

	"github.com/woozymasta/a2s/pkg/a2s/transport"
)

// GetChallenge queries an opaque challenge token
// (A2S_SERVERQUERY_GETCHALLENGE).
//
// The A2S_SERVERQUERY_GETCHALLENGE wire request is obsolete for ordinary queries
// because Query handles challenge exchanges automatically,
// but this method remains supported for explicit protocol access.
func (c *Client) GetChallenge(ctx context.Context) (transport.Challenge, error) {
	packet, _, err := c.Query(ctx, transport.ChallengeRequest)
	if err != nil {
		return transport.Challenge{}, err
	}

	return transport.ParseChallengePayload(packet.Payload)
}
