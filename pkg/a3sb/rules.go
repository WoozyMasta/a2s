package a3sb

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/appid"
	"github.com/woozymasta/a2s/pkg/bread"
)

// Structure for storing data from the A3SBP response
type Rules struct {
	// Internal

	id    uint64  // Steam AppID
	stats [4]byte // a3sb pages count raw/pager/blank/overflow

	// A3SBP fields

	Version     byte        `json:"version"`               // Protocol version
	Flags       *Flags      `json:"flags,omitempty"`       // Flags, I don't know what's actually encoded there
	DLC         []DLCInfo   `json:"dlcs,omitempty"`        // List of information about DLC
	CreatorDLC  []DLCInfo   `json:"creator_dlc,omitempty"` // List of information about Creator DLC (Arma 3 only)
	Difficulty  *Difficulty `json:"difficulty,omitempty"`  // Difficulty (Arma 3 only)
	Mods        []Mod       `json:"mods,omitempty"`        // List of information about modifications
	Signatures  []string    `json:"signatures,omitempty"`  // List of signatures
	Description string      `json:"description,omitempty"` // Server description

	// Extra not standard rules if exists
	ExtraRules map[string]string `json:"extra_rules,omitempty"`

	// DayZ specific rules

	AllowedBuild    uint16     `json:"allowed_build,omitempty"`    // Allowed client build for connect
	ClientPort      uint16     `json:"client_port,omitempty"`      // Client port
	Dedicated       bool       `json:"dedicated,omitempty"`        // Dedicated
	Island          string     `json:"island,omitempty"`           // Name of world
	Language        ServerLang `json:"language,omitempty"`         // DayZ Server Language
	Platform        string     `json:"platform,omitempty"`         // Server OS
	RequiredBuild   uint16     `json:"required_build,omitempty"`   // Required client build for connect
	RequiredVersion uint16     `json:"required_version,omitempty"` // Required client version for connect
	TimeLeft        uint16     `json:"time_left,omitempty"`        // Time for respawn
}

// A2S_RULES for Arma (wrapper)
func (c *Client) GetRulesArma3() (*Rules, error) {
	return c.GetRules(appid.Arma3)
}

// A2S_RULES for DayZ (wrapper)
func (c *Client) GetRulesDayZ() (*Rules, error) {
	return c.GetRules(appid.DayZ)
}

// A2S_RULES for DayZ/Arma
func (c *Client) GetRules(game uint64) (*Rules, error) {
	// Need more for DayZ and Arma
	if c.BufferSize == a2s.DefaultBufferSize {
		c.SetBufferSize(8192)
	}

	data, _, _, err := c.Get(a2s.RulesRequest)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(data)

	// Get A2S_RULES records count
	count, err := bread.Uint16(buf)
	if err != nil {
		return nil, fmt.Errorf("%w count: 0x%X", ErrRules, data[:4])
	}

	var a3sb []byte
	var rawRules = make(map[string]string)
	rules := &Rules{id: game, stats: [4]byte{buf.Bytes()[1], 0, 0, 0}}

	for i := 0; i < int(count); i++ {
		key, err := bread.BytesPage(buf)
		if err != nil {
			return nil, fmt.Errorf("%w key: %w", ErrRules, err)
		}
		value, err := bread.BytesPage(buf)
		if err != nil {
			return nil, fmt.Errorf("%w value: %w", ErrRules, err)
		}

		// Skip a rule key-value pair if the key is empty
		if len(key) == 0 {
			rules.stats[2]++
			continue
		}

		// The length of the value must be no more than 127 bytes according to the specification
		// but will not break logic
		if len(value) > 127 {
			rules.stats[3]++
		}

		// If the key length is 2 bytes (page number and page count) and it is in the page range
		// fill the Arma 3 Server Browser Protocol (A3SBP) byte array with data, handling escape sequences
		if len(key) == 2 && key[0] <= key[1] {
			a3sb = append(a3sb, bread.EscapeSequences(value[:])...)

		} else { // Read A2S_RULES as is after A3SBP bytes
			rawRules[string(key)] = string(value)
		}

		if rules.stats[1] == 0 {
			rules.stats[1] = key[1]
		}
	}

	if buf.Len() != 0 {
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

// Read Arma 3 server browser protocol from merged and prepared data
func (r *Rules) readA3SB(data []byte) error {
	buf := bytes.NewBuffer(data)
	var err error

	if err := r.readVersion(buf); err != nil {
		return fmt.Errorf("%w: %w", ErrVersion, err)
	}

	if err := r.readFlags(buf); err != nil {
		return fmt.Errorf("%w: %w", ErrFlags, err)
	}

	dlcMask, err := bread.Uint16(buf)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDLC, err)
	}

	if err := r.readDifficulty(buf); err != nil {
		return fmt.Errorf("%w: %w", ErrDifficulty, err)
	}

	if dlcMask != 0 {
		if err := r.readDLC(buf, dlcMask); err != nil {
			return fmt.Errorf("%w: %w", ErrDLC, err)
		}
	}

	if err := r.readMods(buf); err != nil {
		return fmt.Errorf("%w: %w", ErrMod, err)
	}

	if err := r.readSignatures(buf); err != nil {
		return fmt.Errorf("%w: %w", ErrSignature, err)
	}

	// Stop here for arma3
	if buf.Len() == 0 {
		return nil
	}

	// Read DayZ server description
	descLen, err := bread.Byte(buf)
	if err != nil {
		return fmt.Errorf("%w length: %w", ErrDescription, err)
	}
	if r.Description, err = bread.StringLen(buf, int(descLen)); err != nil {
		return fmt.Errorf("%w: %w", ErrDescription, err)
	}

	if buf.Len() > 0 {
		return fmt.Errorf("%w: 0x%X (%s)", ErrRulesDataRemains, buf.Bytes(), buf.Bytes())
	}

	return nil
}

func (r *Rules) GetAppID() uint64 {
	return r.id
}

func (r *Rules) GetReaderStats() [4]byte {
	return r.stats
}
