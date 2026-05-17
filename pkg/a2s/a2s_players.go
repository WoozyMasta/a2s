package a2s

import (
	"context"
	"errors"
	"time"

	"github.com/woozymasta/a2s/internal/bread"
)

// Player contains player information from A2S_PLAYER query.
//
// See https://developer.valvesoftware.com/wiki/Server_queries#Response_Format_2
type Player struct {
	Name     string        `json:"name,omitempty"`     // Player name.
	Duration time.Duration `json:"duration,omitempty"` // Session duration.
	Score    int32         `json:"score,omitempty"`    // Signed game score.
	Index    byte          `json:"index,omitempty"`    // Index in the response.
}

// GetPlayers queries the player list (A2S_PLAYER).
// It returns a non-nil empty slice when the server reports no players.
func (c *Client) GetPlayers(ctx context.Context) ([]Player, error) {
	data, _, _, err := c.Get(ctx, PlayerRequest)
	if err != nil {
		return nil, err
	}

	return parsePlayers(data)
}

// parsePlayers parses a standard A2S_PLAYER payload without copying its buffer.
func parsePlayers(data []byte) ([]Player, error) {
	reader := bread.NewReader(data)
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

		if player.Score, err = reader.Int32(); err != nil {
			return nil, errors.Join(ErrPlayerScore, err)
		}

		if player.Duration, err = reader.Duration32(); err != nil {
			return nil, errors.Join(ErrPlayerDuration, err)
		}

		players = append(players, player)
	}

	return players, nil
}
