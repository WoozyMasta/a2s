package a2s

import (
	"encoding/json"
)

// Flag represents request/response type byte in A2S protocol header.
type Flag byte

// EDF represents Extra Data Flag bits in A2S_INFO response.
type EDF byte

// InfoFormat represents engine type (Source or GoldSource) in A2S_INFO response.
type InfoFormat byte

func (i InfoFormat) String() string {
	switch Flag(i) {
	case infoResponseSource:
		return "Source"
	case infoResponseGoldSource:
		return "GoldSource"
	}

	return "unknown"
}

// MarshalJSON converts InfoFormat to JSON string.
func (i InfoFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// ServerType represents the bytes for server type: Dedicated, Local or Proxy (SteamTV/HLTV) in A2S_INFO response
type ServerType byte

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
