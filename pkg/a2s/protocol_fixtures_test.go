package a2s

import (
	"encoding/binary"
	"net"
	"testing"
)

// udpPacketFixture feeds a fixed sequence of datagrams to one local UDP query.
type udpPacketFixture struct {
	conn *net.UDPConn
	done chan struct{}
}

func newUDPPacketFixture(t *testing.T, packets ...[]byte) *udpPacketFixture {
	t.Helper()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen on UDP fixture: %v", err)
	}

	fixture := &udpPacketFixture{
		conn: conn,
		done: make(chan struct{}),
	}

	go func() {
		defer close(fixture.done)

		buffer := make([]byte, 64*1024)
		_, address, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}

		for _, packet := range packets {
			if _, err := conn.WriteToUDP(packet, address); err != nil {
				return
			}
		}
	}()

	t.Cleanup(func() {
		_ = conn.Close()
		<-fixture.done
	})

	return fixture
}

func (f *udpPacketFixture) Addr() *net.UDPAddr {
	address := f.conn.LocalAddr().(*net.UDPAddr)
	return &net.UDPAddr{
		IP:   append(net.IP(nil), address.IP...),
		Port: address.Port,
		Zone: address.Zone,
	}
}

func singlePacketFixture(response Flag, payload []byte) []byte {
	packet := make([]byte, 5+len(payload))
	binary.LittleEndian.PutUint32(packet[:4], singlePacket)
	packet[4] = byte(response)
	copy(packet[5:], payload)
	return packet
}

func sourceSplitPacketSequence(id uint32, payload []byte, chunkSize int) [][]byte {
	if len(payload) == 0 || chunkSize <= 0 {
		return nil
	}

	count := (len(payload) + chunkSize - 1) / chunkSize
	packets := make([][]byte, 0, count)
	for index := 0; index < count; index++ {
		start := index * chunkSize
		end := start + chunkSize
		if end > len(payload) {
			end = len(payload)
		}

		packet := make([]byte, 12+end-start)
		binary.LittleEndian.PutUint32(packet[:4], multiPacket)
		binary.LittleEndian.PutUint32(packet[4:8], id)
		packet[8] = byte(count)
		packet[9] = byte(index)
		binary.LittleEndian.PutUint16(packet[10:12], uint16(len(payload)))
		copy(packet[12:], payload[start:end])
		packets = append(packets, packet)
	}

	return packets
}

func TestSinglePacketFixture(t *testing.T) {
	packet := singlePacketFixture(playerResponse, []byte{0x01, 0x02})

	multi, err := isMultiPacket(packet)
	if err != nil {
		t.Fatalf("isMultiPacket returned error: %v", err)
	}
	if multi {
		t.Fatal("single packet fixture classified as split packet")
	}
	if got, want := packet[4], byte(playerResponse); got != want {
		t.Fatalf("response type = 0x%X, want 0x%X", got, want)
	}
}

func TestSourceSplitPacketFixture(t *testing.T) {
	assembled := singlePacketFixture(rulesResponse, []byte("split fixture"))
	packets := sourceSplitPacketSequence(0x12345678, assembled, 4)
	if len(packets) != 5 {
		t.Fatalf("packet count = %d, want 5", len(packets))
	}

	info, err := parseSplitHeader(packets[0])
	if err != nil {
		t.Fatalf("parseSplitHeader returned error: %v", err)
	}
	if info.goldSrc {
		t.Fatal("Source fixture classified as GoldSource")
	}
	if info.count != len(packets) || info.index != 0 {
		t.Fatalf("split metadata = count %d, index %d; want count %d, index 0", info.count, info.index, len(packets))
	}
	if info.id != 0x12345678 {
		t.Fatalf("split ID = 0x%X, want 0x12345678", info.id)
	}
}

func TestPacketFixtureUDPFeed(t *testing.T) {
	packet := singlePacketFixture(infoResponseSource, []byte("fixture"))
	fixture := newUDPPacketFixture(t, packet)

	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	data, flag, _, err := client.Get(InfoRequest)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if flag != infoResponseSource {
		t.Fatalf("response flag = 0x%X, want 0x%X", flag, infoResponseSource)
	}
	if string(data) != "fixture" {
		t.Fatalf("response payload = %q, want %q", data, "fixture")
	}
}

func TestSplitPacketFixtureUDPFeed(t *testing.T) {
	assembled := singlePacketFixture(rulesResponse, []byte("split fixture"))
	fixture := newUDPPacketFixture(t, sourceSplitPacketSequence(0x1AFEBABE, assembled, 3)...)

	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	data, flag, _, err := client.Get(RulesRequest)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if flag != rulesResponse {
		t.Fatalf("response flag = 0x%X, want 0x%X", flag, rulesResponse)
	}
	if string(data) != "split fixture" {
		t.Fatalf("response payload = %q, want %q", data, "split fixture")
	}
}

func TestMalformedPacketFixtures(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want error
	}{
		{name: "empty", data: nil, want: ErrMultiPacket},
		{name: "truncated single header", data: []byte{0xFF, 0xFF, 0xFF, 0xFF}, want: ErrSinglePacket},
		{name: "truncated split header", data: []byte{0xFE, 0xFF, 0xFF, 0xFF, 0x01, 0x00, 0x00, 0x00}, want: ErrMultiPacket},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := isMultiPacket(test.data)
			if err != test.want {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}
