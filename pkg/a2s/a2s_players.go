package a2s

import (
	"errors"
	"time"

	"github.com/woozymasta/a2s/internal/bread"
)

// Player contains player information from A2S_PLAYER query.
// See https://developer.valvesoftware.com/wiki/Server_queries#Response_Format_2
type Player struct {
	Name     string        `json:"name,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
	Score    uint32        `json:"score,omitempty"`
	Index    byte          `json:"index,omitempty"`
}

// GetPlayers queries player list (A2S_PLAYER).
func (c *Client) GetPlayers() (*[]Player, error) {
	data, _, _, err := c.Get(PlayerRequest)
	if err != nil {
		return nil, err
	}

	if cap(c.parseData) < len(data) {
		c.parseData = make([]byte, len(data)+64)
	}
	c.parseData = c.parseData[:len(data)]
	copy(c.parseData, data)

	reader := bread.NewReader(c.parseData)
	count, err := reader.Byte()
	if err != nil {
		return nil, errors.Join(ErrPlayerCount, err)
	}

	players := make([]Player, 0, int(count))

	for i := 0; i < int(count); i++ {
		player := Player{}

		if player.Index, err = reader.Byte(); err != nil {
			return nil, errors.Join(ErrPlayerIndex, err)
		}

		if player.Name, err = reader.String(); err != nil {
			return nil, errors.Join(ErrPlayerName, err)
		}

		if player.Score, err = reader.Uint32(); err != nil {
			return nil, errors.Join(ErrPlayerScore, err)
		}

		if player.Duration, err = reader.Duration32(); err != nil {
			return nil, errors.Join(ErrPlayerDuration, err)
		}

		players = append(players, player)
	}

	return &players, nil
}
