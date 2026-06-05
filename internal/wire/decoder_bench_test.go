package wire

import (
	"encoding/binary"
	"testing"
)

var benchmarkDecodedString string
var benchmarkByte byte
var benchmarkUint16 uint16
var benchmarkUint32 uint32
var benchmarkUint64 uint64
var benchmarkFloat32 float32

func BenchmarkDecoderCString(b *testing.B) {
	data := []byte("example server\x00trailing")
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		decoder := NewDecoder(data)
		value, err := decoder.CString()
		if err != nil {
			b.Fatal(err)
		}
		benchmarkDecodedString = value
	}
}

func BenchmarkDecoderPrimitives(b *testing.B) {
	data := []byte{
		1,
		1,
		0x34, 0x12,
		0x78, 0x56, 0x34, 0x12,
		0xef, 0xcd, 0xab, 0x89, 0x67, 0x45, 0x23, 0x01,
		0x00, 0x00, 0x80, 0x3f,
	}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		decoder := NewDecoder(data)
		var err error
		if benchmarkByte, err = decoder.Byte(); err != nil {
			b.Fatal(err)
		}
		if benchmarkByte, err = decoder.Byte(); err != nil {
			b.Fatal(err)
		}
		if benchmarkUint16, err = decoder.Uint16(); err != nil {
			b.Fatal(err)
		}
		if benchmarkUint32, err = decoder.Uint32(); err != nil {
			b.Fatal(err)
		}
		if benchmarkUint64, err = decoder.Uint64(); err != nil {
			b.Fatal(err)
		}
		if benchmarkFloat32, err = decoder.Float32(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecoderInfoLikeSequence(b *testing.B) {
	data := []byte{17}
	for _, value := range []string{"server", "map", "folder", "game"} {
		data = append(data, value...)
		data = append(data, 0)
	}
	data = binary.LittleEndian.AppendUint16(data, 1234)
	data = append(data, 1, 16, 2, 'd', 'w', 1, 0)
	data = append(data, "1.0"...)
	data = append(data, 0, 0)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		decoder := NewDecoder(data)
		var err error
		if benchmarkByte, err = decoder.Byte(); err != nil {
			b.Fatal(err)
		}
		for range 4 {
			if benchmarkDecodedString, err = decoder.CString(); err != nil {
				b.Fatal(err)
			}
		}
		if benchmarkUint16, err = decoder.Uint16(); err != nil {
			b.Fatal(err)
		}
		for range 2 {
			if benchmarkByte, err = decoder.Byte(); err != nil {
				b.Fatal(err)
			}
		}
	}
}
