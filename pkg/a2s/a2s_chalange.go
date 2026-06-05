package a2s

import (
	"context"
	"errors"

	"github.com/woozymasta/a2s/internal/wire"
)

// GetChallenge queries challenge number (A2S_SERVERQUERY_GETCHALLENGE).
//
// Deprecated: challenge is handled automatically by Get() method.
func (c *Client) GetChallenge(ctx context.Context) (uint32, error) {
	data, _, _, err := c.Get(ctx, ChallengeRequest)
	if err != nil {
		return 0, err
	}

	return parseChallenge(data)
}

// parseChallenge parses the four-byte little-endian challenge payload.
func parseChallenge(data []byte) (uint32, error) {
	decoder := wire.NewDecoder(data)
	challenge, err := decoder.Uint32()
	if err != nil {
		return 0, errors.Join(ErrChallengeValue, err)
	}

	return challenge, nil
}
