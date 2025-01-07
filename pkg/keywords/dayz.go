package keywords

import (
	"strconv"
	"strings"
	"time"
)

type DayZ struct {
	Shard          string        `json:"shard,omitempty"`
	Time           time.Duration `json:"time,omitempty"`
	TimeDayAccel   float64       `json:"etm,omitempty"`
	TimeNightAccel float64       `json:"entm,omitempty"`
	GamePort       uint16        `json:"port,omitempty"`
	PlayersQueue   uint8         `json:"lqs,omitempty"`
	BattlEye       bool          `json:"battleye,omitempty"`
	NoThirdPerson  bool          `json:"no3rd,omitempty"`
	External       bool          `json:"external,omitempty"`
	PrivateHive    bool          `json:"private,omitempty"`
	Modded         bool          `json:"mod,omitempty"`
	Whitelist      bool          `json:"whitelist,omitempty"`
	FlePatching    bool          `json:"file_patching,omitempty"`
	DLC            bool          `json:"dlc,omitempty"`
}

func ParseDayZ(keywords []string) *DayZ {
	data := &DayZ{}
	data.Parse(keywords)

	return data
}

// parse A2S INFO gametype data for DayZ
func (d *DayZ) Parse(keywords []string) {
	for _, tag := range keywords {
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
			if num, err := strconv.ParseUint(tag[3:], 10, 8); err == nil {
				d.PlayersQueue = uint8(num)
			}

		case strings.HasPrefix(tag, "etm"):
			if num, err := strconv.ParseFloat(tag[3:], 64); err == nil {
				d.TimeDayAccel = num
			}

		case strings.HasPrefix(tag, "entm"):
			if num, err := strconv.ParseFloat(tag[4:], 64); err == nil {
				d.TimeNightAccel = num
			}

		case tag == "mod":
			d.Modded = true

		case strings.HasPrefix(tag, "port"):
			if num, err := strconv.ParseUint(tag[4:], 10, 8); err == nil {
				d.GamePort = uint16(num)
			}

		case tag == "whitelisting":
			d.Whitelist = true

		case tag == "allowedFilePatching":
			d.FlePatching = true

		case tag == "isDLC":
			d.DLC = true

		case strings.Contains(tag, ":"):
			if t, err := time.ParseDuration(tag[:2] + "h" + tag[3:] + "m"); err == nil {
				d.Time = t
			}
		}
	}
}
