package a2s

import (
	"bytes"
	"hash/crc32"
	"testing"
)

func FuzzParseChallenge(f *testing.F) {
	f.Add([]byte{0x78, 0x56, 0x34, 0x12})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseChallenge(data)
	})
}

func FuzzParseChallengeResponse(f *testing.F) {
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, byte(ResponseChallenge), 0x78, 0x56, 0x34, 0x12})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseChallengeResponse(data)
	})
}

func FuzzDecodeInfoSource(f *testing.F) {
	f.Add([]byte{17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeInfo(Packet{Type: ResponseInfo, Payload: data})
	})
}

func FuzzDecodeInfoGoldSource(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeInfo(Packet{Type: ResponseInfoGoldSource, Payload: data})
	})
}

func FuzzParsePlayers(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{1, 0, 'p', 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodePlayers(Packet{Type: ResponsePlayers, Payload: data})
	})
}

func FuzzParseTheShipPlayers(f *testing.F) {
	f.Add([]byte{0, 0})
	f.Add([]byte{1, 0, 0, 'p', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseTheShipPlayers(data)
	})
}

func FuzzParseRules(f *testing.F) {
	f.Add([]byte{0, 0})
	f.Add([]byte{1, 0, 'm', 'o', 'd', 'e', 0, 'c', 'o', 'o', 'p', 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseRules(data)
	})
}

func FuzzParsePacketHeaders(f *testing.F) {
	f.Add([]byte{0xff, 0xff, 0xff, 0xff, byte(ResponsePlayers)})
	f.Add([]byte{0xfe, 0xff, 0xff, 0xff, 1, 0, 0, 0, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = isMultiPacket(data)
		_, _ = parseSplitHeader(data)
	})
}

func FuzzDecompressBzip2(f *testing.F) {
	f.Add([]byte("not bzip2"))
	f.Fuzz(func(t *testing.T, data []byte) {
		var size uint32
		if len(data) > 0 {
			size = uint32(len(data) % 1024)
		}

		_, _ = decompressBzip2(data, size, crc32.ChecksumIEEE(data))
	})
}

func FuzzCreateHeader(f *testing.F) {
	f.Add(byte(InfoRequest), []byte{0xff, 0xff, 0xff, 0xff})
	f.Add(byte(RulesRequest), []byte{0x78, 0x56, 0x34, 0x12})
	f.Fuzz(func(t *testing.T, request byte, challengeData []byte) {
		var challenge Challenge
		copy(challenge[:], challengeData)
		_, _ = createHeader(QueryType(request), challenge)
	})
}

func FuzzBinaryChallengeRoundTrip(f *testing.F) {
	f.Add([]byte{0x78, 0x56, 0x34, 0x12})
	f.Fuzz(func(t *testing.T, data []byte) {
		got, err := parseChallenge(data)
		if len(data) != len(Challenge{}) {
			if err == nil {
				t.Fatalf("parseChallenge(%X) returned nil error", data)
			}
			return
		}
		if err != nil || !bytes.Equal(got[:], data) {
			t.Fatalf("challenge round-trip = (%X, %v), want %X", got, err, data)
		}
	})
}
