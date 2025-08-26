package a2s

import (
	"encoding/json"
)

// Flag represents the bytes for the request and response type in the header
type Flag byte

// EDF represents the bytes for Extra Data Flag (EDF) in A2S_INFO response
type EDF byte

// InfoFormat represents the bytes for engine type: Source or GoldSource in A2S_INFO response
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

// MarshalJSON helper for convert internal type to string in JSON
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

// MarshalJSON helper for convert internal type to string in JSON
func (s ServerType) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// Environment represents the bytes for server OS: Linux, Windows, Mac or Other in A2S_INFO response
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

// MarshalJSON helper for convert internal type to string in JSON
func (e Environment) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

// TheShipMode represents the bytes for TheShip game extra mode info in A2S_INFO response
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

// MarshalJSON helper for convert internal type to string in JSON
func (m TheShipMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}
