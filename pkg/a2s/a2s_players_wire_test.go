package a2s

import (
	"encoding/binary"
	"errors"
	"math"
	"testing"
	"time"
)

func TestDurationFromSeconds32(t *testing.T) {
	for _, test := range []struct {
		name  string
		value float32
		want  time.Duration
	}{
		{name: "zero", value: 0, want: 0},
		{name: "fractional", value: 1.25, want: 1250 * time.Millisecond},
		{name: "positive", value: 2, want: 2 * time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := durationFromSeconds32(test.value); got != test.want {
				t.Fatalf("durationFromSeconds32(%v) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}

func TestDurationFromSeconds32PreservesLegacyRounding(t *testing.T) {
	values := []float32{
		-1000000,
		-123.45679,
		-1.0000001,
		-1.25,
		-0.0000001,
		0.0000001,
		0.5,
		1.0000001,
		1.25,
		123.45679,
		1000000,
	}

	for _, value := range values {
		want := legacyDurationFromSeconds32(value)
		if got := durationFromSeconds32(value); got != want {
			t.Errorf("durationFromSeconds32(%v) = %v, want %v", value, got, want)
		}
	}

	state := uint32(0x12345678)
	for i := 0; i < 100_000; i++ {
		state = state*1664525 + 1013904223
		value := math.Float32frombits(state)
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) || math.Abs(float64(value)) > 1_000_000 {
			continue
		}

		want := legacyDurationFromSeconds32(value)
		if got := durationFromSeconds32(value); got != want {
			t.Fatalf("durationFromSeconds32(%v) = %v, want %v", value, got, want)
		}
	}
}

func legacyDurationFromSeconds32(seconds float32) time.Duration {
	wholeSeconds := int64(seconds)
	nanoseconds := int64(math.Round(float64(seconds-float32(wholeSeconds)) * 1e9))

	return time.Duration(wholeSeconds)*time.Second + time.Duration(nanoseconds)*time.Nanosecond
}

func TestParsePlayersTruncation(t *testing.T) {
	data := playerFixture(math.Float32bits(1.25))

	tests := []struct {
		name string
		end  int
		want error
	}{
		{name: "count", end: 0, want: ErrPlayerCount},
		{name: "index", end: 1, want: ErrPlayerIndex},
		{name: "name", end: 2, want: ErrPlayerName},
		{name: "score", end: 9, want: ErrPlayerScore},
		{name: "duration", end: 13, want: ErrPlayerDuration},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parsePlayers(data[:test.end])
			if !errors.Is(err, test.want) {
				t.Fatalf("parsePlayers() error = %v, want %v", err, test.want)
			}
		})
	}

	for end := 0; end < len(data); end++ {
		t.Run("arbitrary truncation", func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("parsePlayers() panicked for length %d: %v", end, recovered)
				}
			}()
			_, _ = parsePlayers(data[:end])
		})
	}
}

func TestParseTheShipPlayersTruncation(t *testing.T) {
	data := theShipPlayerFixture()

	tests := []struct {
		name string
		end  int
		want error
	}{
		{name: "count", end: 0, want: ErrPlayerCount},
		{name: "index", end: 1, want: ErrPlayerIndex},
		{name: "name", end: 2, want: ErrPlayerName},
		{name: "score", end: 9, want: ErrPlayerScore},
		{name: "duration", end: 13, want: ErrPlayerDuration},
		{name: "deaths", end: 17, want: ErrPlayerDeaths},
		{name: "money", end: 21, want: ErrPlayerMoney},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseTheShipPlayers(data[:test.end])
			if !errors.Is(err, test.want) {
				t.Fatalf("parseTheShipPlayers() error = %v, want %v", err, test.want)
			}
		})
	}

	for end := 0; end < len(data); end++ {
		t.Run("arbitrary truncation", func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Fatalf("parseTheShipPlayers() panicked for length %d: %v", end, recovered)
				}
			}()
			_, _ = parseTheShipPlayers(data[:end])
		})
	}
}

func playerFixture(durationBits uint32) []byte {
	data := []byte{1, 0}
	data = append(data, "player"...)
	data = append(data, 0)
	score := int32(-1)
	data = binary.LittleEndian.AppendUint32(data, uint32(score))
	data = binary.LittleEndian.AppendUint32(data, durationBits)
	return data
}

func theShipPlayerFixture() []byte {
	data := playerFixture(math.Float32bits(1.25))
	data = binary.LittleEndian.AppendUint32(data, 2)
	data = binary.LittleEndian.AppendUint32(data, 3)
	return data
}
