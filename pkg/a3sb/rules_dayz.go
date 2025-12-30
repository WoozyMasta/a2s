package a3sb

import (
	"strconv"

	"github.com/woozymasta/a2s/pkg/keywords/types"
)

// parseRulesDayZ parses DayZ-specific rules from A2S_RULES key-value pairs.
func (r *Rules) parseRulesDayZ(data map[string]string) error {
	var err error
	var extra map[string]string

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
			r.Language = types.ServerLang(language) // #nosec G115

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
			if extra == nil {
				extra = make(map[string]string, 4)
			}
			extra[k] = v
		}
	}

	if len(extra) > 0 {
		r.ExtraRules = extra
	}

	return nil
}

func strToUint16(str string) (uint16, error) {
	number, err := strconv.ParseUint(str, 10, 16)
	if err != nil {
		return 0, err
	}

	return uint16(number), nil // #nosec G115
}
