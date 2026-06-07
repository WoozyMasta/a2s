package a2s

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	// maxUnsupportedResponses bounds retries after an unexpected response type.
	maxUnsupportedResponses = 3

	// maxChallengeResponses bounds challenge responses in one query transaction.
	maxChallengeResponses = 4
)

// Client handles UDP connection and A2S protocol queries.
// Queries on one Client are serialized for the lifetime of each transaction.
type Client struct {
	conn       *net.UDPConn  // UDP connection to the server.
	address    *net.UDPAddr  // Server network address.
	querySem   chan struct{} // Context-aware query serialization.
	readBuf    []byte        // Reusable UDP read buffer.
	timeout    time.Duration // UDP read deadline.
	timeoutMu  sync.RWMutex  // Protects timeout changes and reads.
	queryMu    sync.Mutex    // Serializes queries and lifecycle changes.
	bufferSize uint16        // Maximum UDP datagram size to read.
}

// Option configures a Client before its UDP connection is opened.
type Option func(*Client) error

// New creates a client for a host and port and opens its UDP connection.
func New(host string, port int, opts ...Option) (*Client, error) {
	if host == "" || port < 1 || port > 65535 {
		return nil, ErrInvalidAddress
	}

	return NewWithString(net.JoinHostPort(host, strconv.Itoa(port)), opts...)
}

// NewWithString creates a client from a host:port address and opens its UDP connection.
func NewWithString(address string, opts ...Option) (*Client, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, fmt.Errorf("%w %q: %v", ErrInvalidAddress, address, err)
	}

	return NewWithAddr(udpAddr, opts...)
}

// NewWithAddr creates a client for a resolved address and opens its UDP connection.
func NewWithAddr(addr *net.UDPAddr, opts ...Option) (*Client, error) {
	if err := validateAddress(addr); err != nil {
		return nil, err
	}

	client := &Client{
		address:    cloneAddress(addr),
		timeout:    DefaultDeadlineTimeout,
		bufferSize: DefaultBufferSize,
		readBuf:    make([]byte, DefaultBufferSize),
		querySem:   make(chan struct{}, 1),
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	conn, err := net.DialUDP("udp", nil, client.address)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", client.address, err)
	}

	client.conn = conn
	return client, nil
}

// WithTimeout sets the UDP read deadline used by the client.
func WithTimeout(timeout time.Duration) Option {
	return func(client *Client) error {
		if timeout <= 0 {
			return ErrInvalidTimeout
		}

		client.timeout = timeout
		return nil
	}
}

// WithBufferSize sets the maximum UDP datagram size read by the client.
func WithBufferSize(size uint16) Option {
	return func(client *Client) error {
		return client.setBufferSize(size)
	}
}

// Addr returns a copy of the server network address.
func (c *Client) Addr() *net.UDPAddr {
	if c == nil {
		return nil
	}

	return cloneAddress(c.address)
}

// BufferSize returns the maximum UDP datagram size read by the client.
func (c *Client) BufferSize() uint16 {
	if c == nil {
		return 0
	}
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	return c.bufferSize
}

// Timeout returns the UDP read deadline.
func (c *Client) Timeout() time.Duration {
	if c == nil {
		return 0
	}
	c.timeoutMu.RLock()
	defer c.timeoutMu.RUnlock()

	return c.timeout
}

// SetBufferSize sets the maximum UDP datagram size read by the client.
func (c *Client) SetBufferSize(size uint16) error {
	if c == nil {
		return ErrClientClosed
	}
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	return c.setBufferSize(size)
}

func (c *Client) setBufferSize(size uint16) error {
	if size == 0 {
		return ErrInvalidBufferSize
	}

	c.bufferSize = size
	if cap(c.readBuf) < int(size) {
		c.readBuf = make([]byte, size)
	} else {
		c.readBuf = c.readBuf[:size]
	}

	return nil
}

// SetTimeout sets the UDP read deadline.
func (c *Client) SetTimeout(timeout time.Duration) error {
	if c == nil {
		return ErrClientClosed
	}
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	if timeout <= 0 {
		return ErrInvalidTimeout
	}

	c.timeoutMu.Lock()
	defer c.timeoutMu.Unlock()

	c.timeout = timeout
	return nil
}

// Close closes the UDP connection. It is safe to call multiple times.
func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	return err
}

func validateAddress(addr *net.UDPAddr) error {
	if addr == nil || addr.IP == nil || addr.IP.IsUnspecified() || addr.Port == 0 {
		return ErrInvalidAddress
	}

	return nil
}

func cloneAddress(addr *net.UDPAddr) *net.UDPAddr {
	if addr == nil {
		return nil
	}

	clone := *addr
	clone.IP = append(net.IP(nil), addr.IP...)
	return &clone
}

