package a2s

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"math"
	"net"
	"strconv"
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

func singlePacketFixture(response ResponseType, payload []byte) []byte {
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
	packet := singlePacketFixture(ResponsePlayers, []byte{0x01, 0x02})

	multi, err := isMultiPacket(packet)
	if err != nil {
		t.Fatalf("isMultiPacket returned error: %v", err)
	}
	if multi {
		t.Fatal("single packet fixture classified as split packet")
	}
	if got, want := packet[4], byte(ResponsePlayers); got != want {
		t.Fatalf("response type = 0x%X, want 0x%X", got, want)
	}
}

func TestSourceSplitPacketFixture(t *testing.T) {
	assembled := singlePacketFixture(ResponseRules, []byte("split fixture"))
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
	packet := singlePacketFixture(ResponseInfo, []byte("fixture"))
	fixture := newUDPPacketFixture(t, packet)

	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	data, flag, _, err := client.Get(context.Background(), InfoRequest)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if flag != ResponseInfo {
		t.Fatalf("response flag = 0x%X, want 0x%X", flag, ResponseInfo)
	}
	if string(data) != "fixture" {
		t.Fatalf("response payload = %q, want %q", data, "fixture")
	}
}

func TestSplitPacketFixtureUDPFeed(t *testing.T) {
	assembled := singlePacketFixture(ResponseRules, []byte("split fixture"))
	fixture := newUDPPacketFixture(t, sourceSplitPacketSequence(0x1AFEBABE, assembled, 3)...)

	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	data, flag, _, err := client.Get(context.Background(), RulesRequest)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if flag != ResponseRules {
		t.Fatalf("response flag = 0x%X, want 0x%X", flag, ResponseRules)
	}
	if string(data) != "split fixture" {
		t.Fatalf("response payload = %q, want %q", data, "split fixture")
	}
}

func TestSplitPacketFixtureUDPFeedReordered(t *testing.T) {
	assembled := singlePacketFixture(ResponseRules, []byte("reordered split fixture"))
	packets := sourceSplitPacketSequence(0x1AFEBABE, assembled, 3)
	reordered := append(append([][]byte{}, packets[1:]...), packets[0])
	fixture := newUDPPacketFixture(t, reordered...)

	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	data, flag, _, err := client.Get(context.Background(), RulesRequest)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if flag != ResponseRules {
		t.Fatalf("response flag = 0x%X, want 0x%X", flag, ResponseRules)
	}
	if string(data) != "reordered split fixture" {
		t.Fatalf("response payload = %q, want %q", data, "reordered split fixture")
	}
}

func TestCompressedSplitPacketFixtureUDPFeedReordered(t *testing.T) {
	assembled := append(singlePacketFixture(ResponseRules, []byte("compressed split fixture:")),
		bytes.Repeat([]byte("0123456789abcdef"), 20)...)
	compressed := []byte{
		0x42, 0x5a, 0x68, 0x39, 0x31, 0x41, 0x59, 0x26, 0x53, 0x59, 0xf3, 0x7d, 0x36, 0x5d,
		0x00, 0x00, 0xaf, 0x5d, 0x80, 0xc0, 0x00, 0x40, 0x00, 0x7f, 0xf0, 0x02, 0x00, 0x3f, 0x26,
		0xde, 0x40, 0x00, 0x00, 0xa0, 0x00, 0x72, 0x29, 0x30, 0x1a, 0x09, 0x84, 0x19, 0x31, 0x94,
		0x0a, 0x95, 0x40, 0x1e, 0xa1, 0xa7, 0xa8, 0x6c, 0xa0, 0x3d, 0x43, 0xf6, 0x33, 0xad,
		0xe4, 0x37, 0x11, 0xbc, 0x8a, 0x11, 0xc0, 0x8e, 0x24, 0x72, 0x23, 0x22, 0x39, 0x91,
		0x52, 0x2f, 0x5e, 0xa4, 0x66, 0x45, 0x88, 0xb6, 0xcb, 0x91, 0xf3, 0x16, 0x23, 0x4e,
		0xb9, 0x0f, 0x1a, 0x3d, 0x5f, 0x18, 0xa5, 0x30, 0x2b, 0x48, 0x00, 0x90, 0xa0, 0x8f,
		0x1f, 0xc5, 0xdc, 0x91, 0x4e, 0x14, 0x24, 0x3c, 0xdf, 0x4d, 0x97, 0x40,
	}
	if _, err := decompressBzip2(compressed, uint32(len(assembled)), crc32.ChecksumIEEE(assembled)); err != nil {
		t.Fatalf("invalid compressed fixture: %v", err)
	}

	packets := compressedSourceSplitPacketSequence(
		0x81234567,
		compressed,
		uint32(len(assembled)),
		crc32.ChecksumIEEE(assembled),
		9,
	)
	reordered := append(append([][]byte{}, packets[2:]...), packets[:2]...)
	fixture := newUDPPacketFixture(t, reordered...)

	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	data, flag, _, err := client.Get(context.Background(), RulesRequest)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if flag != ResponseRules {
		t.Fatalf("response flag = 0x%X, want 0x%X", flag, ResponseRules)
	}
	if !bytes.Equal(data, assembled[5:]) {
		t.Fatalf("response payload = %q, want %q", data, assembled[5:])
	}
}

