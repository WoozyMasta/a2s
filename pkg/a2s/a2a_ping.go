package a2s

import (
	"context"
	"fmt"
	"time"

	"github.com/woozymasta/a2s/internal/wire"
)

// GetPing queries server ping (A2A_PING) and returns complete query latency.
// Deprecated: ping is included in all query responses.
func (c *Client) GetPing(ctx context.Context) (time.Duration, error) {
	packet, meta, err := c.Query(ctx, PingRequest)
	if err != nil {
		return 0, err
	}

	if err := parsePing(packet.Payload); err != nil {
		return meta.Duration, err
	}

	return meta.Duration, nil
}

// parsePing validates the payload of an A2A_PING response.
func parsePing(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	decoder := wire.NewDecoder(data)
	if _, err := decoder.CString(); err != nil {
		return fmt.Errorf("%w payload: %w", ErrPingRead, err)
	}

	return nil
}
