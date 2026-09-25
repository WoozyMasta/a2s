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
)
