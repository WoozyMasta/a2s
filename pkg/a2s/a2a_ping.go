package a2s

import (
	"bytes"
	"fmt"
	"time"

	"github.com/woozymasta/a2s/internal/bread"
)

// GetPing queries server ping (A2A_PING).
// Deprecated: ping is included in all query responses.
func (c *Client) GetPing() (time.Duration, error) {
	data, _, duration, err := c.Get(PingRequest)
	if err != nil {
		return 0, err
	}

	buf := bytes.NewBuffer(data)
	if _, err := bread.String(buf); err != nil {
		return duration, fmt.Errorf("%w payload: %w", ErrPingRead, err)
	}

	return duration, nil
}
