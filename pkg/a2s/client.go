package a2s

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

// Client handles UDP connection and A2S protocol queries.
// Queries on one Client are serialized for the lifetime of each transaction.
type Client struct {
	conn       *net.UDPConn  // UDP connection to the server.
	address    *net.UDPAddr  // Server network address.
	readBuf    []byte        // Reusable UDP read buffer.
	timeout    time.Duration // UDP read deadline.
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
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

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
// response type, ping duration and error.
// Automatically handles challenge-response if server requires it.
func (c *Client) Get(requestType Flag) ([]byte, Flag, time.Duration, error) {
	if c == nil {
		return nil, 0, 0, ErrClientClosed
	}
	c.queryMu.Lock()
	defer c.queryMu.Unlock()

	if c.conn == nil {
		return nil, 0, 0, ErrClientClosed
	}

	var (
		lastUnexpectedErr  error
		lastUnexpectedFlag Flag
		lastDuration       time.Duration
	)

	for attempt := 0; attempt < 3; attempt++ {
		// Retry the complete request when a server returns an unsupported response
		// or repeatedly fails the challenge exchange.
		resp, duration, err := c.request(requestType, singlePacket)
		if err != nil {
			if lastUnexpectedErr != nil {
				return nil, lastUnexpectedFlag, lastDuration, errors.Join(lastUnexpectedErr, err)
			}
			return nil, 0, 0, err
		}

		flag := Flag(resp[4])
		retryAfterChallengeError := false

		for challengeAttempt := 0; challengeAttempt < 4 && flag == challengeResponse; challengeAttempt++ {
			if len(resp) < 9 {
				return nil, challengeResponse, duration, fmt.Errorf(
					"%w: %w (got %d bytes, want at least 9)",
					ErrChallengeRead,
					ErrInsufficientData,
					len(resp),
				)
			}

			challenge := binary.BigEndian.Uint32(resp[5:9])
			resp, _, err = c.request(requestType, challenge)
			if err != nil {
				challengeErr := errors.Join(validationErrForRequest(requestType), ErrChallengeLoop, err)
				if requestType == RulesRequest || requestType == PlayerRequest {
					lastUnexpectedErr = challengeErr
					lastUnexpectedFlag = challengeResponse
					lastDuration = duration
					retryAfterChallengeError = true
					break
				}

				return nil, challengeResponse, duration, challengeErr
			}
			flag = Flag(resp[4])
		}

		if retryAfterChallengeError {
			continue
		}

		// If response type is not valid, classify error as ErrQueryUnsupported and continue.
		if err := validateResponseType(requestType, flag); err != nil {
			classified := err
			switch {
			case flag == challengeResponse:
				classified = errors.Join(err, ErrChallengeLoop)
			case requestType != InfoRequest && (flag == infoResponseSource || flag == infoResponseGoldSource):
				classified = errors.Join(err, ErrQueryUnsupported)
			}

			if requestType != InfoRequest &&
				(flag == challengeResponse ||
					flag == infoResponseSource ||
					flag == infoResponseGoldSource) {
				lastUnexpectedErr = classified
				lastUnexpectedFlag = flag
				lastDuration = duration
				continue
			}

			return resp[5:], flag, duration, classified
		}

		return resp[5:], flag, duration, nil
	}

	if lastUnexpectedErr != nil {
		return nil, lastUnexpectedFlag, lastDuration, lastUnexpectedErr
	}

	return nil, 0, 0, validationErrForRequest(requestType)
}

// validationErrForRequest returns an error for an unsupported request type.
func validationErrForRequest(requestType Flag) error {
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

// request creates header, sends request and returns response with ping duration.
// Handles multi-packet responses by collecting and assembling packets.
func (c *Client) request(requestType Flag, challenge uint32) ([]byte, time.Duration, error) {
	req, err := createHeader(requestType, challenge)
	if err != nil {
		return nil, 0, err
	}

	start := time.Now()

	if _, err := c.conn.Write(req); err != nil {
		return nil, 0, err
	}
	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, 0, err
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
			return nil, 0, err
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

	duration := time.Since(start)

	if !readOK {
		if packetErr != nil {
			result := make([]byte, n)
			copy(result, resp[:n])
			return result, 0, packetErr
		}
		return nil, 0, ErrSinglePacket
	}

	multi, err := isMultiPacket(resp[:n])
	if err != nil {
		result := make([]byte, n)
		copy(result, resp[:n])
		return result, 0, err
	}

	if !multi {
		result := make([]byte, n)
		copy(result, resp[:n])
		return result, duration, nil
	}

	// Multi-packet response: extract metadata from the first packet,
	// then collect the remaining packets with the same response identifier.
	info, err := parseSplitHeader(resp[:n])
	if err != nil {
		return nil, 0, err
	}

	if info.count > splitPacketCountMax || info.index < 0 || info.index >= info.count {
		return nil, 0, ErrMultiPacket
	}

	packets := make([][]byte, info.count)
	received := 1
	if n < info.dataOff {
		return nil, 0, ErrMultiPacket
	}
	firstPacketData := make([]byte, n-info.dataOff)
	copy(firstPacketData, resp[info.dataOff:n])
	packets[info.index] = firstPacketData

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
			return nil, 0, err
		}

		if n < splitMin {
			continue
		}

		header := binary.LittleEndian.Uint32(resp[:4])
		if header != multiPacket {
			continue
		}

		if binary.LittleEndian.Uint32(resp[4:8]) != info.id {
			continue
		}

		// Packet belongs to current split response but is too short for its header.
		// Treat as malformed response instead of waiting for read timeout.
		if n < info.headerSize {
			return nil, 0, ErrMultiPacket
		}

		currentPacket := info.readPacketNumber(resp[:n])
		if currentPacket < 0 || currentPacket >= info.count {
			continue
		}

		if packets[currentPacket] == nil {
			packetData := make([]byte, n-info.headerSize)
			copy(packetData, resp[info.headerSize:n])
			packets[currentPacket] = packetData
			received++
		}
	}

	// Calculate total size and assemble packets in order
	totalSize := 0
	for i := 0; i < info.count; i++ {
		if packets[i] == nil {
			return nil, 0, ErrMultiPacketMismatch
		}
		totalSize += len(packets[i])
	}

	assembledResp := make([]byte, 0, totalSize)
	for _, data := range packets {
		assembledResp = append(assembledResp, data...)
	}

	if info.compressed {
		decompressed, err := decompressBzip2(assembledResp, info.unpackedSize, info.crc)
		if err != nil {
			return nil, 0, err
		}
		return decompressed, duration, nil
	}

	return assembledResp, duration, nil
}
