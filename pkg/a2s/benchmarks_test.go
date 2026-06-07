package a2s

import (
	"bytes"
	"context"
	"encoding/binary"
	"hash/crc32"
	"math"
	"net"
	"testing"
	"time"
)

func BenchmarkParseInfoSourceFixture(b *testing.B) {
	data := benchmarkSourceInfo()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := parseInfo(data, ResponseInfo, 25*time.Millisecond); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseInfoGoldSourceFixture(b *testing.B) {
	data := benchmarkGoldSourceInfo()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := parseInfo(data, ResponseInfoGoldSource, 25*time.Millisecond); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParsePlayersFixture(b *testing.B) {
	data := benchmarkPlayers()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := parsePlayers(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseRulesFixture(b *testing.B) {
	data := benchmarkRules()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := parseRules(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseSplitHeadersFixture(b *testing.B) {
	assembled := singlePacketFixture(ResponseRules, bytes.Repeat([]byte("rules"), 256))
	packets := sourceSplitPacketSequence(0x12345678, assembled, 128)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, packet := range packets {
			if _, err := parseSplitHeader(packet); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkGetSplitFixture(b *testing.B) {
	assembled := singlePacketFixture(ResponseRules, bytes.Repeat([]byte("rules"), 256))
	packets := sourceSplitPacketSequence(0x12345678, assembled, 128)
	server := newBenchmarkUDPServer(b, packets)
	client, err := NewWithAddr(server.LocalAddr().(*net.UDPAddr), WithTimeout(time.Second))
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	b.SetBytes(int64(len(assembled)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, _, err := client.Get(context.Background(), RulesRequest); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecompressSplitFixture(b *testing.B) {
	assembled := append(
		singlePacketFixture(ResponseRules, []byte("compressed split fixture:")),
		bytes.Repeat([]byte("0123456789abcdef"), 20)...,
	)
	compressed := benchmarkCompressedPayload()
	checksum := crc32.ChecksumIEEE(assembled)
	b.SetBytes(int64(len(assembled)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := decompressBzip2(compressed, uint32(len(assembled)), checksum); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkSourceInfo() []byte {
	data := []byte{17}
	for _, value := range []string{"Test server", "test_map", "test_folder", "Test game"} {
		data = append(data, value...)
		data = append(data, 0)
	}
	data = binary.LittleEndian.AppendUint16(data, 1234)
	data = append(data, 16, 32, 2, 'd', 'w', 0, 1)
	data = append(data, "1.0"...)
	data = append(data, 0, 0)
	return data
}

func benchmarkGoldSourceInfo() []byte {
	data := make([]byte, 0, 128)
	for _, value := range []string{"127.0.0.1:27015", "Test server", "test_map", "test_folder", "Test game"} {
		data = append(data, value...)
		data = append(data, 0)
	}
	data = append(data, 16, 32, 48, 'd', 'w', 0, 0, 1, 8)
	return data
}

func benchmarkPlayers() []byte {
	data := []byte{16}
	for index := byte(0); index < 16; index++ {
		data = append(data, index)
		data = append(data, "player"...)
		data = append(data, index, 0)
		data = binary.LittleEndian.AppendUint32(data, uint32(int32(index)-8))
		data = binary.LittleEndian.AppendUint32(data, math.Float32bits(float32(index)+0.5))
	}
	return data
}

func benchmarkRules() []byte {
	data := make([]byte, 0, 2048)
	data = binary.LittleEndian.AppendUint16(data, 64)
	for index := 0; index < 64; index++ {
		data = append(data, "rule"...)
		data = append(data, byte('0'+index/10), byte('0'+index%10), 0)
		data = append(data, "value"...)
		data = append(data, byte('0'+index/10), byte('0'+index%10), 0)
	}
	return data
}

func benchmarkCompressedPayload() []byte {
	return []byte{
		0x42, 0x5a, 0x68, 0x39, 0x31, 0x41, 0x59, 0x26, 0x53, 0x59, 0xf3, 0x7d, 0x36, 0x5d,
		0x00, 0x00, 0xaf, 0x5d, 0x80, 0xc0, 0x00, 0x40, 0x00, 0x7f, 0xf0, 0x02, 0x00, 0x3f, 0x26,
		0xde, 0x40, 0x00, 0x00, 0xa0, 0x00, 0x72, 0x29, 0x30, 0x1a, 0x09, 0x84, 0x19, 0x31, 0x94,
		0x0a, 0x95, 0x40, 0x1e, 0xa1, 0xa7, 0xa8, 0x6c, 0xa0, 0x3d, 0x43, 0xf6, 0x33, 0xad,
		0xe4, 0x37, 0x11, 0xbc, 0x8a, 0x11, 0xc0, 0x8e, 0x24, 0x72, 0x23, 0x22, 0x39, 0x91,
		0x52, 0x2f, 0x5e, 0xa4, 0x66, 0x45, 0x88, 0xb6, 0xcb, 0x91, 0xf3, 0x16, 0x23, 0x4e,
		0xb9, 0x0f, 0x1a, 0x3d, 0x5f, 0x18, 0xa5, 0x30, 0x2b, 0x48, 0x00, 0x90, 0xa0, 0x8f,
		0x1f, 0xc5, 0xdc, 0x91, 0x4e, 0x14, 0x24, 0x3c, 0xdf, 0x4d, 0x97, 0x40,
	}
}

func newBenchmarkUDPServer(b *testing.B, packets [][]byte) *net.UDPConn {
	b.Helper()
	server, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		b.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		buffer := make([]byte, 64*1024)
		for {
			_, address, err := server.ReadFromUDP(buffer)
			if err != nil {
				return
			}
			for _, packet := range packets {
				if _, err := server.WriteToUDP(packet, address); err != nil {
					return
				}
			}
		}
	}()

	b.Cleanup(func() {
		_ = server.Close()
		<-done
	})
	return server
}
