package a2s

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

// AppendPlayers appends one complete logical A2S_PLAYER response to dst.
func AppendPlayers(dst []byte, players []Player) ([]byte, error) {
	if len(players) > 255 {
		return dst, fmt.Errorf("%w: %d players exceed uint8 count", ErrPlayerEncode, len(players))
	}
	if err := validatePlayersForEncoding(players); err != nil {
		return dst, err
	}

	dst = appendPacketHeader(dst, ResponsePlayers)
	dst = append(dst, byte(len(players))) // #nosec G115 -- count is validated to fit uint8.
	for _, player := range players {
		dst = append(dst, player.Index)
		dst = append(dst, player.Name...)
		dst = append(dst, 0)
		dst = binary.LittleEndian.AppendUint32(dst, uint32(player.Score)) // #nosec G115 -- preserve signed score bits.

		seconds, err := durationToSeconds32(player.Duration)
		if err != nil {
			return dst, err
		}
		dst = binary.LittleEndian.AppendUint32(dst, math.Float32bits(seconds))
	}

	return dst, nil
}

// validatePlayersForEncoding validates fields that have no lossless wire form.
func validatePlayersForEncoding(players []Player) error {
	for _, player := range players {
		if strings.IndexByte(player.Name, 0) >= 0 {
			return fmt.Errorf("%w: player name contains NUL", ErrPlayerEncode)
		}
		if player.Duration < 0 {
			return fmt.Errorf("%w: player duration cannot be negative", ErrPlayerEncode)
		}
		if _, err := durationToSeconds32(player.Duration); err != nil {
			return err
		}
	}

	return nil
}

// durationToSeconds32 converts a Go duration to the float32 seconds field used by A2S_PLAYER.
func durationToSeconds32(duration time.Duration) (float32, error) {
	if duration < 0 {
		return 0, fmt.Errorf("%w: player duration cannot be negative", ErrPlayerEncode)
	}

	seconds := float32(float64(duration) / float64(time.Second))
	if math.IsNaN(float64(seconds)) || math.IsInf(float64(seconds), 0) {
		return 0, fmt.Errorf("%w: player duration is not representable", ErrPlayerEncode)
	}

	return seconds, nil
}
