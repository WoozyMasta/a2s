package a2s

import (
	"context"
	"errors"

	"github.com/woozymasta/a2s/internal/wire"
)

// GetChallenge queries an opaque challenge token
// (A2S_SERVERQUERY_GETCHALLENGE).
//
// Deprecated: challenge is handled automatically by Query.
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
