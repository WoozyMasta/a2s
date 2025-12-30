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

	if cap(c.parseData) < len(data) {
		c.parseData = make([]byte, len(data)+64)
	}
	c.parseData = c.parseData[:len(data)]
	copy(c.parseData, data)

	reader := bread.NewReader(c.parseData)
	challenge, err := reader.Uint32()
	if err != nil {
		return 0, errors.Join(ErrChallengeValue, err)
	}

	return challenge, nil
}
