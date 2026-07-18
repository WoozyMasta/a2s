package server

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

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

	packetizer     Packetizer    // Encodes logical responses into UDP datagrams.
	panicReporter  PanicReporter // Receives recovered handler panics.
	run            *serverRun    // Active serve loop, if any.
	workers        int           // Fixed number of concurrent UDP workers.
	maxRequestSize int           // Maximum accepted request datagram size.
	lifecycleMu    sync.Mutex    // Serializes server lifecycle transitions.
}

type serverConfig struct {
	policy         ChallengePolicy   // Challenge requirement policy.
	provider       ChallengeProvider // Challenge token issuer and validator.
	packetizer     Packetizer        // Response packetizer.
	panicReporter  PanicReporter     // Receives recovered handler panics.
	workers        int               // Fixed worker count.
	maxRequestSize int               // Maximum accepted request size.
}

// serverRun owns the state of one active ServeContext invocation.
type serverRun struct {
	ctx      context.Context    // Context shared by active handler calls.
	conn     net.PacketConn     // PacketConn owned by the caller.
	reason   error              // First reason the serve loop was stopped.
	cancel   context.CancelFunc // Cancels the active serve loop and handlers.
	done     chan struct{}      // Closed after all workers and cleanup finish.
	stopOnce sync.Once          // Makes shutdown signaling idempotent.
	reasonMu sync.Mutex         // Protects reason during concurrent shutdown.
}

// stop records the first termination reason, cancels handlers, and wakes reads.
func (r *serverRun) stop(reason error) {
	r.stopOnce.Do(func() {
		r.reasonMu.Lock()
		r.reason = reason
		r.reasonMu.Unlock()
		r.cancel()
		// PacketConn has no context-aware ReadFrom. A temporary deadline is
		// the portable way to wake workers without taking ownership of conn.
		_ = r.conn.SetReadDeadline(time.Now())
	})
}

// stopReason returns the first recorded termination reason.
func (r *serverRun) stopReason() error {
	r.reasonMu.Lock()
	defer r.reasonMu.Unlock()

	return r.reason
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
		panicReporter:  config.panicReporter,
		workers:        config.workers,
		maxRequestSize: config.maxRequestSize,
		packetizer:     config.packetizer,
	}, nil
}

// WithPanicReporter configures the callback used for recovered handler panics.
// Passing nil disables reporting and is safe.
func WithPanicReporter(reporter PanicReporter) Option {
	return func(config *serverConfig) error {
		config.panicReporter = reporter
		return nil
	}
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
	if packetizer == nil {
		return func(*serverConfig) error {
			return fmt.Errorf("%w: packetizer is nil", ErrServer)
		}
	}

	return WithPacketizer(packetizer)
}

// WithGoldSourcePacketizer configures legacy GoldSource response framing.
func WithGoldSourcePacketizer(packetizer *GoldSourcePacketizer) Option {
	if packetizer == nil {
		return func(*serverConfig) error {
			return fmt.Errorf("%w: packetizer is nil", ErrServer)
		}
	}

	return WithPacketizer(packetizer)
}

// WithPacketizer replaces the default response packetizer.
func WithPacketizer(packetizer Packetizer) Option {
	return func(config *serverConfig) error {
		if isNilPacketizer(packetizer) {
			return fmt.Errorf("%w: packetizer is nil", ErrServer)
		}

		config.packetizer = packetizer
		return nil
	}
}

