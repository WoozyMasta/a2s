package a2s

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/pkg/bread"
	"github.com/woozymasta/steam/utils/appid"
)

// Read buffer for populate Info struct for Source protocol
func (i *Info) readSourceInfo(buf *bytes.Buffer) error {
	var err error

	if i.Protocol, err = bread.Byte(buf); err != nil {
		return fmt.Errorf("protocol: %w", err)
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

	id, err := bread.Uint16(buf)
	if err != nil {
		return fmt.Errorf("game ID: %w", err)
	}
	i.ID = uint64(id)

	if i.Players, err = bread.Byte(buf); err != nil {
		return fmt.Errorf("player count: %w", err)
	}

	if i.MaxPlayers, err = bread.Byte(buf); err != nil {
		return fmt.Errorf("max player count: %w", err)
	}

	if i.Bots, err = bread.Byte(buf); err != nil {
		return fmt.Errorf("bots count: %w", err)
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

	if i.VAC, err = bread.Bool(buf); err != nil {
		return fmt.Errorf("VAC status: %w", err)
	}

	if i.ID == appid.TheShip.Uint64() {
		if i.TheShip, err = readTheShipInfo(buf); err != nil {
			return fmt.Errorf("TheShip data: %w", err)
		}
	}

	if i.Version, err = bread.String(buf); err != nil {
		return fmt.Errorf("version: %w", err)
	}

	edf, err := bread.Byte(buf)
	if err != nil {
		return fmt.Errorf("extra data flag: %w", err)
	}
	if edf != 0 {
		if err := i.readEDF(buf, EDF(edf)); err != nil {
			return fmt.Errorf("EDF: %w", err)
		}
	}

	return nil
}
