package a2s

import (
	"context"
	"fmt"
	"time"

	"github.com/woozymasta/a2s/internal/bread"
)

// GetPing queries server ping (A2A_PING).
// Deprecated: ping is included in all query responses.
func (c *Client) GetPing(ctx context.Context) (time.Duration, error) {
	data, _, duration, err := c.Get(ctx, PingRequest)
	if err != nil {
		return 0, err
	}

	reader := bread.NewReader(data)
	if _, err := reader.String(); err != nil {
		return duration, fmt.Errorf("%w payload: %w", ErrPingRead, err)
	}

	return duration, nil
}
