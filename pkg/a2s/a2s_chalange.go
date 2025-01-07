package a2s

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/pkg/bread"
)

// Get A2S_SERVERQUERY_GETCHALLENGE (Deprecated)
func (c *Client) GetChallenge() (uint32, error) {
	data, _, _, err := c.Get(ChallengeRequest)
	if err != nil {
		return 0, err
	}

	buf := bytes.NewBuffer(data)
	challenge, err := bread.Uint32(buf)
	if err != nil {
		return 0, fmt.Errorf("%w challenge: %w", ErrChallengeRead, err)
	}

	return challenge, nil
}
