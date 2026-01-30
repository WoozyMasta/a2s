package a2s

import (
	"errors"

	"github.com/woozymasta/a2s/internal/bread"
	"github.com/woozymasta/steam/utils/appid"
)

// readSourceInfo parses Source protocol A2S_INFO response.
func (i *Info) readSourceInfo(r *bread.Reader) error {
	var err error

	if i.Protocol, err = r.Byte(); err != nil {
		return errors.Join(ErrInfoProtocol, err)
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

	id, err := r.Uint16()
	if err != nil {
		return errors.Join(ErrInfoGameID, err)
	}
	i.ID = uint64(id)

	if i.Players, err = r.Byte(); err != nil {
		return errors.Join(ErrInfoPlayerCount, err)
	}

	if i.MaxPlayers, err = r.Byte(); err != nil {
		return errors.Join(ErrInfoMaxPlayerCount, err)
	}

	if i.Bots, err = r.Byte(); err != nil {
		return errors.Join(ErrInfoBotsCount, err)
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

	if i.VAC, err = r.Bool(); err != nil {
		return errors.Join(ErrInfoVAC, err)
	}

	if i.ID == appid.TheShip.Uint64() {
		if i.TheShip, err = readTheShipInfo(r); err != nil {
			return errors.Join(ErrInfoTheShip, err)
		}
	}

	if i.Version, err = r.String(); err != nil {
		return errors.Join(ErrInfoVersion, err)
	}

	edf, err := r.Byte()
	if err != nil {
		if errors.Is(err, bread.ErrUnderflow) {
			return nil
		}
		return errors.Join(ErrInfoEDF, err)
	}
	if edf != 0 {
		if err := i.readEDF(r, EDF(edf)); err != nil {
			return errors.Join(ErrInfoEDF, err)
		}
	}

	return nil
}
