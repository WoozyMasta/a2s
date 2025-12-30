package a2s

import (
	"errors"
	"time"

	"github.com/woozymasta/a2s/internal/bread"
)

// TheShip contains additional game-specific data for The Ship game.
type TheShip struct {
	Mode      TheShipMode `json:"mode"`
	Witnesses byte        `json:"witnesses"`
	Duration  byte        `json:"duration"`
}

// TheShipPlayer contains player data with additional fields for The Ship game.
type TheShipPlayer struct {
	Name     string        `json:"name,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
	Score    uint32        `json:"score,omitempty"`
	Deaths   uint32        `json:"deaths,omitempty"`
	Money    uint32        `json:"money,omitempty"`
	Index    byte          `json:"index,omitempty"`
}

// readTheShipInfo parses The Ship game-specific data from A2S_INFO response.
func readTheShipInfo(r *bread.Reader) (*TheShip, error) {
	theShip := &TheShip{}

	mode, err := r.Byte()
	if err != nil {
		return nil, err
	}
	theShip.Mode = TheShipMode(mode)

	if theShip.Witnesses, err = r.Byte(); err != nil {
		return nil, err
	}

	if theShip.Duration, err = r.Byte(); err != nil {
		return nil, err
	}

	return theShip, nil
}

// GetTheShipPlayers queries player list with The Ship game-specific fields (A2S_PLAYER).
func (c *Client) GetTheShipPlayers() (*[]TheShipPlayer, error) {
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

	players := make([]TheShipPlayer, 0, int(count))

	for i := 0; i < int(count); i++ {
		player := TheShipPlayer{}

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

		if player.Deaths, err = reader.Uint32(); err != nil {
			return nil, errors.Join(ErrPlayerDeaths, err)
		}

		if player.Money, err = reader.Uint32(); err != nil {
			return nil, errors.Join(ErrPlayerMoney, err)
		}

		players = append(players, player)
	}

	return &players, nil
}
