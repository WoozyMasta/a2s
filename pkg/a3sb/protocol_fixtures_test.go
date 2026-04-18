package a3sb

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"testing"

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
