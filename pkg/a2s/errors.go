package a2s

import "errors"

var (
	ErrInfoRead            = errors.New("A2S_INFO: failed to read")
	ErrPlayerRead          = errors.New("A2S_PLAYER: failed to read player")
	ErrRuleRead            = errors.New("A2S_RULES: failed to read")
	ErrPingRead            = errors.New("A2S_PING: failed to read")
	ErrChallengeRead       = errors.New("A2S_SERVERQUERY_GETCHALLENGE: failed to read")
	ErrSinglePacket        = errors.New("received single packet data is too short")
	ErrMultiPacket         = errors.New("received multi packet data is too short")
	ErrWrongByte           = errors.New("unexpected response byte")
	ErrWrongRequest        = errors.New("unsupported request type")
	ErrInsufficientData    = errors.New("insufficient data length")
	ErrMultiPacketInvalid  = errors.New("received invalid packet identifier in response")
	ErrMultiPacketMismatch = errors.New("mismatched number of packets received")

	errBzip2 = errors.New("response compressed with bzip2: not implemented")
)
