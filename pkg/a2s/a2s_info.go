package a2s

import (
	"bytes"
	"fmt"
	"time"
)

// Structure for store A2S_INFO response data
// https://developer.valvesoftware.com/wiki/Server_queries#Response_Format
type Info struct {
	TheShip      *TheShip      `json:"the_ship,omitempty"`       // These fields only exist if server is The Ship
	Mod          *ModInfo      `json:"mod,omitempty"`            // Mod info, present if field Mod is 0x01 [Additional for GoldSource]
	Name         string        `json:"name"`                     // Name of the server
	Map          string        `json:"map"`                      // Map the server has currently loaded
	Folder       string        `json:"folder"`                   // Name of the folder containing the game files
	Game         string        `json:"game,omitempty"`           // Full name of the game
	Version      string        `json:"version"`                  // Version of the game installed on the server
	SourceTVName string        `json:"source_tv_port,omitempty"` // Name of the spectator server for SourceTV (EDF 0x40)
	Address      string        `json:"address,omitempty"`        // IP address and port of the server. [Additional for GoldSource]
	Keywords     []string      `json:"keywords,omitempty"`       // Tags that describe the game according to the server (EDF 0x20)
	Ping         time.Duration `json:"ping"`                     // Server response time (custom)
	ID           uint64        `json:"id"`                       // Steam Application ID of game (Reuse EDF 0x01)
	SteamID      uint64        `json:"steam_id,omitempty"`       // Server SteamID (EDF 0x10)
	Port         uint16        `json:"port,omitempty"`           // Game port number (EDF 0x80)
	SourceTVPort uint16        `json:"source_tv_name,omitempty"` // Spectator port number for SourceTV (EDF 0x40 )
	Format       InfoFormat    `json:"format"`                   // Response format (Source or obsolete GoldSource)
	Protocol     byte          `json:"protocol"`                 // Protocol version used by the server
	Players      byte          `json:"players"`                  // Number of players on the server
	MaxPlayers   byte          `json:"max_players"`              // Maximum number of players the server reports it can hold
	Bots         byte          `json:"bots,omitempty"`           // Number of bots on the server
	ServerType   ServerType    `json:"type"`                     // Indicates the type of server
	Environment  Environment   `json:"environment"`              // Indicates the operating system of the server
	Visibility   bool          `json:"public"`                   // Indicates whether the server requires a password
	VAC          bool          `json:"vac"`                      // Specifies whether the server uses VAC
	EDF          EDF           `json:"EDF,omitempty"`            // If present, specifies additional data fields

	// GameID       uint64        `json:"game_id,omitempty"`        // GameID, already set in ID (EDF 0x01)
}

// Get A2S_INFO
func (c *Client) GetInfo() (*Info, error) {
	data, format, duration, err := c.Get(InfoRequest)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)
	info := &Info{Ping: duration, Format: InfoFormat(format)}

	switch format {
	case infoResponseSource:
		if err := info.readSourceInfo(buf); err != nil {
			return nil, fmt.Errorf("%w Source response: %w", ErrInfoRead, err)
		}

	case infoResponseGoldSource:
		if err := info.readGoldSourceInfo(buf); err != nil {
			return nil, fmt.Errorf("%w GoldSource response: %w", ErrInfoRead, err)
		}

	default:
		return nil, fmt.Errorf("%w header: unsupported format 0x%X", ErrInfoRead, format)
	}

	return info, nil
}