// Get sends request and returns response data (without header),
// response type, complete query duration and error.
//
// Automatically handles challenge-response if server requires it.
// Context covers complete query transaction, including retries,
// challenge exchange, and split-packet assembly.
func (c *Client) Get(ctx context.Context, requestType QueryType) ([]byte, ResponseType, time.Duration, error) {
	if c == nil {
		return nil, 0, 0, ErrClientClosed
	}
	if ctx == nil {
		return nil, 0, 0, ErrNilContext
	}

	effectiveCtx, cancel := c.effectiveContext(ctx)
	defer cancel()
	if err := c.acquireQuery(effectiveCtx); err != nil {
		return nil, 0, 0, err
	}
	defer c.releaseQuery()

	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	if c.conn == nil {
		return nil, 0, 0, ErrClientClosed
	}

	started := time.Now()

	var (
		lastUnexpectedErr      error
		lastUnexpectedResponse ResponseType
	)

	for attempt := 0; attempt < maxUnsupportedResponses; attempt++ {
		resp, responseType, err := c.requestWithChallenge(effectiveCtx, requestType)
		duration := time.Since(started)
		if err != nil {
			if lastUnexpectedErr != nil {
				return nil, lastUnexpectedResponse, duration, errors.Join(lastUnexpectedErr, err)
			}

			return nil, responseType, duration, err
		}

		// If response type is not valid, classify error as ErrQueryUnsupported and continue.
		if err := validateResponseType(requestType, responseType); err != nil {
			classified := err
			switch {
			case responseType == ResponseChallenge:
				classified = errors.Join(err, ErrChallengeLoop)

			case requestType != InfoRequest &&
				(responseType == ResponseInfo || responseType == ResponseInfoGoldSource):
				classified = errors.Join(err, ErrQueryUnsupported)
			}

			if requestType != InfoRequest &&
				(responseType == ResponseChallenge ||
					responseType == ResponseInfo ||
					responseType == ResponseInfoGoldSource) {
				lastUnexpectedErr = classified
				lastUnexpectedResponse = responseType
				continue
			}

			return resp[5:], responseType, duration, classified
		}

		return resp[5:], responseType, duration, nil
	}

	if lastUnexpectedErr != nil {
		return nil, lastUnexpectedResponse, time.Since(started), lastUnexpectedErr
	}

	return nil, 0, time.Since(started), validationErrForRequest(requestType)
}

// effectiveContext applies the client timeout
// only when the caller did not provide a deadline of its own.
func (c *Client) effectiveContext(ctx context.Context) (context.Context, context.CancelFunc) {
	c.timeoutMu.RLock()
	timeout := c.timeout
	c.timeoutMu.RUnlock()

	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}

