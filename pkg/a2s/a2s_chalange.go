package a2s

import (
	"errors"

	"github.com/woozymasta/a2s/internal/bread"
)

// GetChallenge queries challenge number (A2S_SERVERQUERY_GETCHALLENGE).
// Deprecated: challenge is handled automatically by Get() method.
func (c *Client) GetChallenge() (uint32, error) {
	data, _, _, err := c.Get(ChallengeRequest)
	if err != nil {
		return 0, err
	}

	return parseChallenge(data)
}

// parseChallenge parses the four-byte little-endian challenge payload.
func parseChallenge(data []byte) (uint32, error) {
	reader := bread.NewReader(data)
	challenge, err := reader.Uint32()
	if err != nil {
		return 0, errors.Join(ErrChallengeValue, err)
	}

	return challenge, nil
}
