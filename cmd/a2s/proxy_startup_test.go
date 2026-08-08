package main

import (
	"context"
	"encoding/binary"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a2s/server"
)

func TestPrepareProxyStartupSelectsPolicyPacketizerAndCapabilities(t *testing.T) {
	tests := []struct {
		name             string
		infoType         a2s.ResponseType
		infoChallenge    bool
		wantSecurePolicy bool
		wantPacketizer   any
	}{
		{
			name:             "source without info challenge",
			infoType:         a2s.ResponseInfo,
			wantPacketizer:   &server.SourcePacketizer{},
			wantSecurePolicy: false,
		},
		{
			name:             "goldsource with info challenge",
			infoType:         a2s.ResponseInfoGoldSource,
			infoChallenge:    true,
			wantPacketizer:   &server.GoldSourcePacketizer{},
			wantSecurePolicy: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newProxyStartupFixture(t, tt.infoType, tt.infoChallenge, true, true)
			command := proxyStartupCommand(fixture, []string{"auto"})

			preparation, err := prepareProxyStartup(context.Background(), command)
			if err != nil {
				t.Fatalf("prepareProxyStartup() error = %v", err)
			}
			defer preparation.close()

			if preparation.infoMeta.UsedChallenge != tt.infoChallenge {
				t.Fatalf(
					"UsedChallenge = %v, want %v",
					preparation.infoMeta.UsedChallenge,
					tt.infoChallenge,
				)
			}
			if preparation.policy.Required(a2s.InfoRequest) != tt.wantSecurePolicy {
				t.Fatalf("INFO policy = %v, want %v", preparation.policy.Required(a2s.InfoRequest), tt.wantSecurePolicy)
			}
			if !preparation.policy.Required(a2s.PlayerRequest) || !preparation.policy.Required(a2s.RulesRequest) {
				t.Fatal("legacy and secure policies must require PLAYER/RULES challenges")
			}
			if _, ok := preparation.packetizer.(interface {
				Packetize([]byte) ([][]byte, error)
			}); !ok {
				t.Fatalf("packetizer does not implement Packetize: %T", preparation.packetizer)
			}
			switch tt.wantPacketizer.(type) {
			case *server.SourcePacketizer:
				if _, ok := preparation.packetizer.(*server.SourcePacketizer); !ok {
					t.Fatalf("packetizer = %T, want SourcePacketizer", preparation.packetizer)
				}
			case *server.GoldSourcePacketizer:
				if _, ok := preparation.packetizer.(*server.GoldSourcePacketizer); !ok {
					t.Fatalf("packetizer = %T, want GoldSourcePacketizer", preparation.packetizer)
				}
			}

			for _, query := range []a2s.QueryType{a2s.InfoRequest, a2s.PlayerRequest, a2s.RulesRequest} {
				if !preparation.cache.Enabled(query) {
					t.Fatalf("cache does not enable %s", proxyQueryName(query))
				}
				if _, ok := preparation.cache.Load(query); !ok {
					t.Fatalf("cache has no startup packet for %s", proxyQueryName(query))
				}
			}
		})
	}
}

func TestPrepareProxyStartupSkipsUnsupportedAutoQueries(t *testing.T) {
	fixture := newProxyStartupFixture(t, a2s.ResponseInfo, false, true, false)
	command := proxyStartupCommand(fixture, []string{"auto"})
	command.Timeout = 10 * time.Millisecond

	preparation, err := prepareProxyStartup(context.Background(), command)
	if err != nil {
		t.Fatalf("prepareProxyStartup() error = %v", err)
	}
	defer preparation.close()

	if !preparation.cache.Enabled(a2s.InfoRequest) {
		t.Fatal("auto cache did not enable INFO")
	}
	if preparation.cache.Enabled(a2s.PlayerRequest) {
		t.Fatal("auto cache enabled unsupported PLAYER query")
	}
	if preparation.cache.Enabled(a2s.RulesRequest) {
		t.Fatal("auto cache enabled unsupported RULES query")
	}
}

