package a2s

import (
	"context"
	"errors"

	"github.com/woozymasta/a2s/internal/wire"
)

// errInfoInvalidBoolean identifies a malformed A2S boolean field.
var errInfoInvalidBoolean = errors.New("A2S_INFO: boolean field must be 0 or 1")

// Info contains parsed A2S_INFO response data.
//
// See https://developer.valvesoftware.com/wiki/Server_queries#Response_Format
type Info struct {
	// The Ship-specific fields, present only for The Ship servers.
	TheShip *TheShip `json:"the_ship,omitempty"`

	// GoldSource mod information, present when the Mod field is 0x01.
	Mod *ModInfo `json:"mod,omitempty"`

	// GameID is the optional full identifier (EDF 0x01).
	// A nil pointer means that the EDF GameID field was not present.
	GameID *uint64 `json:"game_id,omitempty"`

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

	// Server tags (EDF 0x20).
	Keywords []string `json:"keywords,omitempty"`

	// Server SteamID (EDF 0x10).
	SteamID uint64 `json:"steam_id,omitempty"`

	// AppID is the 16-bit identifier from the base A2S_INFO response.
	AppID uint16 `json:"app_id"`

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

// EffectiveID returns the full GameID when present, otherwise the legacy AppID.
func (i Info) EffectiveID() uint64 {
	if i.GameID != nil {
		return *i.GameID
	}

	return uint64(i.AppID)
}

// GetInfo queries server information (A2S_INFO).
func (c *Client) GetInfo(ctx context.Context) (*Info, error) {
	info, _, err := c.GetInfoWithMeta(ctx)
	return info, err
}

// GetInfoWithMeta queries server information
// and returns transport metadata separately
// from the protocol response model.
func (c *Client) GetInfoWithMeta(ctx context.Context) (*Info, QueryMeta, error) {
	data, format, duration, err := c.Get(ctx, InfoRequest)
	if err != nil {
		return nil, QueryMeta{}, err
	}

	info, err := parseInfo(data, format)
	if err != nil {
		return nil, QueryMeta{}, err
	}

	return info, QueryMeta{Duration: duration}, nil
}

// parseInfo parses an A2S_INFO payload without taking ownership of its buffer.
func parseInfo(data []byte, format ResponseType) (*Info, error) {
	decoder := wire.NewDecoder(data)
	info := &Info{Format: InfoFormat(format)}

	switch format {
	case ResponseInfo:
		if err := info.readSourceInfo(&decoder); err != nil {
			return nil, errors.Join(ErrInfoSourceResponse, err)
		}

	case ResponseInfoGoldSource:
		if err := info.readGoldSourceInfo(&decoder); err != nil {
			return nil, errors.Join(ErrInfoGoldSourceResponse, err)
		}

	default:
		return nil, ErrInfoUnsupportedFormat
	}

	return info, nil
}

// readInfoBool decodes the strict boolean representation used by A2S_INFO.
func readInfoBool(decoder *wire.Decoder) (bool, error) {
	value, err := decoder.Byte()
	if err != nil {
		return false, err
	}

	switch value {
	case 0:
		return false, nil

	case 1:
		return true, nil

	default:
		return false, errInfoInvalidBoolean
	}
}