func TestGetRejectsImpossibleSplitIndex(t *testing.T) {
	packets := sourceSplitPacketSequence(
		0x12345678,
		singlePacketFixture(ResponseRules, []byte("invalid index")),
		3,
	)
	packets[0][9] = packets[0][8]
	fixture := newUDPPacketFixture(t, packets[0])

	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	if _, _, _, err := client.Get(context.Background(), RulesRequest); !errors.Is(err, ErrMultiPacket) {
		t.Fatalf("Get error = %v, want ErrMultiPacket", err)
	}
}

func TestGetRejectsInconsistentSplitFragment(t *testing.T) {
	assembled := singlePacketFixture(ResponseRules, []byte("inconsistent split fixture"))
	first := sourceSplitPacketSequence(0x12345678, assembled, 8)
	second := sourceSplitPacketSequence(0x12345678, assembled, 3)
	fixture := newUDPPacketFixture(t, first[0], second[1])

	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	if _, _, _, err := client.Get(context.Background(), RulesRequest); !errors.Is(err, ErrMultiPacketInconsistent) {
		t.Fatalf("Get error = %v, want ErrMultiPacketInconsistent", err)
	}
}

func TestGetRejectsConflictingDuplicateSplitFragment(t *testing.T) {
	packets := sourceSplitPacketSequence(
		0x12345678,
		singlePacketFixture(ResponseRules, []byte("duplicate split fixture")),
		3,
	)
	duplicate := append([]byte(nil), packets[0]...)
	duplicate[len(duplicate)-1] ^= 0x01
	fixture := newUDPPacketFixture(t, packets[0], duplicate)

	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	if _, _, _, err := client.Get(context.Background(), RulesRequest); !errors.Is(err, ErrMultiPacketConflict) {
		t.Fatalf("Get error = %v, want ErrMultiPacketConflict", err)
	}
}

func compressedSourceSplitPacketSequence(id uint32, compressed []byte, unpackedSize uint32, crc uint32, chunkSize int) [][]byte {
	if len(compressed) == 0 || chunkSize <= 0 {
		return nil
	}

	count := (len(compressed) + chunkSize - 1) / chunkSize
	packets := make([][]byte, 0, count)
	for index := 0; index < count; index++ {
		start := index * chunkSize
		end := start + chunkSize
		if end > len(compressed) {
			end = len(compressed)
		}

		headerSize := srcSplitHeader
		metadataSize := 0
		if index == 0 {
			metadataSize = 8
		}
		packet := make([]byte, headerSize+metadataSize+end-start)
		binary.LittleEndian.PutUint32(packet[:4], multiPacket)
		binary.LittleEndian.PutUint32(packet[4:8], id)
		packet[8] = byte(count)
		packet[9] = byte(index)
		binary.LittleEndian.PutUint16(packet[10:12], uint16(len(compressed)))
		if index == 0 {
			binary.LittleEndian.PutUint32(packet[12:16], unpackedSize)
			binary.LittleEndian.PutUint32(packet[16:20], crc)
		}
		copy(packet[headerSize+metadataSize:], compressed[start:end])
		packets = append(packets, packet)
	}

	return packets
}

