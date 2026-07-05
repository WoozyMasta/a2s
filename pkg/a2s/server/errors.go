package server

import "errors"

var (
	// ErrResponse identifies a failure while normalizing a handler response.
	ErrResponse = errors.New("a2s server: response normalization failed")

	// ErrResponseQuery identifies a response that cannot answer the request type.
	ErrResponseQuery = errors.New("a2s server: response does not match query")

	// ErrChallengeProvider identifies failure to initialize a challenge provider.
	ErrChallengeProvider = errors.New("a2s server: challenge provider initialization failed")

	// ErrChallengeGate identifies invalid challenge gate configuration or use.
	ErrChallengeGate = errors.New("a2s server: challenge gate failed")

	// ErrPacketizer identifies a Source packetizer configuration or runtime failure.
	ErrPacketizer = errors.New("a2s server: packetizer failed")

	// ErrPacketizerInput identifies invalid logical response framing.
	ErrPacketizerInput = errors.New("a2s server: invalid packetizer input")

	// ErrPacketizerSplitSize identifies an invalid Source split size.
	ErrPacketizerSplitSize = errors.New("a2s server: invalid Source split size")

	// ErrPacketizerResponseSize identifies an oversized logical response.
	ErrPacketizerResponseSize = errors.New("a2s server: logical response exceeds packetizer limit")

	// ErrPacketizerFragmentCount identifies a response requiring too many fragments.
	ErrPacketizerFragmentCount = errors.New("a2s server: too many Source response fragments")

	// ErrDrop tells the server to discard the request without sending a response.
	// It is not a server failure and should not be reported
	// as one by the transport implementation.
	ErrDrop = errors.New("a2s server: drop response")
)
