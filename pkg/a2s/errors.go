package a2s

import "errors"

var (
	ErrInfoRead            = errors.New("A2S_INFO: failed to read")                       // fail read A2S_INFO
	ErrPlayerRead          = errors.New("A2S_PLAYER: failed to read player")              // fail read A2S_PLAYER
	ErrRuleRead            = errors.New("A2S_RULES: failed to read")                      // fail read A2S_RULES
	ErrPingRead            = errors.New("A2S_PING: failed to read")                       // fail read A2S_PING
	ErrChallengeRead       = errors.New("A2S_SERVERQUERY_GETCHALLENGE: failed to read")   // fail read A2S_SERVERQUERY_GETCHALLENGE
	ErrSinglePacket        = errors.New("received single packet data is too short")       // fail read single packet
	ErrMultiPacket         = errors.New("received multi packet data is too short")        // fail read multi packet
	ErrWrongByte           = errors.New("unexpected response byte")                       // fail read bytes buffer
	ErrWrongRequest        = errors.New("unsupported request type")                       // fail with unexpected response for request
	ErrInsufficientData    = errors.New("insufficient data length")                       // fail read buffer, insufficient data
	ErrMultiPacketInvalid  = errors.New("received invalid packet identifier in response") // fail with invalid response in multi packet
	ErrMultiPacketMismatch = errors.New("mismatched number of packets received")          // fail multi packet chunks mismatch

	errBzip2 = errors.New("response compressed with bzip2: not implemented")
)
