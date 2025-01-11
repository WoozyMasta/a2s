//go:build !a2s_no_marshal

package types //nolint:list

import (
	"encoding/json"
)

// MarshalJSON for GameType
func (gt GameType) MarshalJSON() ([]byte, error) {
	return json.Marshal(gt.String())
}

// MarshalJSON for ServerLang
func (sl ServerLang) MarshalJSON() ([]byte, error) {
	return json.Marshal(sl.String())
}

// MarshalJSON for Platform
func (p Platform) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// MarshalJSON for ServerState
func (ss ServerState) MarshalJSON() ([]byte, error) {
	return json.Marshal(ss.String())
}
