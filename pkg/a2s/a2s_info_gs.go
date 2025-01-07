package a2s

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/pkg/bread"
)

type ModInfo struct {
	Link         string `json:"link"`          // URL to mod website.
	DownloadLink string `json:"download_link"` // URL to download the mod.
	Version      uint32 `json:"version"`       // Version of mod installed on server.
	Size         uint32 `json:"size"`          // Space (in bytes) the mod takes up.
	Type         bool   `json:"type"`          // Indicates the type of mod: false - single+multi, true - multiplayer only
	DLL          bool   `json:"dll"`           // false - original DLL, true - own DLL
}

// Read buffer for populate Info struct for GoldSource protocol (Obsolete)
func (i *Info) readGoldSourceInfo(buf *bytes.Buffer) error {
	var err error

	if i.Address, err = bread.String(buf); err != nil {
		return fmt.Errorf("server address: %w", err)
	}

	if i.Name, err = bread.String(buf); err != nil {
		return fmt.Errorf("server name: %w", err)
	}

	if i.Map, err = bread.String(buf); err != nil {
		return fmt.Errorf("map name: %w", err)
	}

	if i.Folder, err = bread.String(buf); err != nil {
		return fmt.Errorf("folder name: %w", err)
	}

	if i.Game, err = bread.String(buf); err != nil {
		return fmt.Errorf("game name: %w", err)
	}

	if i.Players, err = bread.Byte(buf); err != nil {
		return fmt.Errorf("player count: %w", err)
	}

	if i.MaxPlayers, err = bread.Byte(buf); err != nil {
		return fmt.Errorf("max player count: %w", err)
	}

	if i.Protocol, err = bread.Byte(buf); err != nil {
		return fmt.Errorf("protocol: %w", err)
	}

	serverType, err := bread.Byte(buf)
	if err != nil {
		return fmt.Errorf("server type: %w", err)
	}
	i.ServerType = ServerType(serverType)

	environment, err := bread.Byte(buf)
	if err != nil {
		return fmt.Errorf("environment type: %w", err)
	}
	i.Environment = Environment(environment)

	if i.Visibility, err = bread.Bool(buf); err != nil {
		return fmt.Errorf("server visibility: %w", err)
	}

	modded, err := bread.Bool(buf)
	if err != nil {
		return fmt.Errorf("modded status: %w", err)
	}

	if modded {
		if i.Mod, err = readGoldSourceMods(buf); err != nil {
			return fmt.Errorf("mod data: %w", err)
		}
	}

	if i.VAC, err = bread.Bool(buf); err != nil {
		return fmt.Errorf("VAC: %w", err)
	}

	if i.Bots, err = bread.Byte(buf); err != nil {
		return fmt.Errorf("bots count: %w", err)
	}

	return nil
}
