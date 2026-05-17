package a2s

import (
	"context"
	"errors"
	"time"

	"github.com/woozymasta/a2s/internal/bread"
)

// Info contains parsed A2S_INFO response data.
//
// A2S_INFO has a legacy 16-bit AppID and an optional EDF GameID.
// This type intentionally exposes one effective ID:
// EDF GameID takes precedence because the legacy value may be truncated
// Callers use ID for game detection and output;
// the raw wire identifiers are not exposed separately.
//
// See https://developer.valvesoftware.com/wiki/Server_queries#Response_Format
type Info struct {
	// The Ship-specific fields, present only for The Ship servers.
	TheShip *TheShip `json:"the_ship,omitempty"`

	// GoldSource mod information, present when the Mod field is 0x01.
	Mod *ModInfo `json:"mod,omitempty"`

	// Server name.
	Name string `json:"name"`

	// Map currently loaded by the server.
	Map string `json:"map"`

	// Folder containing the game files.
	Folder string `json:"folder"`

	// Full game name.
	Game string `json:"game,omitempty"`

	// Installed game version.
	Version string `json:"version"`

	// SourceTV spectator server name (EDF 0x40).
	SourceTVName string `json:"source_tv_name,omitempty"`

	// Server IP address and port from the GoldSource response.
	Address string `json:"address,omitempty"`

	// Server tags from EDF 0x20.
	Keywords []string `json:"keywords,omitempty"`

	// Complete query round-trip time;
	// this field is not sent by the server.
	Ping time.Duration `json:"ping"`

	// Effective game identifier;
	// EDF GameID replaces the legacy AppID when present.
	ID uint64 `json:"id"`

	// Server SteamID (EDF 0x10).
	SteamID uint64 `json:"steam_id,omitempty"`

	// Game port number (EDF 0x80).
	Port uint16 `json:"port,omitempty"`

	// SourceTV spectator port (EDF 0x40).
	SourceTVPort uint16 `json:"source_tv_port,omitempty"`

	// Response format: Source or obsolete GoldSource.
	Format InfoFormat `json:"format"`

	// Protocol version used by the server.
	Protocol byte `json:"protocol"`

	// Current player count.
	Players byte `json:"players"`

	// Maximum player count reported by the server.
	MaxPlayers byte `json:"max_players"`

	// Current bot count.
	Bots byte `json:"bots,omitempty"`

	// Server type.
	ServerType ServerType `json:"server_type"`

	// Server operating system.
	Environment Environment `json:"environment"`

	// Whether the server requires a password.
	Visibility bool `json:"visibility"`

	// Whether Valve Anti-Cheat is enabled.
	VAC bool `json:"vac"`

	// Extra Data Flags indicating which optional fields are present.
	EDF EDF `json:"EDF,omitempty"`
}

// GetInfo queries server information (A2S_INFO).
func (c *Client) GetInfo(ctx context.Context) (*Info, error) {
	data, format, duration, err := c.Get(ctx, InfoRequest)
	if err != nil {
		return nil, err
	}

	return parseInfo(data, format, duration)
}

// parseInfo parses an A2S_INFO payload without taking ownership of its buffer.
func parseInfo(data []byte, format Flag, duration time.Duration) (*Info, error) {
	reader := bread.NewReader(data)
	info := &Info{Ping: duration, Format: InfoFormat(format)}

	switch format {
	case infoResponseSource:
		if err := info.readSourceInfo(reader); err != nil {
			return nil, errors.Join(ErrInfoSourceResponse, err)
		}

	case infoResponseGoldSource:
		if err := info.readGoldSourceInfo(reader); err != nil {
			return nil, errors.Join(ErrInfoGoldSourceResponse, err)
		}

	default:
		return nil, ErrInfoUnsupportedFormat
	}

	return info, nil
}
