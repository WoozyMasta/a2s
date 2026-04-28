package a2s

import (
	"time"
)

const (
	// DefaultDeadlineTimeout is the default UDP deadline in seconds.
	DefaultDeadlineTimeout time.Duration = 5
	// DefaultBufferSize is the default UDP receive buffer size.
	DefaultBufferSize uint16 = 4096

	// singlePacket identifies a complete response in the A2S packet header.
	singlePacket uint32 = 0xFFFFFFFF
	// multiPacket identifies a response split across multiple A2S packets.
	multiPacket uint32 = 0xFFFFFFFE

	// A2S_INFO

	// InfoRequest requests basic information about the server.
	InfoRequest Flag = 0x54
	// infoResponseGoldSource identifies an obsolete GoldSource response.
	infoResponseGoldSource Flag = 0x6D
	// infoResponseSource identifies a Source response.
	infoResponseSource Flag = 0x49
	// infoPayload is the payload used by an A2S_INFO request.
	infoPayload string = "Source Engine Query"

	// edfPort indicates that the game port follows the base response.
	edfPort EDF = 0x80
	// edfSteamID indicates that the server SteamID follows the base response.
	edfSteamID EDF = 0x10
	// edfSourceTV indicates that SourceTV port and name follow the base response.
	edfSourceTV EDF = 0x40
	// edfKeywords indicates that server keywords follow the base response.
	edfKeywords EDF = 0x20
	// edfGameID indicates that the full 64-bit game ID follows the base response.
	edfGameID EDF = 0x01

	// A2S_PLAYER

	// PlayerRequest requests details about each player on the server.
	PlayerRequest Flag = 0x55
	// playerResponse identifies an A2S_PLAYER response.
	playerResponse Flag = 0x44

	// A2S_RULES

	// RulesRequest requests the rules used by the server.
	RulesRequest Flag = 0x56
	// rulesResponse identifies an A2S_RULES response.
	rulesResponse Flag = 0x45

	// A2S_SERVERQUERY_GETCHALLENGE (DEPRECATED)

	// ChallengeRequest requests a challenge for player and rules queries.
	ChallengeRequest Flag = 0x57
	// challengeResponse identifies an A2S challenge response.
	challengeResponse Flag = 0x41

	// A2A_PING (DEPRECATED)

	// PingRequest requests a legacy A2A_PING response.
	PingRequest Flag = 0x69
	// pingResponse identifies an A2A_PING response.
	pingResponse Flag = 0x6A
)
