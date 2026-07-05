package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func TestSourcePacketizerReturnsSinglePacket(t *testing.T) {
	packetizer := testSourcePacketizer(func() uint32 { return 1 })
	data := logicalPacket(a2s.ResponseRules, bytes.Repeat([]byte("rule"), 4))

	packets, err := packetizer.Packetize(data)
	if err != nil {
		t.Fatalf("Packetize() error = %v", err)
	}
	if len(packets) != 1 {
		t.Fatalf("packet count = %d, want 1", len(packets))
	}
	if !bytes.Equal(packets[0], data) {
		t.Fatalf("single packet = %X, want %X", packets[0], data)
	}
	if &packets[0][0] == &data[0] {
		t.Fatal("single packet aliases input")
	}
}

func TestSourcePacketizerSplitsLogicalPacket(t *testing.T) {
	const splitSize = 18
	data := logicalPacket(a2s.ResponseRules, bytes.Repeat([]byte("x"), 25))
	packetizer := testSourcePacketizer(func() uint32 { return 0x81234567 })
	packetizer.SplitSize = splitSize

	packets, err := packetizer.Packetize(data)
	if err != nil {
		t.Fatalf("Packetize() error = %v", err)
	}
	if len(packets) != 5 {
		t.Fatalf("packet count = %d, want 5", len(packets))
	}

	var rebuilt []byte
	for index, packet := range packets {
		if len(packet) > splitSize {
			t.Fatalf("packet %d size = %d, want <= %d", index, len(packet), splitSize)
		}
		if got := binary.LittleEndian.Uint32(packet[:4]); got != sourceSplitMarker {
			t.Fatalf("packet %d marker = 0x%X, want 0x%X", index, got, sourceSplitMarker)
		}
		if got := binary.LittleEndian.Uint32(packet[4:8]); got != 0x01234567 {
			t.Fatalf("packet %d ID = 0x%X, want 0x01234567", index, got)
		}
		if packet[8] != byte(len(packets)) || packet[9] != byte(index) {
			t.Fatalf("packet %d fragment metadata = (%d, %d), want (%d, %d)", index, packet[8], packet[9], len(packets), index)
		}
		if got := binary.LittleEndian.Uint16(packet[10:12]); got != splitSize {
			t.Fatalf("packet %d split size = %d, want %d", index, got, splitSize)
		}
		rebuilt = append(rebuilt, packet[sourceSplitHeaderSize:]...)
	}

	if !bytes.Equal(rebuilt, data) {
		t.Fatalf("rebuilt logical packet = %X, want %X", rebuilt, data)
	}
}

func TestSourcePacketizerRejectsLimits(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*SourcePacketizer)
		data      []byte
		want      error
	}{
		{
			name: "invalid input",
			data: []byte("not a packet"),
			want: ErrPacketizerInput,
		},
		{
			name: "split size too small",
			configure: func(packetizer *SourcePacketizer) {
				packetizer.SplitSize = sourceSplitHeaderSize
			},
			data: logicalPacket(a2s.ResponseRules, bytes.Repeat([]byte("x"), 32)),
			want: ErrPacketizerSplitSize,
		},
		{
			name: "response too large",
			configure: func(packetizer *SourcePacketizer) {
				packetizer.MaxResponseSize = 5
			},
			data: logicalPacket(a2s.ResponseRules, []byte("too large")),
			want: ErrPacketizerResponseSize,
		},
		{
			name: "fragment count overflow",
			configure: func(packetizer *SourcePacketizer) {
				packetizer.SplitSize = sourceSplitHeaderSize + 1
			},
			data: logicalPacket(a2s.ResponseRules, bytes.Repeat([]byte("x"), sourceSplitCountMax)),
			want: ErrPacketizerFragmentCount,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			packetizer := testSourcePacketizer(func() uint32 { return 1 })
			if test.configure != nil {
				test.configure(packetizer)
			}

			_, err := packetizer.Packetize(test.data)
			if !errors.Is(err, test.want) {
				t.Fatalf("Packetize() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestNewSourcePacketizer(t *testing.T) {
	packetizer, err := NewSourcePacketizer()
	if err != nil {
		t.Fatalf("NewSourcePacketizer() error = %v", err)
	}
	if packetizer.SplitSize != DefaultSourceSplitSize {
		t.Fatalf("SplitSize = %d, want %d", packetizer.SplitSize, DefaultSourceSplitSize)
	}
	if packetizer.MaxResponseSize != DefaultSourceMaxResponseSize {
		t.Fatalf("MaxResponseSize = %d, want %d", packetizer.MaxResponseSize, DefaultSourceMaxResponseSize)
	}
}

func TestSourcePacketizerZeroValue(t *testing.T) {
	var packetizer SourcePacketizer
	data := logicalPacket(a2s.ResponseRules, bytes.Repeat([]byte("x"), DefaultSourceSplitSize))

	packets, err := packetizer.Packetize(data)
	if err != nil {
		t.Fatalf("Packetize() error = %v", err)
	}
	if len(packets) < 2 {
		t.Fatalf("packet count = %d, want split response", len(packets))
	}
}

func TestSourcePacketizerRoundTripsThroughA2SClient(t *testing.T) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("ListenUDP() error = %v", err)
	}
	defer conn.Close()

	client, err := a2s.NewWithAddr(conn.LocalAddr().(*net.UDPAddr), a2s.WithTimeout(time.Second))
	if err != nil {
		t.Fatalf("NewWithAddr() error = %v", err)
	}
	defer client.Close()

	logical := logicalPacket(a2s.ResponseRules, bytes.Repeat([]byte("payload"), 8))
	packetizer := testSourcePacketizer(func() uint32 { return 0x12345678 })
	packetizer.SplitSize = 18

	serverErr := make(chan error, 1)
	go func() {
		buffer := make([]byte, 2048)
		_, remote, readErr := conn.ReadFromUDP(buffer)
		if readErr != nil {
			serverErr <- readErr
			return
		}

		packets, packetizeErr := packetizer.Packetize(logical)
		if packetizeErr != nil {
			serverErr <- packetizeErr
			return
		}
		for _, packet := range packets {
			if _, writeErr := conn.WriteToUDP(packet, remote); writeErr != nil {
				serverErr <- writeErr
				return
			}
		}
		serverErr <- nil
	}()

	got, _, err := client.Query(context.Background(), a2s.RulesRequest)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if got.Type != a2s.ResponseRules || !bytes.Equal(got.Payload, logical[5:]) {
		t.Fatalf("Query() packet = %#v, want type 0x%X and payload %X", got, a2s.ResponseRules, logical[5:])
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("UDP server error = %v", err)
	}
}

func logicalPacket(responseType a2s.ResponseType, payload []byte) []byte {
	packet := make([]byte, 5+len(payload))
	binary.LittleEndian.PutUint32(packet[:4], ^uint32(0))
	packet[4] = byte(responseType)
	copy(packet[5:], payload)
	return packet
}

func testSourcePacketizer(nextID func() uint32) *SourcePacketizer {
	return &SourcePacketizer{
		SplitSize:       DefaultSourceSplitSize,
		MaxResponseSize: DefaultSourceMaxResponseSize,
		nextID:          nextID,
	}
}
