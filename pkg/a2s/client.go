package a2s

import (
	"encoding/binary"
	"errors"
	"net"
	"time"
)

// Client handles UDP connection and A2S protocol queries.
type Client struct {
	Conn       *net.UDPConn
	Address    *net.UDPAddr
	packetsBuf map[int][]byte
	parseData  []byte
	readBuf    []byte
	Timeout    time.Duration
	BufferSize uint16
}

// New creates a new client with IP and port and opens UDP connection.
func New(ip string, port int) (*Client, error) {
	return NewWithAddr(&net.UDPAddr{IP: net.ParseIP(ip), Port: port})
}

// NewWithString creates a new client from "ip:port" string and opens UDP connection.
func NewWithString(addr string) (*Client, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	return NewWithAddr(udpAddr)
}

// NewWithAddr creates a new client with address and opens UDP connection.
func NewWithAddr(addr *net.UDPAddr) (*Client, error) {
	client, err := Create(addr)
	if err != nil {
		return nil, err
	}

	if err := client.Dial(); err != nil {
		return nil, err
	}

	return client, nil
}

// Create creates a client without opening connection. Use Dial() to establish connection.
func Create(addr *net.UDPAddr) (*Client, error) {
	return &Client{
		Address:    addr,
		Timeout:    DefaultDeadlineTimeout * time.Second,
		BufferSize: DefaultBufferSize,
		readBuf:    make([]byte, DefaultBufferSize),
		packetsBuf: make(map[int][]byte, 8),
		parseData:  make([]byte, 0, 4096),
	}, nil
}

// Dial establishes UDP connection to the server.
func (c *Client) Dial() error {
	conn, err := net.DialUDP("udp", nil, c.Address)
	if err != nil {
		return err
	}

	c.Conn = conn
	return nil
}

// SetBufferSize sets read buffer size. Default is 1400 bytes.
func (c *Client) SetBufferSize(size uint16) {
	c.BufferSize = size
	if cap(c.readBuf) < int(size) {
		c.readBuf = make([]byte, size)
	} else {
		c.readBuf = c.readBuf[:size]
	}
}

// SetDeadlineTimeout sets read deadline timeout. Default is 5 seconds.
func (c *Client) SetDeadlineTimeout(seconds int) {
	c.Timeout = time.Duration(seconds) * time.Second
}

// Close closes UDP connection.
func (c *Client) Close() error {
	return c.Conn.Close()
}

// Get sends request and returns response data (without header), response type, ping duration and error.
// Automatically handles challenge-response if server requires it.
func (c *Client) Get(requestType Flag) ([]byte, Flag, time.Duration, error) {
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		resp, duration, err := c.request(requestType, singlePacket)
		if err != nil {
			return nil, 0, 0, err
		}
		flag := Flag(resp[4])

		for challengeAttempt := 0; challengeAttempt < 2 && flag == challengeResponse; challengeAttempt++ {
			challenge := binary.BigEndian.Uint32(resp[5:9])
			resp, _, err = c.request(requestType, challenge)
			if err != nil {
				return nil, 0, 0, err
			}
			flag = Flag(resp[4])
		}

		if err := validateResponseType(requestType, flag); err != nil {
			if requestType == RulesRequest && (flag == infoResponseSource || flag == infoResponseGoldSource || flag == challengeResponse) {
				lastErr = err
				continue
			}
			return resp[5:], flag, duration, err
		}

		return resp[5:], flag, duration, nil
	}

	if lastErr == nil {
		lastErr = ErrValidatorRules
	}
	return nil, 0, 0, lastErr
}

// request creates header, sends request and returns response with ping duration.
// Handles multi-packet responses by collecting and assembling packets.
func (c *Client) request(requestType Flag, challenge uint32) ([]byte, time.Duration, error) {
	req, err := createHeader(requestType, challenge)
	if err != nil {
		return nil, 0, err
	}

	start := time.Now()

	if _, err := c.Conn.Write(req); err != nil {
		return nil, 0, err
	}
	if err := c.Conn.SetReadDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, 0, err
	}

	var (
		resp []byte
		n    int
	)
	for attempt := 0; attempt < 3; attempt++ {
		if cap(c.readBuf) < int(c.BufferSize) {
			c.readBuf = make([]byte, c.BufferSize)
		}
		resp = c.readBuf[:c.BufferSize]
		n, err = c.Conn.Read(resp)
		if err != nil {
			return nil, 0, err
		}

		multi, err := isMultiPacket(resp[:n])
		if err != nil && errors.Is(err, ErrMultiPacket) && multi {
			continue // Some servers send a truncated split packet first; read again.
		}
		break
	}

	duration := time.Since(start)

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

	// Multi-packet response: extract metadata from first packet
	info, err := parseSplitHeader(resp[:n])
	if err != nil {
		return nil, 0, err
	}

	for k := range c.packetsBuf {
		delete(c.packetsBuf, k)
	}
	if info.packetCount > 8 && len(c.packetsBuf) == 0 {
		c.packetsBuf = make(map[int][]byte, info.packetCount)
	}

	packets := c.packetsBuf
	if n < info.dataOffset {
		return nil, 0, ErrMultiPacket
	}
	firstPacketData := make([]byte, n-info.dataOffset)
	copy(firstPacketData, resp[info.dataOffset:n])
	packets[info.currentPacket] = firstPacketData

	// Collect remaining packets
	for len(packets) < info.packetCount {
		if cap(c.readBuf) < int(c.BufferSize) {
			c.readBuf = make([]byte, c.BufferSize)
		}

		resp = c.readBuf[:c.BufferSize]
		n, err := c.Conn.Read(resp)
		if err != nil {
			return nil, 0, err
		}

		if binary.LittleEndian.Uint32(resp[4:8]) != info.packetID {
			return nil, 0, ErrMultiPacketInvalid
		}

		currentPacket := info.readPacketNumber(resp[:n])
		if _, exists := packets[currentPacket]; !exists {
			if n < info.baseHeaderSize {
				return nil, 0, ErrMultiPacket
			}
			packetData := make([]byte, n-info.baseHeaderSize)
			copy(packetData, resp[info.baseHeaderSize:n])
			packets[currentPacket] = packetData
		}
	}

	// Calculate total size and assemble packets in order
	totalSize := 0
	for i := 0; i < info.packetCount; i++ {
		if data, exists := packets[i]; exists {
			totalSize += len(data)
		} else {
			return nil, 0, ErrMultiPacketMismatch
		}
	}

	assembledResp := make([]byte, 0, totalSize)
	for i := 0; i < info.packetCount; i++ {
		if data, exists := packets[i]; exists {
			assembledResp = append(assembledResp, data...)
		} else {
			return nil, 0, ErrMultiPacketMismatch
		}
	}

	if info.compressed {
		decompressed, err := decompressBzip2(assembledResp, info.decompressedSize, info.crc)
		if err != nil {
			return nil, 0, err
		}
		return decompressed, duration, nil
	}

	return assembledResp, duration, nil
}
