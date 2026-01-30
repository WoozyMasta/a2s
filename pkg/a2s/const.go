package a2s

import (
	"time"
)

const (
	DefaultDeadlineTimeout time.Duration = 5    // Default deadline timeout in seconds
	DefaultBufferSize      uint16        = 4096 // conservative default to avoid UDP truncation

	singlePacket uint32 = 0xFFFFFFFF // A2S single-packet header
	multiPacket  uint32 = 0xFFFFFFFE // A2S multi-packet header

	// A2S_INFO Basic information about the server.
	InfoRequest            Flag   = 0x54
	infoResponseGoldSource Flag   = 0x6D
	infoResponseSource     Flag   = 0x49
	infoPayload            string = "Source Engine Query"

	// Extra Data Flag (EDF) in A2S_INFO
	edfPort     EDF = 0x80
	edfSteamID  EDF = 0x10
	edfSourceTV EDF = 0x40
	edfKeywords EDF = 0x20
	edfGameID   EDF = 0x01

	// A2S_PLAYER Details about each player on the server
	PlayerRequest  Flag = 0x55
	playerResponse Flag = 0x44

	// A2S_RULES The rules the server is using
	RulesRequest  Flag = 0x56
	rulesResponse Flag = 0x45

	// A2S_SERVERQUERY_GETCHALLENGE Returns a challenge number for use in the player and rules query
	ChallengeRequest  Flag = 0x57 // (DEPRECATED)
	challengeResponse Flag = 0x41

	// A2A_PING Ping the server (DEPRECATED)
	PingRequest  Flag = 0x69
	pingResponse Flag = 0x6A
)
