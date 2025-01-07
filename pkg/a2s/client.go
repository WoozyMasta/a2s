package a2s

import (
	"encoding/binary"
	"net"
	"strconv"
	"time"
)

type Client struct {
	Conn       *net.UDPConn
	Address    *net.UDPAddr
	Timeout    time.Duration
	BufferSize uint16
}

// Create client and open connection
func New(ip string, port int) (*Client, error) {
	client, err := CreateClient(ip, port)
	if err != nil {
		return nil, err
	}

	if err := client.Dial(); err != nil {
		return nil, err
	}

	return client, nil
}

// Create client only
func CreateClient(ip string, port int) (*Client, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(ip, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}

	return &Client{
		Address:    udpAddr,
		Timeout:    DefaultDeadlineTimeout * time.Second,
		BufferSize: DefaultBufferSize,
	}, nil
}

// Open connection for Client
func (c *Client) Dial() error {
	conn, err := net.DialUDP("udp", nil, c.Address)
	if err != nil {
		return err
	}

	c.Conn = conn
	return nil
}

// Set bytes buffer size for Client, default is 1400
func (c *Client) SetBufferSize(size uint16) {
	c.BufferSize = size
}

// Set deadline timeout for Client, default is 5 seconds
func (c *Client) SetDeadlineTimeout(seconds int) {
	c.Timeout = time.Duration(seconds) * time.Second
}

// Close connection for Client
func (c *Client) Close() error {
	return c.Conn.Close()
}

// Execute request in Client with type and return:
//   - response bytes without header
//   - response type
//   - ping (duration from writing bytes to reading response)
//   - error (if any)
//
// Is responsible for processing the challenge and executes it if available
func (c *Client) Get(requestType Flag) ([]byte, Flag, time.Duration, error) {
	resp, duration, err := c.request(requestType, singlePacket)
	if err != nil {
		return nil, 0, 0, err
	}
	flag := Flag(resp[4])

	if flag == challengeResponse {
		challenge := binary.BigEndian.Uint32(resp[5:9])
		// use duration from challenge as most real ping
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

// Creates a request header, executes the request and returns the response, handles multi-packet response processing
func (c *Client) request(requestType Flag, challenge uint32) ([]byte, time.Duration, error) {
	req, err := createHeader(requestType, challenge)
	if err != nil {
		return nil, 0, err
	}

	// start send (ping start)
	start := time.Now()

	if _, err := c.Conn.Write(req); err != nil {
		return nil, 0, err
	}
	if err := c.Conn.SetReadDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, 0, err
	}

	resp := make([]byte, c.BufferSize)
	n, err := c.Conn.Read(resp)
	if err != nil {
		return nil, 0, err
	}

	// end receive (ping end)
	duration := time.Since(start)

	multi, err := isMultiPacket(resp)
	if err != nil {
		return resp, 0, err
	}

	// return single-packet data
	if !multi {
		return resp[:n], duration, nil
	}

	// multi-packet processing
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

	// Add first packet
	packets := make(map[int][]byte)
	packets[currentPacket] = resp[12:n]

	// Read remaining packets
	for len(packets) < packetCount {
		resp := make([]byte, c.BufferSize)
		n, err := c.Conn.Read(resp)
		if err != nil {
			return nil, 0, err
		}

		if binary.LittleEndian.Uint32(resp[4:8]) != packetID {
			return nil, 0, ErrMultiPacketInvalid
		}

		currentPacket = int(resp[9] & 0x0F)
		if _, exists := packets[currentPacket]; !exists {
			packets[currentPacket] = resp[12:n]
		}
	}

	// Combine packets in order
	var assembledResp []byte
	for i := 0; i < packetCount; i++ {
		if data, exists := packets[i]; exists {
			assembledResp = append(assembledResp, data...)
		} else {
			return nil, 0, ErrMultiPacketMismatch
		}
	}

	return assembledResp, duration, nil
}
