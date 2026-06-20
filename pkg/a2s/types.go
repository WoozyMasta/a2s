package a2s

import (
	"encoding/json"
	"time"
)

// QueryType identifies an A2S request type byte.
type QueryType byte

// ResponseType identifies an A2S response type byte.
type ResponseType byte

// Packet is one logical A2S response using single-packet framing.
//
// Payload excludes the four-byte packet marker and response type byte.
// Decoders own the returned payload bytes,
// so the packet is independent from the input buffer used to decode it.
type Packet struct {
	// Payload is the response body without the packet marker and response type.
	Payload []byte

	// Type is the raw response type byte.
	Type ResponseType
}

// Request represents one complete A2S request datagram.
//
// INFO may omit its challenge on the initial request.
// PLAYER and RULES always require a challenge.
// HasChallenge preserves that wire-level distinction;
// Challenge is kept opaque and is not interpreted as an integer.
type Request struct {
	// Type is the request type byte.
	Type QueryType

	// Challenge is the opaque four-byte challenge token.
	Challenge Challenge

	// HasChallenge reports whether Challenge is present on the wire.
	HasChallenge bool
}

// Challenge is the opaque four-byte token used by A2S challenge exchanges.
// Its byte order is preserved exactly as received from or sent to the server.
type Challenge [4]byte

// InitialChallenge is the reserved token used to start a challenge-aware query.
var InitialChallenge = Challenge{0xFF, 0xFF, 0xFF, 0xFF}

// QueryMeta contains transport metadata for one completed client query.
// It is separate from protocol response models.
type QueryMeta struct {
	// Duration is the complete logical query latency.
	Duration time.Duration
}

// EDF represents Extra Data Flag bits in A2S_INFO response.
type EDF byte

// InfoFormat represents engine type (Source or GoldSource) in A2S_INFO response.
type InfoFormat byte

// String returns the human-readable engine name.
func (i InfoFormat) String() string {
	switch ResponseType(i) {
	case ResponseInfo:
		return "Source"

	case ResponseInfoGoldSource:
		return "GoldSource"
	}

	return "unknown"
}

// MarshalJSON converts InfoFormat to JSON string.
func (i InfoFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// ServerType represents the server type byte in an A2S_INFO response.
type ServerType byte

// String returns the human-readable server type.
func (s ServerType) String() string {
	switch s {
	case 0x64, 0x44: // d D
		return "Dedicated"
	case 0x6c, 0x4c: // l L
		return "Local"
	case 0x70, 0x50: // p P
		return "Proxy"
	}

	return "Unknown"
}

// MarshalJSON converts ServerType to JSON string.
func (s ServerType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// Environment represents server operating system in A2S_INFO response.
type Environment byte

// String returns the human-readable operating system name.
func (e Environment) String() string {
	switch e {
	case 0x6c, 0x4c: // l L
		return "Linux"
	case 0x77, 0x57: // w W
		return "Windows"
	case 0x6d, 0x4d: // m M
		return "Mac"
	case 0x6f, 0x4f: // o O
		return "Other"
	}

	return "Unknown"
}

// MarshalJSON converts Environment to JSON string.
func (e Environment) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

// TheShipMode represents game mode for The Ship game in A2S_INFO response.
type TheShipMode byte

// String returns the human-readable The Ship game mode.
func (m TheShipMode) String() string {
	switch m {
	case 0:
		return "Hunt"
	case 1:
		return "Elimination"
	case 2:
		return "Duel"
	case 3:
		return "Deathmatch"
	case 4:
		return "VIP Team"
	case 5:
		return "Team Elimination"
	}

	return "Unknown"
}

// MarshalJSON converts TheShipMode to JSON string.
func (m TheShipMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}
