// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package transport

const (
	// InfoRequest requests basic server information.
	InfoRequest QueryType = 0x54
	// PlayerRequest requests the current player list.
	PlayerRequest QueryType = 0x55
	// RulesRequest requests server-defined key/value rules.
	RulesRequest QueryType = 0x56
	// ChallengeRequest requests an explicit legacy challenge response.
	ChallengeRequest QueryType = 0x57
	// PingRequest requests a legacy A2A_PING response.
	PingRequest QueryType = 0x69

	// ResponseInfo identifies a Source A2S_INFO response.
	ResponseInfo ResponseType = 0x49
	// ResponseInfoGoldSource identifies a GoldSource A2S_INFO response.
	ResponseInfoGoldSource ResponseType = 0x6D
	// ResponsePlayers identifies an A2S_PLAYER response.
	ResponsePlayers ResponseType = 0x44
	// ResponseRules identifies an A2S_RULES response.
	ResponseRules ResponseType = 0x45
	// ResponseChallenge identifies a challenge response.
	ResponseChallenge ResponseType = 0x41
	// ResponsePing identifies an A2A_PING response.
	ResponsePing ResponseType = 0x6A

	// A malformed or unsupported response is retried only a small number of times
	// so unrelated datagrams cannot consume the whole query timeout.
	maxUnsupportedResponses = 3
	maxChallengeResponses   = 4
	maxPacketReads          = 6
)

const (
	// SinglePacketMarker identifies one complete logical packet.
	SinglePacketMarker uint32 = 0xFFFFFFFF
	// SplitPacketMarker identifies a fragmented logical response.
	SplitPacketMarker uint32 = 0xFFFFFFFE
	// PacketHeaderSize is the four-byte marker plus response/request type byte.
	PacketHeaderSize = 5
	// DefaultBufferSize is the default UDP receive buffer size used by Execute.
	DefaultBufferSize uint16 = 8192

	infoPayload = "Source Engine Query"

	splitMin             = 9
	srcSplitHeader       = 12
	srcSplitNoSize       = 10
	goldSrcHeader        = 9
	splitSizeOff         = 10
	splitSizeMax         = 4096
	unpackProbeMax       = 32 * 1024 * 1024
	splitPacketCountMax  = 255
	splitResponseSizeMax = 16 * 1024 * 1024
	maxDecompressedSize  = 16 * 1024 * 1024
)
