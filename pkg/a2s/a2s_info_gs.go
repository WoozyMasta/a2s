package a2s

import (
	"errors"

	"github.com/woozymasta/a2s/internal/bread"
)

// ModInfo contains mod information from GoldSource A2S_INFO response.
type ModInfo struct {
	Link         string `json:"link"`          // URL to mod website.
	DownloadLink string `json:"download_link"` // URL to download the mod.
	Version      uint32 `json:"version"`       // Version of mod installed on server.
	Size         uint32 `json:"size"`          // Space (in bytes) the mod takes up.
	Type         bool   `json:"type"`          // Indicates the type of mod: false - single+multi, true - multiplayer only
	DLL          bool   `json:"dll"`           // false - original DLL, true - own DLL
}

// readGoldSourceInfo parses GoldSource protocol A2S_INFO response (obsolete).
func (i *Info) readGoldSourceInfo(r *bread.Reader) error {
	var err error

	if i.Address, err = r.String(); err != nil {
		return errors.Join(ErrInfoGSAddress, err)
	}

	if i.Name, err = r.String(); err != nil {
		return errors.Join(ErrInfoServerName, err)
	}

	if i.Map, err = r.String(); err != nil {
		return errors.Join(ErrInfoMapName, err)
	}

	if i.Folder, err = r.String(); err != nil {
		return errors.Join(ErrInfoFolderName, err)
	}

	if i.Game, err = r.String(); err != nil {
		return errors.Join(ErrInfoGameName, err)
	}

	if i.Players, err = r.Byte(); err != nil {
		return errors.Join(ErrInfoPlayerCount, err)
	}

	if i.MaxPlayers, err = r.Byte(); err != nil {
		return errors.Join(ErrInfoMaxPlayerCount, err)
	}

	if i.Protocol, err = r.Byte(); err != nil {
		return errors.Join(ErrInfoProtocol, err)
	}

	serverType, err := r.Byte()
	if err != nil {
		return errors.Join(ErrInfoServerType, err)
	}
	i.ServerType = ServerType(serverType)

	environment, err := r.Byte()
	if err != nil {
		return errors.Join(ErrInfoEnvironment, err)
	}
	i.Environment = Environment(environment)

	if i.Visibility, err = r.Bool(); err != nil {
		return errors.Join(ErrInfoVisibility, err)
	}

	modded, err := r.Bool()
	if err != nil {
		return errors.Join(ErrInfoGSModded, err)
	}

	if modded {
		if i.Mod, err = readGoldSourceMods(r); err != nil {
			return errors.Join(ErrInfoGSModData, err)
		}
	}

	if i.VAC, err = r.Bool(); err != nil {
		return errors.Join(ErrInfoVAC, err)
	}

	if i.Bots, err = r.Byte(); err != nil {
		return errors.Join(ErrInfoBotsCount, err)
	}

	return nil
}
