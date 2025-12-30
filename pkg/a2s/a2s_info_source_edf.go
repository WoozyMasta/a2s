package a2s

import (
	"errors"

	"github.com/woozymasta/a2s/internal/bread"
)

// readEDF parses Extra Data Flag fields from Source protocol response.
func (i *Info) readEDF(r *bread.Reader, edf EDF) error {
	var err error
	i.EDF = edf

	if (edf & edfPort) != 0 {
		if i.Port, err = r.Uint16(); err != nil {
			return errors.Join(ErrInfoEDFPort, err)
		}
	}

	if (edf & edfSteamID) != 0 {
		if i.SteamID, err = r.Uint64(); err != nil {
			return errors.Join(ErrInfoEDFSteamID, err)
		}
	}

	if (edf & edfSourceTV) != 0 {
		if i.SourceTVPort, err = r.Uint16(); err != nil {
			return errors.Join(ErrInfoEDFSourceTVPort, err)
		}
		if i.SourceTVName, err = r.String(); err != nil {
			return errors.Join(ErrInfoEDFSourceTVName, err)
		}
	}

	if (edf & edfKeywords) != 0 {
		kwBytes, err := r.BytesPage()
		if err != nil {
			return errors.Join(ErrInfoEDFKeywords, err)
		}

		// Parse comma-separated keywords
		commaCount := 0
		for _, b := range kwBytes {
			if b == ',' {
				commaCount++
			}
		}

		keywords := make([]string, 0, commaCount+1)
		start := 0
		for i, b := range kwBytes {
			if b == ',' {
				if i > start {
					keywords = append(keywords, string(kwBytes[start:i]))
				}
				start = i + 1
			}
		}
		if start < len(kwBytes) {
			keywords = append(keywords, string(kwBytes[start:]))
		}

		i.Keywords = keywords
	}

	if (edf & edfGameID) != 0 {
		if i.ID, err = r.Uint64(); err != nil {
			return errors.Join(ErrInfoEDFGameID, err)
		}
	}

	return nil
}
