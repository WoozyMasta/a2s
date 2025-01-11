package a3sb

import (
	"fmt"
	"strconv"

	"github.com/woozymasta/a2s/pkg/keywords/types"
)

// Parse DayZ specific rules from A2S_RULES keywords
func (r *Rules) parseRulesDayZ(data map[string]string) error {
	var err error
	var extra = make(map[string]string)

	for k, v := range data {
		switch k {
		case "allowedBuild":
			r.AllowedBuild, err = strToUint16(v)
			if err != nil {
				return err
			}

		case "clientPort":
			r.ClientPort, err = strToUint16(v)
			if err != nil {
				return err
			}

		case "dedicated":
			r.Dedicated = (v == "0")

		case "island":
			r.Island = v

		case "language":
			language, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return err
			}
			if language > 4294967295 {
				return fmt.Errorf("language id %d overflow", language)
			}
			r.Language = types.ServerLang(language)

		case "platform":
			switch v {
			case "win":
				r.Platform = "Windows"
			case "lin", "?":
				r.Platform = "Linux"
			default:
				r.Platform = "Other"
			}

		case "requiredBuild":
			r.RequiredBuild, err = strToUint16(v)
			if err != nil {
				return err
			}

		case "requiredVersion":
			r.RequiredVersion, err = strToUint16(v)
			if err != nil {
				return err
			}

		case "timeLeft":
			r.TimeLeft, err = strToUint16(v)
			if err != nil {
				return err
			}

		default:
			extra[k] = v
		}
	}

	if len(extra) > 0 {
		r.ExtraRules = extra
	}

	return nil
}

// Parse string as uint16
func strToUint16(str string) (uint16, error) {
	number, err := strconv.ParseUint(str, 10, 16)
	if err != nil {
		return 0, err
	}
	if number > 65535 {
		return 0, fmt.Errorf("parse string \"%s\" to uint16 overflow", str)
	}

	return uint16(number), nil
}
