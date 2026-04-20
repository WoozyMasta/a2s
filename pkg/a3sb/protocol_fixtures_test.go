package a3sb

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"testing"

	"github.com/woozymasta/a2s/internal/bread"
	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/steam/utils/appid"
)

type rulesUDPFixture struct {
	conn *net.UDPConn
	done chan struct{}
}

func newRulesUDPFixture(t *testing.T, packets ...[]byte) *rulesUDPFixture {
	t.Helper()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("listen on UDP fixture: %v", err)
	}

	fixture := &rulesUDPFixture{
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

func (f *rulesUDPFixture) Addr() *net.UDPAddr {
	address := f.conn.LocalAddr().(*net.UDPAddr)
	return &net.UDPAddr{
		IP:   append(net.IP(nil), address.IP...),
		Port: address.Port,
		Zone: address.Zone,
	}
}

func rulesSinglePacketFixture(payload []byte) []byte {
	packet := make([]byte, 5+len(payload))
	binary.LittleEndian.PutUint32(packet[:4], 0xFFFFFFFF)
	packet[4] = 0x45
	copy(packet[5:], payload)
	return packet
}

type rulesFixtureEntry struct {
	key   []byte
	value []byte
}

func rulesResponsePayload(entries ...rulesFixtureEntry) []byte {
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, uint16(len(entries)))

	for _, entry := range entries {
		payload = append(payload, entry.key...)
		payload = append(payload, 0)
		payload = append(payload, entry.value...)
		payload = append(payload, 0)
	}

	return payload
}

func TestMinimalDayZProtocolFixture(t *testing.T) {
	// v2, zero flags, no DLC, no mods, and no signatures.
	data := []byte{2, 0, 0, 0, 0, 0}
	rules := &Rules{id: appid.DayZ.Uint64()}

	if err := rules.readA3SB(data); err != nil {
		t.Fatalf("readA3SB returned error: %v", err)
	}
	if rules.Version != 2 {
		t.Fatalf("protocol version = %d, want 2", rules.Version)
	}
}

func TestGetRulesRejectsTruncatedCountFixtures(t *testing.T) {
	for length := 0; length <= 3; length++ {
		t.Run(strconv.Itoa(length), func(t *testing.T) {
			fixture := newRulesUDPFixture(t, rulesSinglePacketFixture(make([]byte, length)))

			baseClient, err := a2s.NewWithAddr(fixture.Addr())
			if err != nil {
				t.Fatalf("create a2s client: %v", err)
			}
			defer baseClient.Close()

			client := &Client{Client: baseClient}
			_, err = client.GetRules(appid.DayZ.Uint64())
			if err == nil {
				t.Fatal("GetRules returned nil error for truncated count")
			}
			if length <= 1 && !errors.Is(err, ErrRules) {
				t.Fatalf("error = %v, want ErrRules", err)
			}
		})
	}
}

func TestGetRulesPreservesNonPageRuleKeys(t *testing.T) {
	fixtureData := rulesResponsePayload(
		rulesFixtureEntry{value: []byte("blank")},
		rulesFixtureEntry{key: []byte{0x01}, value: []byte("one")},
		rulesFixtureEntry{
			key:   []byte{0x01, 0x01},
			value: []byte{2, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02, 0x01, 0x02},
		},
		rulesFixtureEntry{key: []byte{0x02, 0x03, 0x04}, value: []byte("long")},
	)
	fixture := newRulesUDPFixture(t, rulesSinglePacketFixture(fixtureData))

	baseClient, err := a2s.NewWithAddr(fixture.Addr())
	if err != nil {
		t.Fatalf("create a2s client: %v", err)
	}
	defer baseClient.Close()

	client := &Client{Client: baseClient}
	rules, err := client.GetRules(appid.DayZ.Uint64())
	if err != nil {
		t.Fatalf("GetRules returned error: %v", err)
	}

	if got := rules.ExtraRules[string([]byte{0x01})]; got != "one" {
		t.Fatalf("one-byte raw rule = %q, want %q", got, "one")
	}
	if got := rules.ExtraRules[string([]byte{0x02, 0x03, 0x04})]; got != "long" {
		t.Fatalf("long raw rule = %q, want %q", got, "long")
	}
	if _, ok := rules.ExtraRules[""]; ok {
		t.Fatal("blank rule key should not be preserved as a raw rule")
	}
	if stats := rules.GetReaderStats(); stats[1] != 1 {
		t.Fatalf("page count stat = %d, want 1", stats[1])
	}
}

