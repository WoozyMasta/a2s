package server

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"runtime"
	"sync"

	"github.com/woozymasta/a2s/pkg/a2s"
)

const defaultMaxRequestSize = 2048

// DefaultMaxRequestSize is the largest request datagram accepted by Server.
// One extra byte is read internally to detect truncation.
const DefaultMaxRequestSize = defaultMaxRequestSize

// Server receives A2S UDP requests and dispatches them to a handler.
//
// Handler calls may run concurrently. Configure a Server with New;
// the zero value is not ready to serve.
// Serve does not close an externally supplied PacketConn.
type Server struct {
	// Handler is the configured handler wrapped by the challenge gate.
	Handler Handler

	packetizer     *SourcePacketizer // Encodes logical responses into UDP datagrams.
	workers        int               // Fixed number of concurrent UDP workers.
	maxRequestSize int               // Maximum accepted request datagram size.
}

type serverConfig struct {
	policy         ChallengePolicy   // Challenge requirement policy.
	provider       ChallengeProvider // Challenge token issuer and validator.
	packetizer     *SourcePacketizer // Source response packetizer.
	workers        int               // Fixed worker count.
	maxRequestSize int               // Maximum accepted request size.
}

// Option configures a Server during construction.
type Option func(*serverConfig) error

// New creates a UDP A2S server with secure challenge handling by default.
func New(handler Handler, options ...Option) (*Server, error) {
	if handler == nil {
		return nil, fmt.Errorf("%w: handler is nil", ErrServer)
	}

	config := serverConfig{
		workers:        DefaultWorkerCount(),
		maxRequestSize: DefaultMaxRequestSize,
		policy:         SecureChallengePolicy(),
	}
	for _, option := range options {
		if option == nil {
			return nil, fmt.Errorf("%w: nil option", ErrServer)
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}

	if config.provider == nil {
		provider, err := NewStatelessChallengeProvider()
		if err != nil {
			return nil, err
		}
		config.provider = provider
	}
	if config.packetizer == nil {
		packetizer, err := NewSourcePacketizer()
		if err != nil {
			return nil, err
		}
		config.packetizer = packetizer
	}

	gate, err := NewChallengeGate(handler, config.policy, config.provider)
	if err != nil {
		return nil, err
	}

	return &Server{
		Handler:        gate,
		workers:        config.workers,
		maxRequestSize: config.maxRequestSize,
		packetizer:     config.packetizer,
	}, nil
}

// DefaultWorkerCount returns the default number of concurrent UDP workers.
func DefaultWorkerCount() int {
	workers := runtime.GOMAXPROCS(0)
	if workers < 1 {
		return 1
	}

	return workers
}

// WithWorkers sets the fixed number of UDP workers used by Serve.
func WithWorkers(workers int) Option {
	return func(config *serverConfig) error {
		if workers < 1 {
			return fmt.Errorf("%w: workers must be positive", ErrServer)
		}

		config.workers = workers
		return nil
	}
}

// WithMaxRequestSize sets the maximum accepted UDP request size.
func WithMaxRequestSize(size int) Option {
	return func(config *serverConfig) error {
		if size < 1 {
			return fmt.Errorf("%w: max request size must be positive", ErrServer)
		}

		config.maxRequestSize = size
		return nil
	}
}

// WithChallengePolicy replaces the default secure challenge policy.
func WithChallengePolicy(policy ChallengePolicy) Option {
	return func(config *serverConfig) error {
		if policy == nil {
			return fmt.Errorf("%w: challenge policy is nil", ErrServer)
		}

		config.policy = policy
		return nil
	}
}

// WithChallengeProvider replaces the default stateless challenge provider.
func WithChallengeProvider(provider ChallengeProvider) Option {
	return func(config *serverConfig) error {
		if provider == nil {
			return fmt.Errorf("%w: challenge provider is nil", ErrServer)
		}

		config.provider = provider
		return nil
	}
}

// WithSourcePacketizer replaces the default Source packetizer.
func WithSourcePacketizer(packetizer *SourcePacketizer) Option {
	return func(config *serverConfig) error {
		if packetizer == nil {
			return fmt.Errorf("%w: packetizer is nil", ErrServer)
		}

		config.packetizer = packetizer
		return nil
	}
}

// ListenAndServe listens on addr and serves A2S UDP requests until the socket
// is closed or an unrecoverable read error occurs.
func (s *Server) ListenAndServe(addr string) error {
	if s == nil {
		return fmt.Errorf("%w: server is nil", ErrServer)
	}

	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return fmt.Errorf("%w: listen on %q: %w", ErrServer, addr, err)
	}
	defer func() {
		_ = conn.Close()
	}()

	return s.Serve(conn)
}

