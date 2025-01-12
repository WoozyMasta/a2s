package a3sb

import (
	"bytes"
	"fmt"

	"github.com/woozymasta/a2s/pkg/bread"
	"github.com/woozymasta/steam/utils/appid"
)

/*
Read protocol version and try set game ID if not set

There are two described versions of the [Protocol v3] and [Protocol v2] for Arma 3.

Currently Arma 3 responds with v3, and DayZ with v2, but v2 DayZ is not equal to the described v2 Arma 3,
DayZ does not have 5 and 6 bytes with bit flags describing the difficulty,
i.e. v2 DayZ has its own protocol with its own versioning.
Considering that at the time of writing the protocol versions of the games are different,
we will automatically assume that v2 is the DayZ response,
and v3 is the response for Arma 3 if the game was not specified explicitly.

! Most likely it will break when DayZ switches to protocol v3 !

[Protocol v3]: https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol3
[Protocol v2]: https://community.bistudio.com/wiki/Arma_3:_ServerBrowserProtocol2
*/
func (r *Rules) readVersion(buf *bytes.Buffer) error {
	version, err := bread.Byte(buf)
	if err != nil {
		return err
	}

	switch version {
	case 1:
		return ErrProtoV1

	case 3:
		if r.id == 0 {
			r.id = appid.Arma3.Uint64()
		}
		if r.id == appid.DayZ.Uint64() {
			return ErrProtoV3
		}

	case 2:
		if r.id == 0 {
			r.id = appid.DayZ.Uint64()
		}

	default:
		return fmt.Errorf("%w: protocol version %d", ErrProtoNewest, version)
	}

	r.Version = version

	return nil
}
