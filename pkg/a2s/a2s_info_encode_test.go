package a2s

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestAppendInfoMatchesSourceFixture(t *testing.T) {
	payload := readProtocolFixture(t, "source_info_payload.hex")
	info, err := DecodeInfo(Packet{Type: ResponseInfo, Payload: payload})
	if err != nil {
		t.Fatalf("DecodeInfo() error = %v", err)
	}

	encoded, err := AppendInfo(nil, *info)
	if err != nil {
		t.Fatalf("AppendInfo() error = %v", err)
	}

	// The fixture contains one ignored trailing zero after the EDF byte.
	// Typed encoding emits the canonical fields represented by Info only.
	want := singlePacketFixture(ResponseInfo, payload[:len(payload)-1])
	if !bytes.Equal(encoded, want) {
		t.Fatalf("encoded Source response = %X, want %X", encoded, want)
	}

	decodedPacket, err := DecodePacket(encoded)
	if err != nil {
		t.Fatalf("DecodePacket() error = %v", err)
	}
	decodedInfo, err := DecodeInfo(decodedPacket)
	if err != nil {
		t.Fatalf("DecodeInfo() round-trip error = %v", err)
	}
	if !reflect.DeepEqual(decodedInfo, info) {
		t.Fatalf("round-trip Info = %+v, want %+v", decodedInfo, info)
	}
}

func TestAppendInfoMatchesGoldSourceFixture(t *testing.T) {
	payload := readProtocolFixture(t, "goldsource_info_payload.hex")
	info, err := DecodeInfo(Packet{Type: ResponseInfoGoldSource, Payload: payload})
	if err != nil {
		t.Fatalf("DecodeInfo() error = %v", err)
	}

	encoded, err := AppendInfo(nil, *info)
	if err != nil {
		t.Fatalf("AppendInfo() error = %v", err)
	}

	want := singlePacketFixture(ResponseInfoGoldSource, payload)
	if !bytes.Equal(encoded, want) {
		t.Fatalf("encoded GoldSource response = %X, want %X", encoded, want)
	}
}

func TestAppendInfoRejectsUnrepresentableStrings(t *testing.T) {
	info := Info{
		Format: InfoFormat(ResponseInfo),
		Name:   "invalid\x00name",
	}

	if _, err := AppendInfo(nil, info); !errors.Is(err, ErrInfoEncode) {
		t.Fatalf("AppendInfo() error = %v, want ErrInfoEncode", err)
	}
}

func TestAppendInfoRejectsUnknownEDF(t *testing.T) {
	info := Info{
		Format: InfoFormat(ResponseInfo),
		EDF:    EDF(0x02),
	}

	if _, err := AppendInfo(nil, info); !errors.Is(err, ErrInfoEncode) {
		t.Fatalf("AppendInfo() error = %v, want ErrInfoEncode", err)
	}
}

func TestAppendInfoDerivesEDFFromFields(t *testing.T) {
	gameID := uint64(107410)
	info := Info{
		Format:      InfoFormat(ResponseInfo),
		Name:        "server",
		Map:         "map",
		Folder:      "folder",
		Game:        "game",
		Version:     "1.0",
		AppID:       1234,
		Port:        27015,
		Keywords:    []string{"one", "two"},
		GameID:      &gameID,
		ServerType:  ServerType('d'),
		Environment: Environment('l'),
		Visibility:  true,
		VAC:         true,
		Players:     1,
		MaxPlayers:  16,
		EDF:         0,
	}

	encoded, err := AppendInfo(nil, info)
	if err != nil {
		t.Fatalf("AppendInfo() error = %v", err)
	}
	decoded, err := decodeInfoPacketForTest(encoded)
	if err != nil {
		t.Fatalf("decode encoded info: %v", err)
	}
	if decoded.Port != info.Port || decoded.GameID == nil || *decoded.GameID != gameID ||
		!reflect.DeepEqual(decoded.Keywords, info.Keywords) {
		t.Fatalf("decoded optional fields = %+v, want %+v", decoded, info)
	}
	if decoded.EDF != edfPort|edfKeywords|edfGameID {
		t.Fatalf("derived EDF = 0x%X, want 0x%X", decoded.EDF, edfPort|edfKeywords|edfGameID)
	}
}

func decodeInfoPacketForTest(data []byte) (*Info, error) {
	packet, err := DecodePacket(data)
	if err != nil {
		return nil, err
	}
	return DecodeInfo(packet)
}
