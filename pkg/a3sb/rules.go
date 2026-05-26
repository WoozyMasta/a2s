package a3sb

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/woozymasta/a2s/internal/a2srules"
	"github.com/woozymasta/a2s/internal/bread"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/appid"
	"github.com/woozymasta/a2s/pkg/keywords/types"
)

// DefaultRulesBufferSize is the A3SB-compatible A2S receive buffer size.
const DefaultRulesBufferSize uint16 = a2s.DefaultBufferSize

// Rules contains parsed A3SB rules response data.
//
// When GetRules is called in automatic mode and the response is native A2S,
// the typed A3SB fields remain zero-valued
// and ExtraRules contains the complete ordinary A2S rules map.
// Binary A3SB page values are never exposed through ExtraRules
// after an A3SB page candidate has been identified.
type Rules struct {
	Flags      *Flags      `json:"flags,omitempty"`      // Flags, I don't know what's actually encoded there
	Difficulty *Difficulty `json:"difficulty,omitempty"` // Difficulty (Arma 3 only)

	// ExtraRules contains ordinary A2S key/value properties that are not represented by typed fields.
	//
	// In native A2S automatic fallback it contains the complete rules map;
	// it never contains rejected A3SB pages.
	ExtraRules map[string]string `json:"extra_rules,omitempty"`

	Description     string           `json:"description,omitempty"`      // Server description
	Island          string           `json:"island,omitempty"`           // Name of world [DayZ]
	Platform        string           `json:"platform,omitempty"`         // Server OS [DayZ]
	DLC             []DLCInfo        `json:"dlcs,omitempty"`             // List of information about DLC
	CreatorDLC      []DLCInfo        `json:"creator_dlc,omitempty"`      // List of information about Creator DLC (Arma 3 only)
	Mods            []Mod            `json:"mods,omitempty"`             // List of information about modifications
	Signatures      []string         `json:"signatures,omitempty"`       // List of signatures
	id              uint64           ``                                  // Steam AppID used to select protocol variants.
	Language        types.ServerLang `json:"language,omitempty"`         // DayZ Server Language [DayZ]
	AllowedBuild    uint16           `json:"allowed_build,omitempty"`    // Allowed client build for connect [DayZ]
	ClientPort      uint16           `json:"client_port,omitempty"`      // Client port [DayZ]
	RequiredBuild   uint16           `json:"required_build,omitempty"`   // Required client build for connect [DayZ]
	RequiredVersion uint16           `json:"required_version,omitempty"` // Required client version for connect [DayZ]
	TimeLeft        uint16           `json:"time_left,omitempty"`        // Time for respawn [DayZ]
	stats           [4]byte          ``                                  // A3SB page counts: raw, paged, blank, overflow.
	Version         byte             `json:"version"`                    // Protocol version
	Dedicated       bool             `json:"dedicated,omitempty"`        // Dedicated [DayZ]
}

// a3sbEnvelope contains the validated outer response and assembled page data.
// Page values are copied while the envelope is built because the input buffer
// belongs to the query transaction and must not be retained by a result.
type a3sbEnvelope struct {
	extraRules   map[string]string
	encodedPages []byte
	pageCount    byte
	blankCount   byte
	overflow     byte
}

// GetRulesArma3 returns A2S_RULES for Arma 3.
func (c *Client) GetRulesArma3(ctx context.Context) (*Rules, error) {
	return c.GetRules(ctx, appid.Arma3)
}

// GetRulesDayZ returns A2S_RULES for DayZ.
func (c *Client) GetRulesDayZ(ctx context.Context) (*Rules, error) {
	return c.GetRules(ctx, appid.DayZ)
}

// GetRules parses A2S_RULES using the selected A3SB layout.
//
// A non-zero game selects an explicit layout and never falls back to native A2S parsing.
// With game set to zero, the complete response is classified automatically:
// native A2S rules are returned in ExtraRules,
// while A3SB version 2 and version 3 select DayZ and Arma 3 respectively.
func (c *Client) GetRules(ctx context.Context, game uint64) (*Rules, error) {
	data, _, _, err := c.Get(ctx, a2s.RulesRequest)
	if err != nil {
		return nil, err
	}

	result, err := a2srules.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRules, err)
	}

	if game == 0 {
		return parseAutomatic(result)
	}

	envelope, err := buildPageEnvelope(result.Entries, result.Remaining, false)
	if err != nil {
		return nil, err
	}

	return parseA3SBEnvelope(envelope, game)
}

