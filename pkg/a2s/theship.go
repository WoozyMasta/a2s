package a2s

import (
	"bytes"
	"fmt"
	"time"

	"github.com/woozymasta/a2s/pkg/bread"
)

type TheShip struct {
	Mode      TheShipMode `json:"mode"`
	Witnesses byte        `json:"witnesses"`
	Duration  byte        `json:"duration"`
}

type TheShipPlayer struct {
	Name     string        `json:"name,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
	Score    uint32        `json:"score,omitempty"`
	Deaths   uint32        `json:"deaths,omitempty"`
	Money    uint32        `json:"money,omitempty"`
	Index    byte          `json:"index,omitempty"`
}

// Read extra info for TheShip game for A2S_INFO
func readTheShipInfo(buf *bytes.Buffer) (*TheShip, error) {
	theShip := &TheShip{}

	mode, err := bread.Byte(buf)
	if err != nil {
		return nil, err
	}
	theShip.Mode = TheShipMode(mode)

	if theShip.Witnesses, err = bread.Byte(buf); err != nil {
		return nil, err
	}

	if theShip.Duration, err = bread.Byte(buf); err != nil {
		return nil, err
	}

	return theShip, nil
}

// Get A2S_PLAYER for TheShip game
func (c *Client) GetTheShipPlayers() (*[]TheShipPlayer, error) {
	data, _, _, err := c.Get(PlayerRequest)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)
	count, err := bread.Byte(buf)
	if err != nil {
		return nil, fmt.Errorf("%w count: %w", ErrPlayerRead, err)
	}

	players := []TheShipPlayer{}

	for i := 0; i < int(count); i++ {
		player := TheShipPlayer{}

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

		if player.Deaths, err = bread.Uint32(buf); err != nil {
			return nil, fmt.Errorf("%w deaths: %w", ErrPlayerRead, err)
		}

		if player.Money, err = bread.Uint32(buf); err != nil {
			return nil, fmt.Errorf("%w money: %w", ErrPlayerRead, err)
		}

		players = append(players, player)
	}

	return &players, nil
}
