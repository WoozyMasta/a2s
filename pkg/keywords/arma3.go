package keywords

import (
	"time"

	"github.com/woozymasta/a2s/pkg/keywords/types"
)

// Arma3 keywords
// https://community.bistudio.com/wiki/Arma_3:_STEAMWORKSquery
type Arma3 struct {
	GameType            types.GameType    `json:"gametype,omitempty"`             // Type of game
	Platform            types.Platform    `json:"platform,omitempty"`             // Server OS
	LoadedContentHash   string            `json:"loaded_content_hash,omitempty"`  // Hash of loaded content
	Country             string            `json:"country,omitempty"`              // Server country
	Island              string            `json:"island,omitempty"`               // Island name (unimplemented)
	Unknowns            []string          `json:"unknowns,omitempty"`             // Unparsed keywords
	TimeLeft            time.Duration     `json:"time_left,omitempty"`            // Time left of mission
	RequiredVersion     uint32            `json:"required_version,omitempty"`     // Required game version
	RequiredBuildNo     uint32            `json:"required_buildno,omitempty"`     // Required game build number
	Language            types.ServerLang  `json:"language,omitempty"`             // Server Language
	Longitude           int32             `json:"longitude,omitempty"`            // Coordinates: Longitude
	Latitude            int32             `json:"latitude,omitempty"`             // Coordinates: Latitude
	ServerState         types.ServerState `json:"server_state,omitempty"`         // State of server
	BattlEye            bool              `json:"battleye,omitempty"`             // Protected with BattlEye
	Difficulty          byte              `json:"difficulty,omitempty"`           // Difficulty on server
	EqualModRequired    bool              `json:"equal_mod_required,omitempty"`   // Require all mods equal server
	Lock                bool              `json:"lock,omitempty"`                 // Locked state
	VerifySignatures    bool              `json:"verify_signatures,omitempty"`    // Verify signatures
	Dedicated           bool              `json:"dedicated,omitempty"`            // Is dedicated server
	Param1              byte              `json:"param_1,omitempty"`              // First params
	Param2              byte              `json:"param_2,omitempty"`              // Second params
	AllowedFilePatching bool              `json:"allowed_filepatching,omitempty"` // Enabled fle patching
}

// ParseArma3 parser for Arma3 keywords
func ParseArma3(keywords []string) *Arma3 {
	data := &Arma3{}
	data.Parse(keywords)

	return data
}

// Parse A2S INFO gametype data for Arma3
func (d *Arma3) Parse(keywords []string) {
	for _, tag := range keywords {
		if tag == "" {
			continue
		}

		val := string([]rune(tag)[1:])

		switch string([]rune(tag)[0]) {
		case "b":
			d.BattlEye = parseBool(val)

		case "r":
			d.RequiredVersion = parseUint32(val)

		case "n":
			d.RequiredBuildNo = parseUint32(val)

		case "s":
			d.ServerState = types.ServerState(ParseUint8(val))

		case "i":
			d.Difficulty = ParseUint8(val)

		case "m":
			d.EqualModRequired = parseBool(val)

		case "l":
			d.Lock = parseBool(val)

		case "v":
			d.VerifySignatures = parseBool(val)

		case "d":
			d.Dedicated = parseBool(val)

		case "t":
			d.GameType = types.GameType(val)

		case "g":
			d.Language = types.ServerLang(parseUint32(val))

		case "c":
			d.Longitude, d.Latitude = parseCoordinates(val)

		case "p":
			d.Platform = types.Platform(val)

		case "h":
			d.LoadedContentHash = val

		case "o":
			d.Country = val

		case "e":
			if t, err := time.ParseDuration(val + "m"); err == nil {
				d.TimeLeft = t
			}

		case "j":
			d.Param1 = ParseUint8(val)

		case "k":
			d.Param1 = ParseUint8(val)

		case "f":
			d.AllowedFilePatching = parseBool(val)

		case "y":
			d.Island = val

		default:
			d.Unknowns = append(d.Unknowns, tag)
		}
	}
}
