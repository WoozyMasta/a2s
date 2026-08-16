// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"errors"

	"github.com/woozymasta/a2s/internal/wire"
)

// readGoldSourceMods parses mod information from GoldSource protocol (obsolete).
func readGoldSourceMods(r *wire.Decoder) (*ModInfo, error) {
	info := &ModInfo{}

	var err error
	if info.Link, err = r.CString(); err != nil {
		return nil, errors.Join(ErrInfoGSModLink, err)
	}

	if info.DownloadLink, err = r.CString(); err != nil {
		return nil, errors.Join(ErrInfoGSModDownloadLink, err)
	}

	if info.Version, err = r.Uint32(); err != nil {
		return nil, errors.Join(ErrInfoGSModVersion, err)
	}

	if info.Size, err = r.Uint32(); err != nil {
		return nil, errors.Join(ErrInfoGSModSize, err)
	}

	if info.Type, err = readInfoBool(r); err != nil {
		return nil, errors.Join(ErrInfoGSModType, err)
	}

	if info.DLL, err = readInfoBool(r); err != nil {
		return nil, errors.Join(ErrInfoGSModDLL, err)
	}

	return info, nil
}
