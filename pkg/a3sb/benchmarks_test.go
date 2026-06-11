package a3sb

import (
	"encoding/binary"
	"testing"

	"github.com/woozymasta/a2s/internal/a2srules"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/appid"
)

func BenchmarkParseA3SBArma3Fixture(b *testing.B) {
	envelope := a3sbEnvelope{encodedPages: benchmarkArma3Payload()}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := parseA3SBEnvelope(envelope, appid.Arma3); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseA3SBDayZFixture(b *testing.B) {
	envelope := a3sbEnvelope{
		encodedPages: benchmarkDayZPayload(),
		extraRules: a2s.Rules{
			{Name: "allowedBuild", Value: "123"},
			{Name: "clientPort", Value: "2303"},
			{Name: "dedicated", Value: "1"},
			{Name: "island", Value: "Chernarus"},
			{Name: "language", Value: "0"},
			{Name: "platform", Value: "win"},
			{Name: "requiredBuild", Value: "123"},
			{Name: "requiredVersion", Value: "1"},
			{Name: "timeLeft", Value: "45"},
		},
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := parseA3SBEnvelope(envelope, appid.DayZ); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseAutomaticA3SBFixture(b *testing.B) {
	result := a2srules.Result{
		Entries: []a2srules.Entry{{
			Key:   []byte{1, 1},
			Value: benchmarkEncodeA3SB(benchmarkArma3Payload()),
		}},
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := parseAutomatic(result); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAssembleA3SBPagesFixture(b *testing.B) {
	pages := map[byte][]byte{
		1: []byte("page-1"),
		2: []byte("page-2"),
		3: []byte("page-3"),
		4: []byte("page-4"),
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := assemblePages(pages, 4); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkArma3Payload() []byte {
	data := []byte{3, 0, 0, 0, 0x41, 0x01, 2}
	data = binary.LittleEndian.AppendUint32(data, 0x12345678)
	data = append(data, 4)
	data = binary.LittleEndian.AppendUint32(data, 123456)
	data = append(data, 4, 'm', 'o', 'd', '1')
	data = binary.LittleEndian.AppendUint32(data, 0x87654321)
	data = append(data, 4)
	data = binary.LittleEndian.AppendUint32(data, 654321)
	data = append(data, 4, 'm', 'o', 'd', '2', 2, 4, 'k', 'e', 'y', '1', 4, 'k', 'e', 'y', '2')
	return data
}

func benchmarkDayZPayload() []byte {
	data := []byte{2, 0, 0, 0, 1}
	data = binary.LittleEndian.AppendUint32(data, 0x12345678)
	data = append(data, 4)
	data = binary.LittleEndian.AppendUint32(data, 123456)
	data = append(data, 4, 'm', 'o', 'd', '1', 1, 4, 'k', 'e', 'y', '1')
	data = append(data, 11)
	data = append(data, "Test server"...)
	return data
}

func benchmarkEncodeA3SB(data []byte) []byte {
	encoded := make([]byte, 0, len(data))
	for _, value := range data {
		switch value {
		case 0x00:
			encoded = append(encoded, 0x01, 0x02)
		case 0x01:
			encoded = append(encoded, 0x01, 0x01)
		case 0xff:
			encoded = append(encoded, 0x01, 0x03)
		default:
			encoded = append(encoded, value)
		}
	}
	return encoded
}
