// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a2s

import "errors"

var (
	// ErrInvalidAddress indicates that the server address is missing or invalid.
	ErrInvalidAddress = errors.New("a2s: invalid server address")
	// ErrInvalidTimeout indicates that the client timeout is not positive.
	ErrInvalidTimeout = errors.New("a2s: timeout must be positive")
	// ErrInvalidBufferSize indicates that the client buffer size is not positive.
	ErrInvalidBufferSize = errors.New("a2s: buffer size must be positive")
	// ErrClientClosed indicates that the client connection has already been closed.
	ErrClientClosed = errors.New("a2s: client connection is closed")
	// ErrNilContext indicates that a nil context was passed to a client operation.
	ErrNilContext = errors.New("a2s: nil context")

	// A2S_INFO errors

	// ErrInfoRead indicates that the A2S_INFO response could not be read.
	ErrInfoRead = errors.New("A2S_INFO: failed to read")
	// ErrInfoSourceResponse indicates that the Source A2S_INFO response is invalid.
	ErrInfoSourceResponse = errors.New("A2S_INFO: Source response failed")
	// ErrInfoGoldSourceResponse indicates that the GoldSource A2S_INFO response is invalid.
	ErrInfoGoldSourceResponse = errors.New("A2S_INFO: GoldSource response failed")
	// ErrInfoUnsupportedFormat indicates that the A2S_INFO response format is unsupported.
	ErrInfoUnsupportedFormat = errors.New("A2S_INFO: unsupported format")
	// ErrInfoEncode indicates that an A2S_INFO request could not be encoded.
	ErrInfoEncode = errors.New("A2S_INFO: encode failed")
	// ErrInfoProtocol indicates that the protocol version could not be read.
	ErrInfoProtocol = errors.New("A2S_INFO: protocol read failed")
	// ErrInfoServerName indicates that the server name could not be read.
	ErrInfoServerName = errors.New("A2S_INFO: server name read failed")
	// ErrInfoMapName indicates that the map name could not be read.
	ErrInfoMapName = errors.New("A2S_INFO: map name read failed")
	// ErrInfoFolderName indicates that the game folder name could not be read.
	ErrInfoFolderName = errors.New("A2S_INFO: folder name read failed")
	// ErrInfoGameName indicates that the game name could not be read.
	ErrInfoGameName = errors.New("A2S_INFO: game name read failed")
	// ErrInfoAppID indicates that the legacy AppID could not be read.
	ErrInfoAppID = errors.New("A2S_INFO: AppID read failed")
	// ErrInfoPlayerCount indicates that the current player count could not be read.
	ErrInfoPlayerCount = errors.New("A2S_INFO: player count read failed")
	// ErrInfoMaxPlayerCount indicates that the maximum player count could not be read.
	ErrInfoMaxPlayerCount = errors.New("A2S_INFO: max player count read failed")
	// ErrInfoBotsCount indicates that the bot count could not be read.
	ErrInfoBotsCount = errors.New("A2S_INFO: bots count read failed")
	// ErrInfoServerType indicates that the server type could not be read.
	ErrInfoServerType = errors.New("A2S_INFO: server type read failed")
	// ErrInfoEnvironment indicates that the server environment could not be read.
	ErrInfoEnvironment = errors.New("A2S_INFO: environment read failed")
	// ErrInfoVisibility indicates that the server visibility could not be read.
	ErrInfoVisibility = errors.New("A2S_INFO: visibility read failed")
	// ErrInfoVAC indicates that the VAC status could not be read.
	ErrInfoVAC = errors.New("A2S_INFO: VAC read failed")
	// ErrInfoVersion indicates that the game version could not be read.
	ErrInfoVersion = errors.New("A2S_INFO: version read failed")
	// ErrInfoEDF indicates that the extra data flags could not be read.
	ErrInfoEDF = errors.New("A2S_INFO: EDF read failed")
	// ErrInfoEDFPort indicates that the EDF game port could not be read.
	ErrInfoEDFPort = errors.New("A2S_INFO: EDF port read failed")
	// ErrInfoEDFSteamID indicates that the EDF server SteamID could not be read.
	ErrInfoEDFSteamID = errors.New("A2S_INFO: EDF SteamID read failed")
	// ErrInfoEDFSourceTVPort indicates that the EDF SourceTV port could not be read.
	ErrInfoEDFSourceTVPort = errors.New("A2S_INFO: EDF SourceTV port read failed")
	// ErrInfoEDFSourceTVName indicates that the EDF SourceTV name could not be read.
	ErrInfoEDFSourceTVName = errors.New("A2S_INFO: EDF SourceTV name read failed")
	// ErrInfoEDFKeywords indicates that the EDF server keywords could not be read.
	ErrInfoEDFKeywords = errors.New("A2S_INFO: EDF keywords read failed")
	// ErrInfoEDFGameID indicates that the EDF full game identifier could not be read.
	ErrInfoEDFGameID = errors.New("A2S_INFO: EDF GameID read failed")
	// ErrInfoTheShip indicates that The Ship-specific data could not be read.
	ErrInfoTheShip = errors.New("A2S_INFO: TheShip data read failed")
	// ErrInfoGSAddress indicates that the GoldSource server address could not be read.
	ErrInfoGSAddress = errors.New("A2S_INFO: GoldSource address read failed")
	// ErrInfoGSModded indicates that the GoldSource modded status could not be read.
	ErrInfoGSModded = errors.New("A2S_INFO: GoldSource modded status read failed")
	// ErrInfoGSModData indicates that GoldSource mod data could not be read.
	ErrInfoGSModData = errors.New("A2S_INFO: GoldSource mod data read failed")
	// ErrInfoGSModLink indicates that the GoldSource mod link could not be read.
	ErrInfoGSModLink = errors.New("A2S_INFO: GoldSource mod link read failed")
	// ErrInfoGSModDownloadLink indicates that the GoldSource mod download link could not be read.
	ErrInfoGSModDownloadLink = errors.New("A2S_INFO: GoldSource mod download link read failed")
	// ErrInfoGSModVersion indicates that the GoldSource mod version could not be read.
	ErrInfoGSModVersion = errors.New("A2S_INFO: GoldSource mod version read failed")
	// ErrInfoGSModSize indicates that the GoldSource mod size could not be read.
	ErrInfoGSModSize = errors.New("A2S_INFO: GoldSource mod size read failed")
	// ErrInfoGSModType indicates that the GoldSource mod type could not be read.
	ErrInfoGSModType = errors.New("A2S_INFO: GoldSource mod type read failed")
	// ErrInfoGSModDLL indicates that the GoldSource mod DLL status could not be read.
	ErrInfoGSModDLL = errors.New("A2S_INFO: GoldSource mod DLL read failed")

	// A2S_PLAYER errors

	// ErrPlayerRead indicates that a player entry could not be read.
	ErrPlayerRead = errors.New("A2S_PLAYER: failed to read player")
	// ErrPlayerEncode indicates that an A2S_PLAYER request could not be encoded.
	ErrPlayerEncode = errors.New("A2S_PLAYER: encode failed")
	// ErrPlayerCount indicates that the player count could not be read.
	ErrPlayerCount = errors.New("A2S_PLAYER: count read failed")
	// ErrPlayerIndex indicates that a player index could not be read.
	ErrPlayerIndex = errors.New("A2S_PLAYER: index read failed")
	// ErrPlayerName indicates that a player name could not be read.
	ErrPlayerName = errors.New("A2S_PLAYER: name read failed")
	// ErrPlayerScore indicates that a player score could not be read.
	ErrPlayerScore = errors.New("A2S_PLAYER: score read failed")
	// ErrPlayerDuration indicates that a player duration could not be read.
	ErrPlayerDuration = errors.New("A2S_PLAYER: duration read failed")
	// ErrPlayerDeaths indicates that a player death count could not be read.
	ErrPlayerDeaths = errors.New("A2S_PLAYER: deaths read failed")
	// ErrPlayerMoney indicates that a player money value could not be read.
	ErrPlayerMoney = errors.New("A2S_PLAYER: money read failed")

	// A2S_RULES errors

	// ErrRuleRead indicates that an A2S_RULES response could not be read.
	ErrRuleRead = errors.New("A2S_RULES: failed to read")
	// ErrRuleEncode indicates that an A2S_RULES request could not be encoded.
	ErrRuleEncode = errors.New("A2S_RULES: encode failed")
	// ErrRuleCount indicates that the A2S_RULES entry count could not be read.
	ErrRuleCount = errors.New("A2S_RULES: count read failed")
	// ErrRuleKey indicates that an A2S_RULES key could not be read.
	ErrRuleKey = errors.New("A2S_RULES: key read failed")
	// ErrRuleValue indicates that an A2S_RULES value could not be read.
	ErrRuleValue = errors.New("A2S_RULES: value read failed")

	// A2A_PING errors

	// ErrPingRead indicates that an A2S_PING response could not be read.
	ErrPingRead = errors.New("A2S_PING: failed to read")
	// ErrPingPayload indicates that an A2A_PING payload is invalid.
	ErrPingPayload = errors.New("A2A_PING: payload read failed")

	// A2S_SERVERQUERY_GETCHALLENGE errors

	// ErrChallengeRead indicates that a challenge response could not be read.
	ErrChallengeRead = errors.New("A2S_SERVERQUERY_GETCHALLENGE: failed to read")
	// ErrChallengeValue indicates that a challenge value is invalid or missing.
	ErrChallengeValue = errors.New("A2S_SERVERQUERY_GETCHALLENGE: value read failed")

	// Multi-packet errors

	// ErrSinglePacket indicates that a single-packet response is too short.
	ErrSinglePacket = errors.New("received single packet data is too short")
	// ErrMultiPacket indicates that a multi-packet response is too short.
	ErrMultiPacket = errors.New("received multi packet data is too short")
	// ErrWrongByte indicates that a response contains an unexpected byte.
	ErrWrongByte = errors.New("unexpected response byte")
	// ErrWrongRequest indicates that a request type is unsupported.
	ErrWrongRequest = errors.New("unsupported request type")
	// ErrHeaderWrongRequest indicates that a packet header contains an unsupported request type.
	ErrHeaderWrongRequest = errors.New("unsupported request type in header")
	// ErrRequestHeader indicates that an A2S request header is invalid.
	ErrRequestHeader = errors.New("invalid A2S request header")
	// ErrRequestPayload indicates that an A2S request payload is invalid.
	ErrRequestPayload = errors.New("invalid A2S request payload")
	// ErrPacketHeader indicates that an A2S packet header is invalid.
	ErrPacketHeader = errors.New("invalid A2S packet header")
	// ErrInsufficientData indicates that a response has insufficient data.
	ErrInsufficientData = errors.New("insufficient data length")
	// ErrMultiPacketInvalid indicates that a response has an invalid packet identifier.
	ErrMultiPacketInvalid = errors.New("received invalid packet identifier in response")
	// ErrMultiPacketMismatch indicates that multi-packet metadata does not match.
	ErrMultiPacketMismatch = errors.New("mismatched number of packets received")
	// ErrMultiPacketSize indicates that a multi-packet response exceeds the size limit.
	ErrMultiPacketSize = errors.New("multi packet response exceeds size limit")
	// ErrMultiPacketConflict indicates that duplicate multi-packet fragments conflict.
	ErrMultiPacketConflict = errors.New("conflicting duplicate multi packet fragment")
	// ErrMultiPacketInconsistent indicates that multi-packet fragment metadata is inconsistent.
	ErrMultiPacketInconsistent = errors.New("inconsistent multi packet fragment metadata")
	// ErrChallengeLoop indicates that a server keeps returning challenge responses.
	ErrChallengeLoop = errors.New("server keeps returning challenge response")
	// ErrQueryUnsupported indicates that the server does not support a requested query.
	ErrQueryUnsupported = errors.New("server does not support requested query")

	// Validator errors

	// ErrValidatorHeader indicates that a response has an invalid header byte.
	ErrValidatorHeader = errors.New("validator: wrong header byte")
	// ErrValidatorInfo indicates that a response is not a valid A2S_INFO response.
	ErrValidatorInfo = errors.New("validator: wrong A2S_INFO: response")
	// ErrValidatorPlayer indicates that a response is not a valid A2S_PLAYER response.
	ErrValidatorPlayer = errors.New("validator: wrong A2S_PLAYER: response")
	// ErrValidatorRules indicates that a response is not a valid A2S_RULES response.
	ErrValidatorRules = errors.New("validator: wrong A2S_RULES: response")
	// ErrValidatorPing indicates that a response is not a valid A2A_PING response.
	ErrValidatorPing = errors.New("validator: wrong A2A_PING response")
	// ErrValidatorChallenge indicates that a response is not a valid challenge response.
	ErrValidatorChallenge = errors.New("validator: wrong A2S_SERVERQUERY_GETCHALLENGE: response")
	// ErrValidatorRequest indicates that a request type is invalid.
	ErrValidatorRequest = errors.New("validator: wrong request type")

	// Bzip2 errors

	// ErrDecompressSize indicates that decompressed data exceeds the size limit.
	ErrDecompressSize = errors.New("bz2 decompressed size exceeds limit")
	// ErrDecompressFailed indicates that bzip2 decompression failed.
	ErrDecompressFailed = errors.New("bz2 decompression failed")
	// ErrDecompressSizeMismatch indicates that the decompressed size is unexpected.
	ErrDecompressSizeMismatch = errors.New("bz2 decompressed size mismatch")
	// ErrDecompressCRC indicates that the decompressed data has an invalid CRC32.
	ErrDecompressCRC = errors.New("bz2 CRC32 checksum mismatch")
)
