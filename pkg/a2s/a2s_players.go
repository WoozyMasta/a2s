// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/woozymasta/a2s/internal/wire"
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
	packet, _, err := c.Query(ctx, PlayerRequest)
	if err != nil {
		return nil, err
	}

	return DecodePlayers(packet)
}

// DecodePlayers parses a standard logical A2S_PLAYER response packet.
//
// The response type must be ResponsePlayers.
// The Ship's extended player payload remains available through GetTheShipPlayers.
func DecodePlayers(packet Packet) ([]Player, error) {
	if packet.Type != ResponsePlayers {
		return nil, errors.Join(ErrPlayerRead, fmt.Errorf("unexpected response type 0x%X", packet.Type))
	}

	decoder := wire.NewDecoder(packet.Payload)
	count, err := decoder.Byte()
	if err != nil {
		return nil, errors.Join(ErrPlayerCount, err)
	}

	players := make([]Player, 0, int(count))

	for i := 0; i < int(count); i++ {
		player := Player{}

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

		players = append(players, player)
	}

	return players, nil
}

// durationFromSeconds32 converts the float32 seconds
// used by A2S_PLAYER to a time.Duration
// while preserving the protocol's fractional-second rounding.
func durationFromSeconds32(seconds float32) time.Duration {
	return time.Duration(math.Round(float64(seconds) * float64(time.Second)))
}
