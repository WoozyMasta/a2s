package a3sb

import (
	"bytes"
	"context"
	"fmt"

	"github.com/woozymasta/a2s/internal/appid"
	"github.com/woozymasta/a2s/internal/bread"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/keywords/types"
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
	id              uint64            ``                                  // Steam AppID used to select protocol variants.
	Language        types.ServerLang  `json:"language,omitempty"`         // DayZ Server Language [DayZ]
	AllowedBuild    uint16            `json:"allowed_build,omitempty"`    // Allowed client build for connect [DayZ]
	ClientPort      uint16            `json:"client_port,omitempty"`      // Client port [DayZ]
	RequiredBuild   uint16            `json:"required_build,omitempty"`   // Required client build for connect [DayZ]
	RequiredVersion uint16            `json:"required_version,omitempty"` // Required client version for connect [DayZ]
	TimeLeft        uint16            `json:"time_left,omitempty"`        // Time for respawn [DayZ]
	stats           [4]byte           ``                                  // A3SB page counts: raw, paged, blank, overflow.
	Version         byte              `json:"version"`                    // Protocol version
	Dedicated       bool              `json:"dedicated,omitempty"`        // Dedicated [DayZ]
}

// GetRulesArma3 returns A2S_RULES for Arma 3.
func (c *Client) GetRulesArma3(ctx context.Context) (*Rules, error) {
	return c.GetRules(ctx, appid.Arma3)
}

// GetRulesDayZ returns A2S_RULES for DayZ.
func (c *Client) GetRulesDayZ(ctx context.Context) (*Rules, error) {
	return c.GetRules(ctx, appid.DayZ)
}

// GetRules parses A2S_RULES response using A3SB for Arma 3 and DayZ.
func (c *Client) GetRules(ctx context.Context, game uint64) (*Rules, error) {
	if c.BufferSize() == a2s.DefaultBufferSize {
		if err := c.SetBufferSize(DefaultRulesBufferSize); err != nil {
			return nil, fmt.Errorf("set rules buffer size: %w", err)
		}
	}

	data, _, _, err := c.Get(ctx, a2s.RulesRequest)
	if err != nil {
		return nil, err
	}

	reader := bread.NewReader(data)

	count, err := reader.Uint16()
	if err != nil {
		return nil, fmt.Errorf("%w count: 0x%X", ErrRules, data)
	}

	// A3SB pages and ordinary key/value rules share one A2S_RULES response.
	// Keep raw pages separate until they can be ordered and decoded as a whole.
	pageValues := make(map[byte][]byte)
	var pageCount byte
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

		// A3SB pages have 2-byte keys: [page_number, page_count].
		// Ordinary rules use textual keys and remain available to the DayZ parser.
		if len(key) == 2 {
			pageNumber := key[0]
			advertisedCount := key[1]
			if pageNumber == 0 || advertisedCount == 0 || pageNumber > advertisedCount {
				return nil, fmt.Errorf(
					"%w: page %d of %d",
					ErrRulesPageMetadata,
					pageNumber,
					advertisedCount,
				)
			}

			if pageCount == 0 {
				pageCount = advertisedCount
			} else if pageCount != advertisedCount {
				return nil, fmt.Errorf(
					"%w: page %d advertises %d, want %d",
					ErrRulesPageMetadata,
					pageNumber,
					advertisedCount,
					pageCount,
				)
			}

			if previous, ok := pageValues[pageNumber]; ok {
				if !bytes.Equal(previous, value) {
					return nil, fmt.Errorf("%w: page %d", ErrRulesPageConflict, pageNumber)
				}

				continue
			}

			pageValues[pageNumber] = append([]byte(nil), value...)
		} else {
			if rawRules == nil {
				rawRules = make(map[string]string, 8)
			}
			rawRules[string(key)] = string(value)
		}
	}

	if reader.Len() != 0 {
		return nil, ErrRulesDataRemains
	}

	encodedPages, err := assemblePages(pageValues, pageCount)
	if err != nil {
		return nil, err
	}

	rules.stats[1] = pageCount
	a3sb := bread.AppendEscapeSequences(nil, encodedPages)

	if err := rules.readA3SB(a3sb); err != nil {
		return nil, err
	}

	if err := rules.parseRulesDayZ(rawRules); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRulesDayZ, err)
	}

	return rules, nil
}

// assemblePages orders one-based A3SB pages and concatenates their raw values.
func assemblePages(pages map[byte][]byte, pageCount byte) ([]byte, error) {
	if pageCount == 0 {
		return nil, fmt.Errorf("%w: no pages", ErrRulesPageMetadata)
	}

	totalSize := 0
	for pageNumber := 1; pageNumber <= int(pageCount); pageNumber++ {
		page, ok := pages[byte(pageNumber)]
		if !ok {
			return nil, fmt.Errorf("%w: page %d of %d", ErrRulesPageMissing, pageNumber, pageCount)
		}
		totalSize += len(page)
	}

	assembled := make([]byte, 0, totalSize)
	for pageNumber := 1; pageNumber <= int(pageCount); pageNumber++ {
		assembled = append(assembled, pages[byte(pageNumber)]...)
	}

	return assembled, nil
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

	// Arma 3 ends after signatures; remaining bytes identify the DayZ suffix.
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
