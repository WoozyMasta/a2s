package a2s

import (
	"context"
	"errors"
	"time"

	"github.com/woozymasta/a2s/internal/wire"
)

// TheShip contains additional game-specific data for The Ship game.
type TheShip struct {
	Mode      TheShipMode `json:"mode"`      // Current game mode.
	Witnesses byte        `json:"witnesses"` // Number of witnesses.
	Duration  byte        `json:"duration"`  // Remaining game duration.
}

// TheShipPlayer contains player data with additional fields for The Ship game.
type TheShipPlayer struct {
	Name     string        `json:"name,omitempty"`     // Player name.
	Duration time.Duration `json:"duration,omitempty"` // Session duration.
	Score    int32         `json:"score,omitempty"`    // Signed game score.
	Deaths   uint32        `json:"deaths,omitempty"`   // Number of deaths.
	Money    uint32        `json:"money,omitempty"`    // In-game money.
	Index    byte          `json:"index,omitempty"`    // Index in the response.
}

// readTheShipInfo parses The Ship game-specific data from A2S_INFO response.
func readTheShipInfo(r *wire.Decoder) (*TheShip, error) {
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

// GetTheShipPlayers queries the player list with The Ship game-specific fields.
// It returns a non-nil empty slice when the server reports no players.
func (c *Client) GetTheShipPlayers(ctx context.Context) ([]TheShipPlayer, error) {
	packet, _, err := c.Query(ctx, PlayerRequest)
	if err != nil {
		return nil, err
	}

	return parseTheShipPlayers(packet.Payload)
}

// parseTheShipPlayers parses The Ship's extended A2S_PLAYER payload.
func parseTheShipPlayers(data []byte) ([]TheShipPlayer, error) {
	decoder := wire.NewDecoder(data)
	count, err := decoder.Byte()
	if err != nil {
		return nil, errors.Join(ErrPlayerCount, err)
	}

	players := make([]TheShipPlayer, 0, int(count))

	for i := 0; i < int(count); i++ {
		player := TheShipPlayer{}

		if player.Index, err = decoder.Byte(); err != nil {
			return nil, errors.Join(ErrPlayerIndex, err)
		}

		if player.Name, err = decoder.CString(); err != nil {
			return nil, errors.Join(ErrPlayerName, err)
		}

		if player.Score, err = decoder.Int32(); err != nil {
			return nil, errors.Join(ErrPlayerScore, err)
		}

		var seconds float32
		if seconds, err = decoder.Float32(); err != nil {
			return nil, errors.Join(ErrPlayerDuration, err)
		}
		player.Duration = durationFromSeconds32(seconds)

		if player.Deaths, err = decoder.Uint32(); err != nil {
			return nil, errors.Join(ErrPlayerDeaths, err)
		}

		if player.Money, err = decoder.Uint32(); err != nil {
			return nil, errors.Join(ErrPlayerMoney, err)
		}

		players = append(players, player)
	}

	return players, nil
}
