package server

import (
	"context"
	"encoding/binary"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
)

func BenchmarkDecodeRequest(b *testing.B) {
	data, err := a2s.AppendRequest(nil, a2s.Request{Type: a2s.PingRequest})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := a2s.DecodeRequest(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChallengeValidate(b *testing.B) {
	provider, err := NewStatelessChallengeProvider()
	if err != nil {
		b.Fatal(err)
	}
	remote := netip.MustParseAddrPort("127.0.0.1:27015")
	challenge := provider.Issue(remote)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !provider.Validate(remote, challenge) {
			b.Fatal("challenge validation failed")
		}
	}
}

func BenchmarkEncodeInfo(b *testing.B) {
	info := benchmarkInfo()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := a2s.AppendInfo(nil, info); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeRules(b *testing.B) {
	rules := benchmarkRules(64)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := a2s.AppendRules(nil, rules); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSourcePacketization(b *testing.B) {
	logical := benchmarkLogicalRules(64)
	packetizer := &SourcePacketizer{
		SplitSize:       DefaultSourceSplitSize,
		MaxResponseSize: DefaultSourceMaxResponseSize,
		nextID:          func() uint32 { return 1 },
	}
	b.SetBytes(int64(len(logical)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := packetizer.Packetize(logical); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStateHandle(b *testing.B) {
	state := NewState()
	if err := state.Store(Snapshot{
		Info:    ptrToInfo(benchmarkInfo()),
		Players: benchmarkPlayers(16),
		Rules:   benchmarkRules(64),
	}); err != nil {
		b.Fatal(err)
	}
	request := &Request{Query: a2s.Request{Type: a2s.RulesRequest}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := state.Handle(context.Background(), request); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLocalUDPInfo(b *testing.B) {
	info := benchmarkInfo()
	client := newBenchmarkClient(b, HandlerFunc(func(_ context.Context, request *Request) (Response, error) {
		if request.Query.Type != a2s.InfoRequest {
			return nil, ErrDrop
		}

		return InfoResponse{Info: info}, nil
	}))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.GetInfo(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLocalUDPLargeRules(b *testing.B) {
	rules := benchmarkRules(256)
	client := newBenchmarkClient(b, HandlerFunc(func(_ context.Context, request *Request) (Response, error) {
		if request.Query.Type != a2s.RulesRequest {
			return nil, ErrDrop
		}

		return RulesResponse{Rules: rules}, nil
	}))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := client.GetRules(context.Background()); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkInfo() a2s.Info {
	return a2s.Info{
		Format:      a2s.InfoFormat(a2s.ResponseInfo),
		Name:        "benchmark server",
		Map:         "benchmark_map",
		Folder:      "benchmark",
		Game:        "Benchmark Game",
		Version:     "1.0.0",
		AppID:       1234,
		Protocol:    17,
		Players:     8,
		MaxPlayers:  64,
		Bots:        1,
		ServerType:  a2s.ServerType('d'),
		Environment: a2s.Environment('l'),
		VAC:         true,
	}
}

func benchmarkPlayers(count int) []a2s.Player {
	players := make([]a2s.Player, count)
	for index := range players {
		players[index] = a2s.Player{
			Name:     "benchmark_player",
			Index:    byte(index),
			Score:    int32(index),
			Duration: time.Duration(index+1) * time.Minute,
		}
	}

	return players
}

func benchmarkRules(count int) a2s.Rules {
	rules := make(a2s.Rules, count)
	for index := range rules {
		rules[index] = a2s.Rule{
			Name:  "rule_" + string(rune('a'+index%26)),
			Value: "benchmark_value_0123456789",
		}
	}

	return rules
}

func benchmarkLogicalRules(count int) []byte {
	payload, err := a2s.AppendRules(nil, benchmarkRules(count))
	if err != nil {
		panic(err)
	}

	logical := make([]byte, 5+len(payload))
	binary.LittleEndian.PutUint32(logical[:4], 0xFFFFFFFF)
	logical[4] = byte(a2s.ResponseRules)
	copy(logical[5:], payload)
	return logical
}

func newBenchmarkClient(b *testing.B, handler Handler) *a2s.Client {
	b.Helper()

	server, err := New(
		handler,
		WithChallengePolicy(NoChallengePolicy()),
		WithWorkers(1),
	)
	if err != nil {
		b.Fatal(err)
	}

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(conn) }()

	client, err := a2s.NewWithAddr(conn.LocalAddr().(*net.UDPAddr), a2s.WithTimeout(time.Second))
	if err != nil {
		_ = conn.Close()
		<-serveErr
		b.Fatal(err)
	}

	b.Cleanup(func() {
		_ = client.Close()
		_ = server.Shutdown(context.Background())
		_ = conn.Close()
		if err := <-serveErr; err != ErrServerClosed {
			b.Errorf("benchmark server error = %v, want %v", err, ErrServerClosed)
		}
	})

	return client
}

func ptrToInfo(info a2s.Info) *a2s.Info {
	return &info
}