func TestPrepareProxyStartupFailsBeforeServingOnInfoError(t *testing.T) {
	fixture := newProxyStartupFixture(t, a2s.ResponseInfo, false, false, false)
	command := proxyStartupCommand(fixture, []string{"info"})
	command.Timeout = 10 * time.Millisecond

	if _, err := prepareProxyStartup(context.Background(), command); err == nil {
		t.Fatal("prepareProxyStartup() returned nil error for unavailable INFO")
	}
}

func proxyStartupCommand(fixture *proxyStartupFixture, cache []string) *ProxyCommand {
	return &ProxyCommand{
		Args: ServerArgs{
			Host: fixture.addr.IP.String(),
			Port: strconv.Itoa(fixture.addr.Port),
		},
		Listen:  ":27016",
		Cache:   cache,
		Timeout: 250 * time.Millisecond,
		Buffer:  8192,
	}
}

type proxyStartupFixture struct {
	conn          *net.UDPConn
	addr          *net.UDPAddr
	info          []byte
	infoChallenge bool
	respondInfo   bool
	respondExtra  bool
	done          chan struct{}
}

func newProxyStartupFixture(
	t *testing.T,
	infoType a2s.ResponseType,
	infoChallenge bool,
	respondInfo bool,
	respondExtra bool,
) *proxyStartupFixture {
	t.Helper()

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatalf("ListenUDP() error = %v", err)
	}

	info, err := proxyStartupInfoPacket(infoType)
	if err != nil {
		_ = conn.Close()
		t.Fatalf("proxyStartupInfoPacket() error = %v", err)
	}

	fixture := &proxyStartupFixture{
		conn:          conn,
		addr:          conn.LocalAddr().(*net.UDPAddr),
		info:          info,
		infoChallenge: infoChallenge,
		respondInfo:   respondInfo,
		respondExtra:  respondExtra,
		done:          make(chan struct{}),
	}
	go fixture.serve()

	t.Cleanup(func() {
		_ = conn.Close()
		<-fixture.done
	})

	return fixture
}

func (f *proxyStartupFixture) serve() {
	defer close(f.done)

	buffer := make([]byte, 64*1024)
	infoChallenged := false
	for {
		n, remote, err := f.conn.ReadFromUDP(buffer)
		if err != nil {
			return
		}
		if n < 5 {
			continue
		}

		query := a2s.QueryType(buffer[4])
		if query == a2s.InfoRequest && f.infoChallenge && !infoChallenged {
			infoChallenged = true
			_, _ = f.conn.WriteToUDP(proxyStartupChallengePacket(), remote)
			continue
		}

		var response []byte
		switch query {
		case a2s.InfoRequest:
			if f.respondInfo {
				response = f.info
			}

		case a2s.PlayerRequest:
			if f.respondExtra {
				response, _ = a2s.AppendPlayers(nil, nil)
			}

		case a2s.RulesRequest:
			if f.respondExtra {
				response, _ = a2s.AppendRules(nil, nil)
			}
		}

		if len(response) > 0 {
			_, _ = f.conn.WriteToUDP(response, remote)
		}
	}
}

func proxyStartupInfoPacket(responseType a2s.ResponseType) ([]byte, error) {
	info := a2s.Info{
		Format:      a2s.InfoFormat(responseType),
		Name:        "proxy test",
		Map:         "test_map",
		Folder:      "test_game",
		Game:        "Test Game",
		Version:     "1.0",
		ServerType:  a2s.ServerType('d'),
		Environment: a2s.Environment('l'),
	}
	if responseType == a2s.ResponseInfoGoldSource {
		info.Address = "127.0.0.1:27015"
	}

	return a2s.AppendInfo(nil, info)
}

func proxyStartupChallengePacket() []byte {
	packet := make([]byte, 0, 9)
	packet = binary.LittleEndian.AppendUint32(packet, ^uint32(0))
	packet = append(packet, byte(a2s.ResponseChallenge))
	packet = append(packet, 1, 2, 3, 4)

	return packet
}