// Serve reads and handles UDP datagrams using a fixed worker pool.
// The supplied PacketConn remains owned by the caller and is not closed.
func (s *Server) Serve(conn net.PacketConn) error {
	if s == nil {
		return fmt.Errorf("%w: server is nil", ErrServer)
	}
	if conn == nil {
		return fmt.Errorf("%w: packet connection is nil", ErrServer)
	}
	if s.Handler == nil || s.packetizer == nil || s.workers < 1 || s.maxRequestSize < 1 {
		return fmt.Errorf("%w: server is not initialized", ErrServer)
	}

	var workers sync.WaitGroup
	errorsCh := make(chan error, s.workers)
	workers.Add(s.workers)
	for range s.workers {
		go func() {
			defer workers.Done()
			s.serveWorker(conn, errorsCh)
		}()
	}

	workers.Wait()
	close(errorsCh)
	for err := range errorsCh {
		return err
	}

	return nil
}

// serveWorker handles datagrams until ReadFrom reports a connection error.
func (s *Server) serveWorker(conn net.PacketConn, errorsCh chan<- error) {
	buffer := make([]byte, s.maxRequestSize+1)
	for {
		n, remote, err := conn.ReadFrom(buffer)
		if err != nil {
			errorsCh <- err
			return
		}
		if n > s.maxRequestSize {
			continue
		}

		request, err := a2s.DecodeRequest(buffer[:n])
		if err != nil {
			continue
		}
		remoteAddr, err := addrPort(remote)
		if err != nil {
			continue
		}

		serverRequest := &Request{Remote: remoteAddr, Query: request}
		response, err := s.Handler.Handle(context.Background(), serverRequest)
		if err != nil || response == nil {
			continue
		}

		packet, err := NormalizeResponse(serverRequest, response)
		if err != nil {
			continue
		}
		logical, err := a2s.AppendPacket(nil, packet)
		if err != nil {
			continue
		}
		datagrams, err := s.packetizer.Packetize(logical)
		if err != nil {
			continue
		}
		for _, datagram := range datagrams {
			if _, err := conn.WriteTo(datagram, remote); err != nil {
				continue
			}
		}
	}
}

// addrPort converts the standard UDP address returned by PacketConn
// into the canonical server endpoint used by handlers and challenge providers.
func addrPort(remote net.Addr) (netip.AddrPort, error) {
	if remote == nil {
		return netip.AddrPort{}, fmt.Errorf("%w: remote address is nil", ErrServer)
	}

	udpAddr, ok := remote.(*net.UDPAddr)
	if ok && udpAddr != nil {
		if udpAddr.Port < 0 || udpAddr.Port > 65535 {
			return netip.AddrPort{}, fmt.Errorf("%w: invalid remote port", ErrServer)
		}

		addr, valid := netip.AddrFromSlice(udpAddr.IP)
		if !valid {
			return netip.AddrPort{}, fmt.Errorf("%w: invalid remote IP", ErrServer)
		}

		return netip.AddrPortFrom(addr.Unmap(), uint16(udpAddr.Port)), nil
	}

	parsed, err := netip.ParseAddrPort(remote.String())
	if err != nil {
		return netip.AddrPort{}, fmt.Errorf("%w: unsupported remote address %T: %v", ErrServer, remote, err)
	}

	return netip.AddrPortFrom(parsed.Addr().Unmap(), parsed.Port()), nil
}