func TestReadDifficultyConsumesFixedWidthField(t *testing.T) {
	marker := byte(0xAA)
	tests := []struct {
		name       string
		data       []byte
		want       *Difficulty
		wantErr    bool
		wantMarker bool
	}{
		{
			name:       "zero bytes",
			data:       []byte{0x00, 0x00, marker},
			wantMarker: true,
		},
		{
			name:       "zero value with nonzero second byte",
			data:       []byte{0x00, 0x01, marker},
			wantMarker: true,
		},
		{
			name: "normal values",
			data: []byte{0xC9, 0x01, marker},
			want: &Difficulty{
				Level:         1,
				AILevel:       1,
				AdvanceFlight: false,
				ThirdPerson:   true,
				Crosshair:     true,
			},
			wantMarker: true,
		},
		{
			name:    "truncated zero first byte",
			data:    []byte{0x00},
			wantErr: true,
		},
		{
			name:    "truncated normal first byte",
			data:    []byte{0xC9},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rules := &Rules{id: appid.Arma3.Uint64()}
			reader := bread.NewReader(test.data)

			err := rules.readDifficulty(reader)
			if test.wantErr {
				if err == nil {
					t.Fatal("readDifficulty returned nil error for truncated field")
				}
				if !errors.Is(err, bread.ErrUnderflow) {
					t.Fatalf("error = %v, want bread.ErrUnderflow", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("readDifficulty returned error: %v", err)
			}
			if (rules.Difficulty == nil) != (test.want == nil) {
				t.Fatalf("difficulty = %+v, want %+v", rules.Difficulty, test.want)
			}
			if rules.Difficulty != nil && *rules.Difficulty != *test.want {
				t.Fatalf("difficulty = %+v, want %+v", *rules.Difficulty, *test.want)
			}
			if test.wantMarker {
				if got, err := reader.Byte(); err != nil || got != marker {
					t.Fatalf("marker read = 0x%X, %v; want 0x%X", got, err, marker)
				}
			}
		})
	}
}

func TestReadDLCOrdersBitsAndHashesDeterministically(t *testing.T) {
	const mask = uint16(0x2201)

	hashData := make([]byte, 12)
	binary.LittleEndian.PutUint32(hashData[0:4], 0x11111111)
	binary.LittleEndian.PutUint32(hashData[4:8], 0x22222222)
	binary.LittleEndian.PutUint32(hashData[8:12], 0x33333333)

	for iteration := 0; iteration < 100; iteration++ {
		rules := &Rules{id: appid.Arma3.Uint64()}
		if err := rules.readDLC(bread.NewReader(hashData), mask); err != nil {
			t.Fatalf("readDLC returned error: %v", err)
		}

		if len(rules.DLC) != 3 {
			t.Fatalf("DLC count = %d, want 3", len(rules.DLC))
		}

		want := []struct {
			name string
			id   uint64
			hash uint32
		}{
			{name: "Karts", id: 288520, hash: 0x11111111},
			{name: "Tanks", id: 798390, hash: 0x22222222},
			{name: "Unknown DLC 8192", id: 0, hash: 0x33333333},
		}
		for index, expected := range want {
			got := rules.DLC[index]
			if got.Name != expected.name || got.ID != expected.id || got.Hash != expected.hash {
				t.Fatalf("iteration %d DLC[%d] = %+v, want %+v", iteration, index, got, expected)
			}
		}
	}
}
