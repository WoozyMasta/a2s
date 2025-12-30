package a2s

import (
	"errors"

	"github.com/woozymasta/a2s/internal/bread"
)

// readGoldSourceMods parses mod information from GoldSource protocol (obsolete).
func readGoldSourceMods(r *bread.Reader) (*ModInfo, error) {
	info := &ModInfo{}

	var err error
	if info.Link, err = r.String(); err != nil {
		return nil, errors.Join(ErrInfoGSModLink, err)
	}

	if info.DownloadLink, err = r.String(); err != nil {
		return nil, errors.Join(ErrInfoGSModDownloadLink, err)
	}

	if info.Version, err = r.Uint32(); err != nil {
		return nil, errors.Join(ErrInfoGSModVersion, err)
	}

	if info.Size, err = r.Uint32(); err != nil {
		return nil, errors.Join(ErrInfoGSModSize, err)
	}

	if info.Type, err = r.Bool(); err != nil {
		return nil, errors.Join(ErrInfoGSModType, err)
	}

	if info.DLL, err = r.Bool(); err != nil {
		return nil, errors.Join(ErrInfoGSModDLL, err)
	}

	return info, nil
}