// parseAutomatic classifies one already fetched A2S_RULES response.
func parseAutomatic(result a2srules.Result) (*Rules, error) {
	if !hasPageOneCandidate(result.Entries) {
		return nativeRules(result.Entries), nil
	}

	envelope, err := buildPageEnvelope(result.Entries, result.Remaining, true)
	if err != nil {
		return nil, err
	}

	if len(envelope.encodedPages) == 0 {
		return nil, ErrRulesPageMissing
	}

	version := envelope.encodedPages[0]
	var firstGame, secondGame uint64
	switch version {
	case 1:
		return nil, ErrProtoV1
	case 2:
		firstGame, secondGame = appid.DayZ, appid.Arma3
	case 3:
		firstGame, secondGame = appid.Arma3, appid.DayZ
	default:
		return nil, fmt.Errorf("%w: protocol version %d", ErrProtoNewest, version)
	}

	first, firstErr := parseA3SBEnvelope(envelope, firstGame)
	if firstErr == nil {
		return first, nil
	}

	second, secondErr := parseA3SBEnvelope(envelope, secondGame)
	if secondErr == nil {
		return second, nil
	}

	// A coherent A3SB envelope must not degrade into binary strings
	// in ExtraRules after both known layouts reject it.
	return nil, errors.Join(firstErr, secondErr)
}

// nativeRules returns ordinary A2S rules only when no A3SB page-1 candidate was found.
// This prevents binary carrier pages from leaking into ExtraRules.
func nativeRules(entries []a2srules.Entry) *Rules {
	return &Rules{ExtraRules: a2srules.Map(entries)}
}

// hasPageOneCandidate checks only the unambiguous first A3SB page marker.
// Other two-byte keys are not enough to classify an ordinary A2S response.
func hasPageOneCandidate(entries []a2srules.Entry) bool {
	for _, entry := range entries {
		if len(entry.Key) == 2 && entry.Key[0] == 1 && entry.Key[1] != 0 {
			return true
		}
	}

	return false
}

// buildPageEnvelope validates and assembles A3SB pages.
// When requirePageOne is true, page 1 must be present;
// explicit game mode passes false to retain the existing strict error for responses without any pages.
func buildPageEnvelope(entries []a2srules.Entry, remaining []byte, requirePageOne bool) (a3sbEnvelope, error) {
	if len(remaining) != 0 {
		return a3sbEnvelope{}, ErrRulesDataRemains
	}

	pageValues := make(map[byte][]byte)
	var pageCount byte
	var blankCount byte
	var overflow byte
	var rawRules map[string]string
	pageOnePresent := false

	for _, entry := range entries {
		if len(entry.Key) == 0 {
			blankCount++
			continue
		}

		if len(entry.Value) > 127 {
			overflow++
		}

		if len(entry.Key) != 2 {
			if rawRules == nil {
				rawRules = make(map[string]string, 8)
			}
			rawRules[string(entry.Key)] = string(entry.Value)
			continue
		}

		pageNumber := entry.Key[0]
		advertisedCount := entry.Key[1]
		if pageNumber == 1 {
			pageOnePresent = true
		}
		if pageNumber == 0 || advertisedCount == 0 || pageNumber > advertisedCount {
			return a3sbEnvelope{}, fmt.Errorf(
				"%w: page %d of %d",
				ErrRulesPageMetadata,
				pageNumber,
				advertisedCount,
			)
		}

		if pageCount == 0 {
			pageCount = advertisedCount
		} else if pageCount != advertisedCount {
			return a3sbEnvelope{}, fmt.Errorf(
				"%w: page %d advertises %d, want %d",
				ErrRulesPageMetadata,
				pageNumber,
				advertisedCount,
				pageCount,
			)
		}

		if previous, ok := pageValues[pageNumber]; ok {
			if !bytes.Equal(previous, entry.Value) {
				return a3sbEnvelope{}, fmt.Errorf("%w: page %d", ErrRulesPageConflict, pageNumber)
			}

			continue
		}

		pageValues[pageNumber] = append([]byte(nil), entry.Value...)
	}

	if requirePageOne && !pageOnePresent {
		return a3sbEnvelope{}, ErrRulesPageMissing
	}

	encodedPages, err := assemblePages(pageValues, pageCount)
	if err != nil {
		return a3sbEnvelope{}, err
	}

	return a3sbEnvelope{
		encodedPages: bread.AppendEscapeSequences(nil, encodedPages),
		extraRules:   rawRules,
		pageCount:    pageCount,
		blankCount:   blankCount,
		overflow:     overflow,
	}, nil
}

// parseA3SBEnvelope parses one validated payload with one explicit layout.
func parseA3SBEnvelope(envelope a3sbEnvelope, game uint64) (*Rules, error) {
	rules := &Rules{
		id: game,
		stats: [4]byte{
			countByte(len(envelope.extraRules)),
			envelope.pageCount,
			envelope.blankCount,
			envelope.overflow,
		},
	}

	if err := rules.readA3SB(envelope.encodedPages); err != nil {
		return nil, err
	}

	if isDayZGame(game) {
		if err := rules.parseRulesDayZ(envelope.extraRules); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrRulesDayZ, err)
		}
	} else {
		rules.ExtraRules = envelope.extraRules
	}

	return rules, nil
}

// isDayZGame identifies both stable and experimental DayZ AppIDs.
func isDayZGame(game uint64) bool {
	return game == appid.DayZ || game == appid.DayZExperimental
}

// countByte stores bounded parser statistics in the legacy byte-sized field.
func countByte(value int) byte {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}

	return byte(value)
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
