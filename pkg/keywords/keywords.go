// Package keywords provide additional parsers for tags (sv_tag) from the A2S_INFO response
package keywords

import (
	"fmt"

	"github.com/woozymasta/a2s/pkg/appid"
)

// Universal function for outputting result depending on application ID,
// if parser exists it will return updated structure, otherwise it will return error
func Parse(id uint64, keywords any) (any, error) {
	kw, ok := keywords.([]string)
	if !ok {
		return nil, fmt.Errorf("already processed data, parsing only for slice of strings")
	}

	switch id {
	case appid.DayZ, appid.DayZExp:
		data := &DayZ{}
		data.Parse(kw)
		return data, nil

	default:
		return nil, fmt.Errorf("unsupported application ID %d", id)
	}
}
