package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"net"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func TestTransparentLogicalPacketProxy(t *testing.T) {
	rules := make(a2s.Rules, 0, 32)
	for index := 0; index < 32; index++ {
		rules = append(rules, a2s.Rule{
			Name:  "rule_" + string(rune('a'+index)),
			Value: strings.Repeat("value", 20),
		})
	}

	upstreamProvider := &proxyChallengeProvider{token: a2s.Challenge{1, 2, 3, 4}}
	upstreamHandler := HandlerFunc(func(_ context.Context, request *Request) (Response, error) {
		if request.Query.Type != a2s.RulesRequest {
			return nil, ErrDrop
		}
		return RulesResponse{Rules: rules}, nil
	})
	upstreamServer, upstreamConn, upstreamErr := newProxyServer(
		t,
		upstreamHandler,
		upstreamProvider,
		64,
	)
	defer upstreamConn.Close()
	defer upstreamServer.Shutdown(context.Background())

	upstreamClient, err := a2s.NewWithAddr(
		upstreamConn.LocalAddr().(*net.UDPAddr),
		a2s.WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatalf("NewWithAddr(upstream) error = %v", err)
	}
	defer upstreamClient.Close()

	expected, _, err := upstreamClient.Query(context.Background(), a2s.RulesRequest)
	if err != nil {
		t.Fatalf("upstream Query() error = %v", err)
	}
	if got := upstreamConn.splitWrites(); got == 0 {
		t.Fatal("upstream response was not split")
	}
	upstreamSplitsBeforeProxy := upstreamConn.splitWrites()

	downstreamProvider := &proxyChallengeProvider{token: a2s.Challenge{5, 6, 7, 8}}
	downstreamHandler := HandlerFunc(func(ctx context.Context, _ *Request) (Response, error) {
		packet, _, err := upstreamClient.Query(ctx, a2s.RulesRequest)
		if err != nil {
			return nil, err
		}

		return PacketResponse{Packet: packet}, nil
	})
	downstreamServer, downstreamConn, downstreamErr := newProxyServer(
		t,
		downstreamHandler,
		downstreamProvider,
		4096,
	)
	defer downstreamConn.Close()
	defer downstreamServer.Shutdown(context.Background())

	downstreamClient, err := a2s.NewWithAddr(
		downstreamConn.LocalAddr().(*net.UDPAddr),
		a2s.WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatalf("NewWithAddr(downstream) error = %v", err)
	}
	defer downstreamClient.Close()

	got, _, err := downstreamClient.Query(context.Background(), a2s.RulesRequest)
	if err != nil {
		t.Fatalf("downstream Query() error = %v", err)
	}
	if got.Type != expected.Type || !bytes.Equal(got.Payload, expected.Payload) {
		t.Fatalf(
			"proxied packet differs: got type 0x%X payload %d bytes, want type 0x%X payload %d bytes",
			got.Type,
			len(got.Payload),
			expected.Type,
			len(expected.Payload),
		)
	}
	if got := upstreamConn.splitWrites(); got <= upstreamSplitsBeforeProxy {
		t.Fatal("proxy did not issue an independent upstream query")
	}
	if got := downstreamConn.splitWrites(); got != 0 {
		t.Fatalf("downstream unexpectedly split logical packet into %d datagrams", got)
	}
	if upstreamProvider.issued.Load() == 0 || downstreamProvider.issued.Load() == 0 {
		t.Fatalf(
			"challenge providers were not both used: upstream=%d downstream=%d",
			upstreamProvider.issued.Load(),
			downstreamProvider.issued.Load(),
		)
	}
	if upstreamProvider.token == downstreamProvider.token {
		t.Fatal("proxy challenge providers unexpectedly share a token")
	}

	if err := downstreamServer.Shutdown(context.Background()); err != nil {
		t.Fatalf("downstream Shutdown() error = %v", err)
	}
	if err := <-downstreamErr; !errors.Is(err, ErrServerClosed) {
		t.Fatalf("downstream Serve() error = %v, want ErrServerClosed", err)
	}
	if err := upstreamServer.Shutdown(context.Background()); err != nil {
		t.Fatalf("upstream Shutdown() error = %v", err)
	}
	if err := <-upstreamErr; !errors.Is(err, ErrServerClosed) {
		t.Fatalf("upstream Serve() error = %v, want ErrServerClosed", err)
	}
}

type proxyChallengeProvider struct {
	token  a2s.Challenge
	issued atomic.Int32
}

func (p *proxyChallengeProvider) Issue(netip.AddrPort) a2s.Challenge {
	p.issued.Add(1)
	return p.token
}

func (p *proxyChallengeProvider) Validate(_ netip.AddrPort, challenge a2s.Challenge) bool {
	return challenge == p.token
}

type recordingPacketConn struct {
	net.PacketConn
	mu     sync.Mutex
	writes [][]byte
}

func (c *recordingPacketConn) WriteTo(data []byte, address net.Addr) (int, error) {
	c.mu.Lock()
	c.writes = append(c.writes, bytes.Clone(data))
	c.mu.Unlock()

	return c.PacketConn.WriteTo(data, address)
}

func (c *recordingPacketConn) splitWrites() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	count := 0
	for _, data := range c.writes {
		if len(data) >= 4 && binary.LittleEndian.Uint32(data[:4]) == 0xFFFFFFFE {
			count++
		}
	}

	return count
}

func newProxyServer(
	t *testing.T,
	handler Handler,
	provider ChallengeProvider,
	splitSize int,
) (*Server, *recordingPacketConn, <-chan error) {
	t.Helper()

	serverInstance, err := New(
		handler,
		WithChallengeProvider(provider),
		WithSourcePacketizer(&SourcePacketizer{SplitSize: splitSize}),
		WithWorkers(1),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error = %v", err)
	}
	recorded := &recordingPacketConn{PacketConn: conn}
	serveErr := make(chan error, 1)
	go func() { serveErr <- serverInstance.Serve(recorded) }()

	return serverInstance, recorded, serveErr
}
