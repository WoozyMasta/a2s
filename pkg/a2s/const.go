// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import (
	"time"
)

const (
	// DefaultDeadlineTimeout is the default UDP read deadline.
	DefaultDeadlineTimeout time.Duration = 5 * time.Second
	// DefaultBufferSize is the default UDP receive buffer size.
	// It accommodates the largest supported single A2S/A3SB datagrams.
	DefaultBufferSize uint16 = 8192

	// singlePacket identifies a complete response in the A2S packet header.
	singlePacket uint32 = 0xFFFFFFFF
	// multiPacket identifies a response split across multiple A2S packets.
	multiPacket uint32 = 0xFFFFFFFE

	// A2S_INFO

	// InfoRequest requests basic information about the server.
	InfoRequest QueryType = 0x54
	// ResponseInfoGoldSource identifies an obsolete GoldSource response.
	ResponseInfoGoldSource ResponseType = 0x6D
	// ResponseInfo identifies a Source A2S_INFO response.
	ResponseInfo ResponseType = 0x49
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
	PlayerRequest QueryType = 0x55
	// ResponsePlayers identifies an A2S_PLAYER response.
	ResponsePlayers ResponseType = 0x44

	// A2S_RULES

	// RulesRequest requests the rules used by the server.
	RulesRequest QueryType = 0x56
	// ResponseRules identifies an A2S_RULES response.
	ResponseRules ResponseType = 0x45

	// A2S_SERVERQUERY_GETCHALLENGE (DEPRECATED)

	// ChallengeRequest requests a challenge for player and rules queries.
	ChallengeRequest QueryType = 0x57
	// ResponseChallenge identifies an A2S challenge response.
	ResponseChallenge ResponseType = 0x41

	// A2A_PING (DEPRECATED)

	// PingRequest requests a legacy A2A_PING response.
	PingRequest QueryType = 0x69
	// ResponsePing identifies an A2A_PING response.
	ResponsePing ResponseType = 0x6A
)
