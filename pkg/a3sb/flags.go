package a3sb

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/internal/bread"
)

// Some flags stored in response (wtf is it?)
type Flags struct {
	Flag0 bool `json:"0,omitempty"` // Some 0 flag
	Flag1 bool `json:"1,omitempty"` // Some 1 flag
	Flag2 bool `json:"2,omitempty"` // Some 2 flag
	Flag3 bool `json:"3,omitempty"` // Some 3 flag
	Flag4 bool `json:"4,omitempty"` // Some 4 flag
	Flag5 bool `json:"5,omitempty"` // Some 5 flag
	Flag6 bool `json:"6,omitempty"` // Some 6 flag
	Flag7 bool `json:"7,omitempty"` // Some 7 flag
}

// Read flags from Arma 3 server browser proto
// i don't know what is it
func (r *Rules) readFlags(buf *bytes.Buffer) error {
	value, err := bread.Byte(buf)
	if err != nil {
		return fmt.Errorf("flags: %w", err)
	}
	if value == 0 {
		return nil
	}

	r.Flags = &Flags{
		Flag0: value&(1<<0) != 0,
		Flag1: value&(1<<1) != 0,
		Flag2: value&(1<<2) != 0,
		Flag3: value&(1<<3) != 0,
		Flag4: value&(1<<4) != 0,
		Flag5: value&(1<<5) != 0,
		Flag6: value&(1<<6) != 0,
		Flag7: value&(1<<7) != 0,
	}

	return nil
}