// isNilPacketizer handles both a nil interface and an interface containing a typed nil pointer.
func isNilPacketizer(packetizer Packetizer) bool {
	if packetizer == nil {
		return true
	}

	value := reflect.ValueOf(packetizer)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
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
//
// Use Shutdown to stop Serve.
// The supplied PacketConn remains owned by the caller and is not closed.
// Serve returns ErrServerClosed after Shutdown.
func (s *Server) Serve(conn net.PacketConn) error {
	return s.ServeContext(context.Background(), conn)
}

// ServeContext reads and handles UDP datagrams until the context is canceled,
// Shutdown is called, or the PacketConn reports an unexpected error.
//
// The supplied PacketConn remains owned by the caller and is not closed.
// During cancellation the server temporarily sets its read deadline
// to wake blocked workers, then clears that deadline before returning.
func (s *Server) ServeContext(ctx context.Context, conn net.PacketConn) error {
	if s == nil {
		return fmt.Errorf("%w: server is nil", ErrServer)
	}
	if ctx == nil {
		return fmt.Errorf("%w: context is nil", ErrServer)
	}
	if conn == nil {
		return fmt.Errorf("%w: packet connection is nil", ErrServer)
	}
	if s.Handler == nil || s.packetizer == nil || s.workers < 1 || s.maxRequestSize < 1 {
		return fmt.Errorf("%w: server is not initialized", ErrServer)
	}

	serveCtx, cancel := context.WithCancel(ctx)
	run := &serverRun{
		ctx:    serveCtx,
		cancel: cancel,
		conn:   conn,
		done:   make(chan struct{}),
	}

	s.lifecycleMu.Lock()
	if s.run != nil {
		s.lifecycleMu.Unlock()
		cancel()
		return fmt.Errorf("%w: %w", ErrServer, ErrServerRunning)
	}
	s.run = run
	s.lifecycleMu.Unlock()

	defer func() {
		cancel()
		_ = conn.SetReadDeadline(time.Time{})
		s.lifecycleMu.Lock()
		if s.run == run {
			s.run = nil
		}
		s.lifecycleMu.Unlock()
		close(run.done)
	}()

	go func() {
		select {
		case <-ctx.Done():
			run.stop(ctx.Err())
		case <-run.done:
		}
	}()
	if err := ctx.Err(); err != nil {
		run.stop(err)
	}

	var workers sync.WaitGroup
	workers.Add(s.workers)
	for range s.workers {
		go func() {
			defer workers.Done()
			s.serveWorker(run)
		}()
	}

	workers.Wait()
	if run.stopReason() == nil {
		if err := ctx.Err(); err != nil {
			run.stop(err)
		} else {
			run.stop(ErrServer)
		}
	}

	return run.stopReason()
}

// Shutdown stops the active serve loop and waits for its workers to exit.
// It is safe to call Shutdown more than once. A nil context is rejected.
func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("%w: server is nil", ErrServer)
	}
	if ctx == nil {
		return fmt.Errorf("%w: context is nil", ErrServer)
	}

	s.lifecycleMu.Lock()
	run := s.run
	s.lifecycleMu.Unlock()
	if run == nil {
		return nil
	}

	run.stop(ErrServerClosed)
	select {
	case <-run.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// serveWorker handles datagrams until the serve context is canceled or
// ReadFrom reports a connection error.
func (s *Server) serveWorker(run *serverRun) {
	conn := run.conn
	buffer := make([]byte, s.maxRequestSize+1)
	for {
		n, remote, err := conn.ReadFrom(buffer)
		if err != nil {
			if run.ctx.Err() == nil {
				run.stop(err)
			}
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
		response, err := s.handle(run.ctx, serverRequest)
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

// handle invokes the configured handler and converts a panic
// into a dropped response after notifying the optional reporter.
func (s *Server) handle(ctx context.Context, request *Request) (response Response, err error) {
	defer func() {
		if value := recover(); value != nil {
			s.reportPanic(ctx, PanicReport{
				Request: request,
				Value:   value,
				Stack:   debug.Stack(),
			})
			response = nil
			err = nil
		}
	}()

	return s.Handler.Handle(ctx, request)
}

// reportPanic invokes the configured reporter
// without allowing reporter failures to terminate the serving worker.
func (s *Server) reportPanic(ctx context.Context, report PanicReport) {
	if s.panicReporter == nil {
		return
	}

	defer func() {
		_ = recover()
	}()
	s.panicReporter(ctx, report)
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
