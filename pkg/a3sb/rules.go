package a3sb

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/bread"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/keywords/types"
	"github.com/woozymasta/steam/utils/appid"
)

// DefaultRulesBufferSize is default buffer size for A3SB rules responses.
const DefaultRulesBufferSize uint16 = 8192

// Rules contains parsed A3SB rules response data.
type Rules struct {
	Flags           *Flags            `json:"flags,omitempty"`            // Flags, I don't know what's actually encoded there
	Difficulty      *Difficulty       `json:"difficulty,omitempty"`       // Difficulty (Arma 3 only)
	ExtraRules      map[string]string `json:"extra_rules,omitempty"`      // Extra not standard rules if exists
	Description     string            `json:"description,omitempty"`      // Server description
	Island          string            `json:"island,omitempty"`           // Name of world [DayZ]
	Platform        string            `json:"platform,omitempty"`         // Server OS [DayZ]
	DLC             []DLCInfo         `json:"dlcs,omitempty"`             // List of information about DLC
	CreatorDLC      []DLCInfo         `json:"creator_dlc,omitempty"`      // List of information about Creator DLC (Arma 3 only)
	Mods            []Mod             `json:"mods,omitempty"`             // List of information about modifications
	Signatures      []string          `json:"signatures,omitempty"`       // List of signatures
	id              uint64            ``                                  // Steam AppID
	Language        types.ServerLang  `json:"language,omitempty"`         // DayZ Server Language [DayZ]
	AllowedBuild    uint16            `json:"allowed_build,omitempty"`    // Allowed client build for connect [DayZ]
	ClientPort      uint16            `json:"client_port,omitempty"`      // Client port [DayZ]
	RequiredBuild   uint16            `json:"required_build,omitempty"`   // Required client build for connect [DayZ]
	RequiredVersion uint16            `json:"required_version,omitempty"` // Required client version for connect [DayZ]
	TimeLeft        uint16            `json:"time_left,omitempty"`        // Time for respawn [DayZ]
	stats           [4]byte           ``                                  // a3sb pages count raw/pager/blank/overflow
	Version         byte              `json:"version"`                    // Protocol version
	Dedicated       bool              `json:"dedicated,omitempty"`        // Dedicated [DayZ]
}

// GetRulesArma3 returns A2S_RULES for Arma 3.
func (c *Client) GetRulesArma3() (*Rules, error) {
	return c.GetRules(appid.Arma3.Uint64())
}

// GetRulesDayZ returns A2S_RULES for DayZ.
func (c *Client) GetRulesDayZ() (*Rules, error) {
	return c.GetRules(appid.DayZ.Uint64())
}

// GetRules parses A2S_RULES response using A3SB for Arma 3 and DayZ.
func (c *Client) GetRules(game uint64) (*Rules, error) {
	if c.BufferSize == a2s.DefaultBufferSize {
		c.SetBufferSize(DefaultRulesBufferSize)
	}

	data, _, _, err := c.Get(a2s.RulesRequest)
	if err != nil {
		return nil, err
	}

	reader := bread.NewReader(data)

	count, err := reader.Uint16()
	if err != nil {
		return nil, fmt.Errorf("%w count: 0x%X", ErrRules, data[:4])
	}

	var a3sb []byte
	var rawRules map[string]string
	rules := &Rules{id: game, stats: [4]byte{data[1], 0, 0, 0}}

	for i := 0; i < int(count); i++ {
		key, err := reader.BytesPage()
		if err != nil {
			return nil, fmt.Errorf("%w key: %w", ErrRules, err)
		}
		value, err := reader.BytesPage()
		if err != nil {
			return nil, fmt.Errorf("%w value: %w", ErrRules, err)
		}

		if len(key) == 0 {
			rules.stats[2]++
			continue
		}

		if len(value) > 127 {
			rules.stats[3]++
		}

		// A3SBP pages have 2-byte keys: [page_number, page_count]
		if len(key) == 2 && key[0] <= key[1] {
			if a3sb == nil {
				remainingPages := int(count) - i
				estimatedSize := remainingPages * 64
				if estimatedSize > len(data) {
					estimatedSize = len(data)
				}
				a3sb = make([]byte, 0, estimatedSize)
			}
			a3sb = bread.AppendEscapeSequences(a3sb, value)
		} else {
			if rawRules == nil {
				rawRules = make(map[string]string, 8)
			}
			rawRules[string(key)] = string(value)
		}

		if rules.stats[1] == 0 {
			rules.stats[1] = key[1]
		}
	}

	if reader.Len() != 0 {
		return nil, ErrRulesDataRemains
	}

	if err := rules.readA3SB(a3sb); err != nil {
		return nil, err
	}

	if err := rules.parseRulesDayZ(rawRules); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRulesDayZ, err)
	}

	return rules, nil
}

// readA3SB parses Arma 3 Server Browser Protocol data.
func (r *Rules) readA3SB(data []byte) error {
	reader := bread.NewReader(data)
	var err error

	if err := r.readVersion(reader); err != nil {
		return fmt.Errorf("%w: %w", ErrVersion, err)
	}

	if err := r.readFlags(reader); err != nil {
		return fmt.Errorf("%w: %w", ErrFlags, err)
	}

	dlcMask, err := reader.Uint16()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDLC, err)
	}

	if err := r.readDifficulty(reader); err != nil {
		return fmt.Errorf("%w: %w", ErrDifficulty, err)
	}

	if dlcMask != 0 {
		if err := r.readDLC(reader, dlcMask); err != nil {
			return fmt.Errorf("%w: %w", ErrDLC, err)
		}
	}

	if err := r.readMods(reader); err != nil {
		return fmt.Errorf("%w: %w", ErrMod, err)
	}

	if err := r.readSignatures(reader); err != nil {
		return fmt.Errorf("%w: %w", ErrSignature, err)
	}

	// Stop here for arma3
	if reader.Len() == 0 {
		return nil
	}

	// DayZ-specific: server description
	descLen, err := reader.Byte()
	if err != nil {
		return fmt.Errorf("%w length: %w", ErrDescription, err)
	}
	if r.Description, err = reader.StringLen(int(descLen)); err != nil {
		return fmt.Errorf("%w: %w", ErrDescription, err)
	}

	if reader.Len() > 0 {
		// Get remaining bytes for error message
		pos := reader.Pos()
		remaining := data[pos:]
		return fmt.Errorf("%w: 0x%X (%s)", ErrRulesDataRemains, remaining, remaining)
	}

	return nil
}

// GetAppID returns the Steam AppID.
func (r *Rules) GetAppID() uint64 {
	return r.id
}

// GetReaderStats returns parsing statistics: [raw, pager, blank, overflow].
func (r *Rules) GetReaderStats() [4]byte {
	return r.stats
}
