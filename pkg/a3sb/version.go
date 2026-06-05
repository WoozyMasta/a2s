package a3sb

import (
	"fmt"

	"github.com/woozymasta/a2s/internal/wire"
	"github.com/woozymasta/a2s/pkg/appid"
)

/*
readVersion reads and validates the A3SB protocol version for the selected game layout.

Arma 3 currently responds with [Protocol v3].
DayZ currently responds with [Protocol v2],
but its v2 layout is not the same as the v2 layout described for Arma 3:
DayZ does not contain the two difficulty bytes present in the Arma 3 payload.
In practice, these are separate game layouts sharing the version field.

When the game is not specified, parseAutomatic currently uses this observed distinction
as a parser-selection heuristic: v2 is tried as DayZ first and v3 as Arma 3 first.
The version byte is not treated as a permanent game identity;
the opposite parser is attempted when the preferred layout fails.

A future DayZ protocol change may initially break automatic detection
or cause the wrong layout to be selected if the new payload remains structurally valid.
The fallback reduces this risk but does not eliminate it;
the selection rule and fixtures must be updated when that protocol change occurs.

! Warning: this heuristic may break when DayZ switches to protocol v3 !

[Protocol v3]: https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol3
[Protocol v2]: https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol2
*/
func (r *Rules) readVersion(reader *wire.Decoder) error {
	version, err := reader.Byte()
	if err != nil {
		return err
	}

	switch version {
	case 1:
		return ErrProtoV1

	case 3:
		if r.id == 0 {
			r.id = appid.Arma3
		}
		if isDayZGame(r.id) {
			return ErrProtoV3
		}

	case 2:
		if r.id == 0 {
			r.id = appid.DayZ
		}

	default:
		return fmt.Errorf("%w: protocol version %d", ErrProtoNewest, version)
	}

	r.Version = version

	return nil
}