// acquireQuery reserves the client for one complete query transaction.
func (c *Client) acquireQuery(ctx context.Context) error {
	if c.querySem == nil {
		return ErrClientClosed
	}

	select {
	case c.querySem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// releaseQuery releases the client query reservation.
func (c *Client) releaseQuery() {
	<-c.querySem
}

// requestWithChallenge executes one request transaction,
// including its bounded challenge exchange.
// ChallengeRequest returns its challenge
// as the final response and must never enter this exchange.
func (c *Client) requestWithChallenge(ctx context.Context, requestType QueryType) ([]byte, ResponseType, error) {
	challenge := singlePacket

	for attempt := 0; attempt < maxChallengeResponses; attempt++ {
		resp, err := c.request(ctx, requestType, challenge)
		if err != nil {
			return nil, 0, err
		}

		responseType := ResponseType(resp[4])
		if responseType != ResponseChallenge || requestType == ChallengeRequest || requestType == PingRequest {
			return resp, responseType, nil
		}

		if attempt == maxChallengeResponses-1 {
			return resp, ResponseChallenge, ErrChallengeLoop
		}

		challenge, err = parseChallengeResponse(resp)
		if err != nil {
			return resp, ResponseChallenge, err
		}
	}

	return nil, ResponseChallenge, ErrChallengeLoop
}

// parseChallengeResponse reads a challenge from a complete A2S response.
func parseChallengeResponse(data []byte) (uint32, error) {
	if len(data) < 9 {
		return 0, fmt.Errorf(
			"%w: %w (got %d bytes, want at least 9)",
			ErrChallengeRead,
			ErrInsufficientData,
			len(data),
		)
	}

	return binary.LittleEndian.Uint32(data[5:9]), nil
}

// validationErrForRequest returns an error for an unsupported request type.
func validationErrForRequest(requestType QueryType) error {
	switch requestType {
	case InfoRequest:
		return ErrValidatorInfo

	case PlayerRequest:
		return ErrValidatorPlayer

	case RulesRequest:
		return ErrValidatorRules

	case PingRequest:
		return ErrValidatorPing

	case ChallengeRequest:
		return ErrValidatorChallenge

	default:
		return ErrValidatorRequest
	}
}

// request creates header, sends request and returns a complete response.
// Handles multi-packet responses by collecting and assembling packets.
func (c *Client) request(ctx context.Context, requestType QueryType, challenge uint32) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req, err := createHeader(requestType, challenge)
	if err != nil {
		return nil, err
	}

	if _, err := c.conn.Write(req); err != nil {
		return nil, contextError(ctx, err)
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.Timeout())
	}
	if err := c.conn.SetReadDeadline(deadline); err != nil {
		return nil, err
	}

	var (
		resp      []byte
		n         int
		packetErr error
	)
	readOK := false

	for attempt := 0; attempt < 6; attempt++ {
		if cap(c.readBuf) < int(c.bufferSize) {
			c.readBuf = make([]byte, c.bufferSize)
		}
		resp = c.readBuf[:c.bufferSize]
		n, err = c.conn.Read(resp)
		if err != nil {
			return nil, contextError(ctx, err)
		}

		_, packetErr = isMultiPacket(resp[:n])
		if packetErr != nil {
			if errors.Is(packetErr, ErrMultiPacket) ||
				errors.Is(packetErr, ErrSinglePacket) ||
				errors.Is(packetErr, ErrValidatorHeader) {
				continue // Ignore truncated or unrelated datagrams and keep reading.
			}
			break
		}

		readOK = true
		break
	}

	if !readOK {
		if packetErr != nil {
			result := make([]byte, n)
			copy(result, resp[:n])
			return result, packetErr
		}
		return nil, ErrSinglePacket
	}

	multi, err := isMultiPacket(resp[:n])
	if err != nil {
		result := make([]byte, n)
		copy(result, resp[:n])
		return result, err
	}

	if !multi {
		result := make([]byte, n)
		copy(result, resp[:n])
		return result, nil
	}

	// Multi-packet response: classify the first datagram, then collect raw/ fragments.
	// Compression metadata is parsed from fragment zero
	// after all fragments have been identified, because UDP may reorder their arrival.
	info, err := parseSplitHeader(resp[:n])
	if err != nil {
		return nil, err
	}

	if info.count > splitPacketCountMax || info.index < 0 || info.index >= info.count {
		return nil, ErrMultiPacket
	}
	if n > splitResponseSizeMax {
		return nil, ErrMultiPacketSize
	}

	packets := make([][]byte, info.count)
	received := 1
	receivedSize := n
	firstPacket := make([]byte, n)
	copy(firstPacket, resp[:n])
	packets[info.index] = firstPacket

	// Collect remaining packets.
	// Unrelated datagrams are ignored because UDP does not guarantee
	// that the next datagram belongs to this request.
	for received < info.count {
		if cap(c.readBuf) < int(c.bufferSize) {
			c.readBuf = make([]byte, c.bufferSize)
		}

		resp = c.readBuf[:c.bufferSize]
		n, err := c.conn.Read(resp)
		if err != nil {
			return nil, contextError(ctx, err)
		}

		if n < splitMin {
			continue
		}

		if binary.LittleEndian.Uint32(resp[4:8]) != info.id {
			continue
		}

		packetInfo, err := parseSplitHeader(resp[:n])
		if err != nil {
			return nil, errors.Join(ErrMultiPacketInconsistent, err)
		}
		if err := validateSplitFragment(info, packetInfo); err != nil {
			return nil, err
		}

		currentPacket := packetInfo.index
		if currentPacket < 0 || currentPacket >= info.count {
			continue
		}

		if packets[currentPacket] == nil {
			if n > splitResponseSizeMax-receivedSize {
				return nil, ErrMultiPacketSize
			}

			packet := make([]byte, n)
			copy(packet, resp[:n])
			packets[currentPacket] = packet
			receivedSize += n
			received++
		} else if !bytes.Equal(packets[currentPacket], resp[:n]) {
			return nil, ErrMultiPacketConflict
		}
	}

	// Fragment zero owns compression metadata
	// and determines the payload offset for the complete response.
	// Parse it only after reassembly has all indexes.
	firstInfo, err := parseSplitHeader(packets[0])
	if err != nil || firstInfo.index != 0 {
		return nil, ErrMultiPacket
	}
	info = firstInfo

	// Calculate total size and assemble packets in protocol order.
	totalSize := 0
	for i := 0; i < info.count; i++ {
		if packets[i] == nil {
			return nil, ErrMultiPacketMismatch
		}

		packetInfo, err := parseSplitHeader(packets[i])
		if err != nil {
			return nil, errors.Join(ErrMultiPacketInconsistent, err)
		}
		if err := validateSplitFragment(info, packetInfo); err != nil || packetInfo.index != i {
			return nil, ErrMultiPacketInconsistent
		}

		dataOff := info.headerSize
		if i == 0 {
			dataOff = info.dataOff
		}
		if len(packets[i]) < dataOff {
			return nil, ErrMultiPacket
		}

		packetSize := len(packets[i]) - dataOff
		if packetSize > splitResponseSizeMax-totalSize {
			return nil, ErrMultiPacketSize
		}

		totalSize += packetSize
	}

	assembledResp := make([]byte, 0, totalSize)
	for i, packet := range packets {
		dataOff := info.headerSize
		if i == 0 {
			dataOff = info.dataOff
		}
		assembledResp = append(assembledResp, packet[dataOff:]...)
	}

	if info.compressed {
		decompressed, err := decompressBzip2(assembledResp, info.unpackedSize, info.crc)
		if err != nil {
			return nil, err
		}
		return decompressed, nil
	}

	return assembledResp, nil
}

// contextError prefers cancellation or deadline errors over socket timeout errors
// so callers can reliably use errors.Is with context errors.
func contextError(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}

	return err
}
