package a2s

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/woozymasta/a2s/internal/bread"
)

// Read buffer for populate Info struct with EDF (Extra Data Flag) for Source protocol
func (i *Info) readEDF(buf *bytes.Buffer, edf EDF) error {
	var err error
	i.EDF = edf

	if (edf & edfPort) != 0 {
		if i.Port, err = bread.Uint16(buf); err != nil {
			return fmt.Errorf("game port: %w", err)
		}
	}

	if (edf & edfSteamID) != 0 {
		if i.SteamID, err = bread.Uint64(buf); err != nil {
			return fmt.Errorf("SteamID: %w", err)
		}
	}

	if (edf & edfSourceTV) != 0 {
		if i.SourceTVPort, err = bread.Uint16(buf); err != nil {
			return fmt.Errorf("SourceTV port: %w", err)
		}
		if i.SourceTVName, err = bread.String(buf); err != nil {
			return fmt.Errorf("SourceTV name: %w", err)
		}
	}

	if (edf & edfKeywords) != 0 {
		kw, err := bread.String(buf)
		if err != nil {
			return fmt.Errorf("keywords name: %w", err)
		}
		i.Keywords = strings.Split(kw, ",")
	}

	if (edf & edfGameID) != 0 {
		if i.ID, err = bread.Uint64(buf); err != nil {
			return fmt.Errorf("GameID: %w", err)
		}
	}

	return nil
}