func TestRulesPreserveRawAndParsedValues(t *testing.T) {
	payload := []byte{3, 0}
	for _, entry := range [][2]string{
		{"number", "00123"},
		{"boolean", "true"},
		{"encoded", "c2VydmVyIHRleHQ="},
	} {
		payload = append(payload, entry[0]...)
		payload = append(payload, 0)
		payload = append(payload, entry[1]...)
		payload = append(payload, 0)
	}

	t.Run("raw", func(t *testing.T) {
		fixture := newUDPPacketFixture(t, singlePacketFixture(ResponseRules, payload))
		client, err := NewWithAddr(fixture.Addr())
		if err != nil {
			t.Fatalf("create client: %v", err)
		}
		defer client.Close()

		rules, err := client.GetRules(context.Background())
		if err != nil {
			t.Fatalf("GetRules returned error: %v", err)
		}

		want := map[string]string{
			"number":  "00123",
			"boolean": "true",
			"encoded": "c2VydmVyIHRleHQ=",
		}
		for key, value := range want {
			got, ok := rules.Get(key)
			if !ok || got != value {
				t.Errorf("raw rule %q = %q, want %q", key, got, value)
			}
		}
	})

	t.Run("parsed", func(t *testing.T) {
		fixture := newUDPPacketFixture(t, singlePacketFixture(ResponseRules, payload))
		client, err := NewWithAddr(fixture.Addr())
		if err != nil {
			t.Fatalf("create client: %v", err)
		}
		defer client.Close()

		rules, err := client.GetParsedRules(context.Background())
		if err != nil {
			t.Fatalf("GetParsedRules returned error: %v", err)
		}

		if got, want := rules["number"], int64(123); got != want {
			t.Errorf("parsed number = %#v, want %#v", got, want)
		}
		if got, want := rules["boolean"], true; got != want {
			t.Errorf("parsed boolean = %#v, want %#v", got, want)
		}
		if got, want := rules["encoded"], "server text"; got != want {
			t.Errorf("parsed Base64 value = %#v, want %#v", got, want)
		}
	})
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

func TestGetRejectsTruncatedChallengeFixtures(t *testing.T) {
	for length := 5; length <= 8; length++ {
		t.Run(strconv.Itoa(length), func(t *testing.T) {
			packet := make([]byte, length)
			binary.LittleEndian.PutUint32(packet[:4], singlePacket)
			packet[4] = byte(ResponseChallenge)
			fixture := newUDPPacketFixture(t, packet)

			client, err := NewWithAddr(fixture.Addr())
			if err != nil {
				t.Fatalf("create client: %v", err)
			}
			defer client.Close()

			_, _, _, err = client.Get(context.Background(), RulesRequest)
			if err == nil {
				t.Fatal("Get returned nil error for truncated challenge")
			}
			if !errors.Is(err, ErrChallengeRead) {
				t.Fatalf("error = %v, want ErrChallengeRead", err)
			}
			if !errors.Is(err, ErrInsufficientData) {
				t.Fatalf("error = %v, want ErrInsufficientData", err)
			}
		})
	}
}

func TestGetAcceptsCompleteChallengeFixture(t *testing.T) {
	challenge := singlePacketFixture(ResponseChallenge, []byte{0x01, 0x02, 0x03, 0x04})
	response := singlePacketFixture(ResponseRules, []byte("rules"))
	fixture := newUDPPacketFixture(t, challenge, response)

	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	data, flag, _, err := client.Get(context.Background(), RulesRequest)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if flag != ResponseRules {
		t.Fatalf("response flag = 0x%X, want 0x%X", flag, ResponseRules)
	}
	if string(data) != "rules" {
		t.Fatalf("response payload = %q, want %q", data, "rules")
	}
}

func TestGetPlayersParsesSignedScores(t *testing.T) {
	payload := []byte{3}
	for index, score := range []int32{-1, 0, math.MaxInt32} {
		payload = append(payload, byte(index))
		payload = append(payload, "player"...)
		payload = append(payload, byte('0'+index), 0)
		payload = binary.LittleEndian.AppendUint32(payload, uint32(score))
		payload = binary.LittleEndian.AppendUint32(payload, math.Float32bits(1))
	}

	fixture := newUDPPacketFixture(t, singlePacketFixture(ResponsePlayers, payload))
	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	players, err := client.GetPlayers(context.Background())
	if err != nil {
		t.Fatalf("GetPlayers returned error: %v", err)
	}
	if got, want := len(players), 3; got != want {
		t.Fatalf("player count = %d, want %d", got, want)
	}

	want := []int32{-1, 0, math.MaxInt32}
	for index, player := range players {
		if player.Score != want[index] {
			t.Errorf("player %d score = %d, want %d", index, player.Score, want[index])
		}
	}
}

func TestGetTheShipPlayersParsesSignedScores(t *testing.T) {
	payload := []byte{1, 0}
	payload = append(payload, "player"...)
	payload = append(payload, 0)
	score := int32(-1)
	payload = binary.LittleEndian.AppendUint32(payload, uint32(score))
	payload = binary.LittleEndian.AppendUint32(payload, math.Float32bits(1))
	payload = binary.LittleEndian.AppendUint32(payload, 2)
	payload = binary.LittleEndian.AppendUint32(payload, 3)

	fixture := newUDPPacketFixture(t, singlePacketFixture(ResponsePlayers, payload))
	client, err := NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer client.Close()

	players, err := client.GetTheShipPlayers(context.Background())
	if err != nil {
		t.Fatalf("GetTheShipPlayers returned error: %v", err)
	}
	if got, want := players[0].Score, int32(-1); got != want {
		t.Fatalf("The Ship player score = %d, want %d", got, want)
	}
}
