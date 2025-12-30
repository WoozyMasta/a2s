package a2s

import (
	"encoding/binary"
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
	resp, duration, err := c.request(requestType, singlePacket)
	if err != nil {
		return nil, 0, 0, err
	}
	flag := Flag(resp[4])

	if flag == challengeResponse {
		challenge := binary.BigEndian.Uint32(resp[5:9])
		resp, _, err = c.request(requestType, challenge)
		if err != nil {
			return nil, 0, 0, err
		}
		flag = Flag(resp[4])
	}

	if err := validateResponseType(requestType, flag); err != nil {
		return resp[5:], flag, duration, err
	}

	return resp[5:], flag, duration, nil
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

	if cap(c.readBuf) < int(c.BufferSize) {
		c.readBuf = make([]byte, c.BufferSize)
	}
	resp := c.readBuf[:c.BufferSize]
	n, err := c.Conn.Read(resp)
	if err != nil {
		return nil, 0, err
	}

	duration := time.Since(start)

	multi, err := isMultiPacket(resp)
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
	packetID := binary.LittleEndian.Uint32(resp[4:8])
	packetCount := int(resp[8] & 0x0F)
	currentPacket := int(resp[9] & 0x0F)

	if (packetID & 0x80000000) != 0 {
		return nil, 0, errBzip2

		// TODO: implement Bzip2 unpacking, need examples of working servers with this answer
		//! this code is incorrect, need collect all multi-packets, while reading the CRC and size only from the first packet
		/*
			decompressed, _ := decompressBzip2(resp, n)
			return decompressed, duration, nil
		*/
	}

	for k := range c.packetsBuf {
		delete(c.packetsBuf, k)
	}
	if packetCount > 8 && len(c.packetsBuf) == 0 {
		c.packetsBuf = make(map[int][]byte, packetCount)
	}

	packets := c.packetsBuf
	firstPacketData := make([]byte, n-12)
	copy(firstPacketData, resp[12:n])
	packets[currentPacket] = firstPacketData

	// Collect remaining packets
	for len(packets) < packetCount {
		if cap(c.readBuf) < int(c.BufferSize) {
			c.readBuf = make([]byte, c.BufferSize)
		}

		resp = c.readBuf[:c.BufferSize]
		n, err := c.Conn.Read(resp)
		if err != nil {
			return nil, 0, err
		}

		if binary.LittleEndian.Uint32(resp[4:8]) != packetID {
			return nil, 0, ErrMultiPacketInvalid
		}

		currentPacket = int(resp[9] & 0x0F)
		if _, exists := packets[currentPacket]; !exists {
			packetData := make([]byte, n-12)
			copy(packetData, resp[12:n])
			packets[currentPacket] = packetData
		}
	}

	// Calculate total size and assemble packets in order
	totalSize := 0
	for i := 0; i < packetCount; i++ {
		if data, exists := packets[i]; exists {
			totalSize += len(data)
		} else {
			return nil, 0, ErrMultiPacketMismatch
		}
	}

	assembledResp := make([]byte, 0, totalSize)
	for i := 0; i < packetCount; i++ {
		if data, exists := packets[i]; exists {
			assembledResp = append(assembledResp, data...)
		} else {
			return nil, 0, ErrMultiPacketMismatch
		}
	}

	return assembledResp, duration, nil
}
