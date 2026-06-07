package a2s

import (
	"encoding/binary"
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

func FuzzParseInfoSource(f *testing.F) {
	f.Add([]byte{17, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseInfo(data, ResponseInfo, 0)
	})
}

func FuzzParseInfoGoldSource(f *testing.F) {
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseInfo(data, ResponseInfoGoldSource, 0)
	})
}

func FuzzParsePlayers(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{1, 0, 'p', 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parsePlayers(data)
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
	f.Add(byte(InfoRequest), uint32(singlePacket))
	f.Add(byte(RulesRequest), uint32(0x12345678))
	f.Fuzz(func(t *testing.T, request byte, challenge uint32) {
		_, _ = createHeader(QueryType(request), challenge)
	})
}

func FuzzBinaryChallengeRoundTrip(f *testing.F) {
	f.Add(uint32(0x12345678))
	f.Fuzz(func(t *testing.T, value uint32) {
		data := make([]byte, 4)
		binary.LittleEndian.PutUint32(data, value)
		got, err := parseChallenge(data)
		if err != nil || got != value {
			t.Fatalf("challenge round-trip = (%d, %v), want %d", got, err, value)
		}
	})
}
