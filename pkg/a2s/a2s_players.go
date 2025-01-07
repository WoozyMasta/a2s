package a2s

import (
	"bytes"
	"fmt"
	"time"

	"github.com/woozymasta/a2s/pkg/bread"
)

// https://developer.valvesoftware.com/wiki/Server_queries#Response_Format_2
type Player struct {
	Index    byte          `json:"index,omitempty"`
	Name     string        `json:"name,omitempty"`
	Score    uint32        `json:"score,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

// Get A2S_PLAYER
func (c *Client) GetPlayers() (*[]Player, error) {
	data, _, _, err := c.Get(PlayerRequest)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)
	count, err := bread.Byte(buf)
	if err != nil {
		return nil, fmt.Errorf("%w count: %w", ErrPlayerRead, err)
	}

	players := []Player{}

	for i := 0; i < int(count); i++ {
		player := Player{}

		if player.Index, err = bread.Byte(buf); err != nil {
			return nil, fmt.Errorf("%w index: %w", ErrPlayerRead, err)
		}

		if player.Name, err = bread.String(buf); err != nil {
			return nil, fmt.Errorf("%w name: %w", ErrPlayerRead, err)
		}

		if player.Score, err = bread.Uint32(buf); err != nil {
			return nil, fmt.Errorf("%w score: %w", ErrPlayerRead, err)
		}

		if player.Duration, err = bread.Duration32(buf); err != nil {
			return nil, fmt.Errorf("%w duration: %w", ErrPlayerRead, err)
		}

		players = append(players, player)
	}

	return &players, nil
}
