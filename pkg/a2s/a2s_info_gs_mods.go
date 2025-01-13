package a2s

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/internal/bread"
)

// Read buffer and return ModInfo struct for GoldSource protocol (Obsolete)
func readGoldSourceMods(buf *bytes.Buffer) (*ModInfo, error) {
	info := &ModInfo{}

	var err error
	if info.Link, err = bread.String(buf); err != nil {
		return nil, fmt.Errorf("link: %w", err)
	}

	if info.DownloadLink, err = bread.String(buf); err != nil {
		return nil, fmt.Errorf("download link: %w", err)
	}

	if info.Version, err = bread.Uint32(buf); err != nil {
		return nil, fmt.Errorf("version: %w", err)
	}

	if info.Size, err = bread.Uint32(buf); err != nil {
		return nil, fmt.Errorf("size: %w", err)
	}

	if info.Type, err = bread.Bool(buf); err != nil {
		return nil, fmt.Errorf("type: %w", err)
	}

	if info.DLL, err = bread.Bool(buf); err != nil {
		return nil, fmt.Errorf("DLL: %w", err)
	}

	return info, nil
}
