package keywords

import (
	"strings"
	"time"
)

// DayZ keywords
type DayZ struct {
	Shard          string        `json:"shard,omitempty"`         // Shard name 000/001 for official
	Unknowns       []string      `json:"unknowns,omitempty"`      // Unparsed keywords
	Time           time.Duration `json:"time,omitempty"`          // Time on server
	TimeDayAccel   float64       `json:"etm,omitempty"`           // Time acceleration
	TimeNightAccel float64       `json:"entm,omitempty"`          // Night time acceleration
	GamePort       uint16        `json:"port,omitempty"`          // Game port
	PlayersQueue   uint8         `json:"lqs,omitempty"`           // Players in queue
	BattlEye       bool          `json:"battleye,omitempty"`      // Protected with BattlEye
	NoThirdPerson  bool          `json:"no3rd,omitempty"`         // 3rd person view disabled
	External       bool          `json:"external,omitempty"`      // Is external server (community server)
	PrivateHive    bool          `json:"private,omitempty"`       // Hive is private
	Modded         bool          `json:"mod,omitempty"`           // Require mods
	Whitelist      bool          `json:"whitelist,omitempty"`     // White list enabled for join
	FlePatching    bool          `json:"file_patching,omitempty"` // Enabled fle patching
	DLC            bool          `json:"dlc,omitempty"`           // Require DLC
}

// Parser for DayZ keywords
func ParseDayZ(keywords []string) *DayZ {
	data := &DayZ{}
	data.Parse(keywords)

	return data
}

// Parse A2S INFO gametype data for DayZ
func (d *DayZ) Parse(keywords []string) {
	for _, tag := range keywords {
		if tag == "" {
			continue
		}

		switch {
		case tag == "battleye":
			d.BattlEye = true

		case tag == "no3rd":
			d.NoThirdPerson = true

		case tag == "external":
			d.External = true

		case tag == "privHive":
			d.PrivateHive = true

		case strings.HasPrefix(tag, "shard"):
			d.Shard = strings.TrimPrefix(tag, "shard")

		case strings.HasPrefix(tag, "lqs"):
			d.PlayersQueue = ParseUint8(tag[3:])

		case strings.HasPrefix(tag, "etm"):
			d.TimeDayAccel = parseFloat64(tag[3:])

		case strings.HasPrefix(tag, "entm"):
			d.TimeNightAccel = parseFloat64(tag[4:])

		case tag == "mod":
			d.Modded = true

		case strings.HasPrefix(tag, "port"):
			d.GamePort = ParseUint16(tag[4:])

		case tag == "whitelisting":
			d.Whitelist = true

		case tag == "allowedFilePatching":
			d.FlePatching = true

		case tag == "isDLC":
			d.DLC = true

		case len(tag) == 5 && strings.Contains(tag, ":"):
			if t, err := time.ParseDuration(tag[:2] + "h" + tag[3:] + "m"); err == nil {
				d.Time = t
			}

		default:
			d.Unknowns = append(d.Unknowns, tag)
		}
	}
}
