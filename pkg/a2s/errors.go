package a2s

import "errors"

var (
	// A2S_INFO errors

	ErrInfoRead               = errors.New("A2S_INFO: failed to read")
	ErrInfoSourceResponse     = errors.New("A2S_INFO: Source response failed")
	ErrInfoGoldSourceResponse = errors.New("A2S_INFO: GoldSource response failed")
	ErrInfoUnsupportedFormat  = errors.New("A2S_INFO: unsupported format")
	ErrInfoProtocol           = errors.New("A2S_INFO: protocol read failed")
	ErrInfoServerName         = errors.New("A2S_INFO: server name read failed")
	ErrInfoMapName            = errors.New("A2S_INFO: map name read failed")
	ErrInfoFolderName         = errors.New("A2S_INFO: folder name read failed")
	ErrInfoGameName           = errors.New("A2S_INFO: game name read failed")
	ErrInfoGameID             = errors.New("A2S_INFO: game ID read failed")
	ErrInfoPlayerCount        = errors.New("A2S_INFO: player count read failed")
	ErrInfoMaxPlayerCount     = errors.New("A2S_INFO: max player count read failed")
	ErrInfoBotsCount          = errors.New("A2S_INFO: bots count read failed")
	ErrInfoServerType         = errors.New("A2S_INFO: server type read failed")
	ErrInfoEnvironment        = errors.New("A2S_INFO: environment read failed")
	ErrInfoVisibility         = errors.New("A2S_INFO: visibility read failed")
	ErrInfoVAC                = errors.New("A2S_INFO: VAC read failed")
	ErrInfoVersion            = errors.New("A2S_INFO: version read failed")
	ErrInfoEDF                = errors.New("A2S_INFO: EDF read failed")
	ErrInfoEDFPort            = errors.New("A2S_INFO: EDF port read failed")
	ErrInfoEDFSteamID         = errors.New("A2S_INFO: EDF SteamID read failed")
	ErrInfoEDFSourceTVPort    = errors.New("A2S_INFO: EDF SourceTV port read failed")
	ErrInfoEDFSourceTVName    = errors.New("A2S_INFO: EDF SourceTV name read failed")
	ErrInfoEDFKeywords        = errors.New("A2S_INFO: EDF keywords read failed")
	ErrInfoEDFGameID          = errors.New("A2S_INFO: EDF GameID read failed")
	ErrInfoTheShip            = errors.New("A2S_INFO: TheShip data read failed")
	ErrInfoGSAddress          = errors.New("A2S_INFO: GoldSource address read failed")
	ErrInfoGSModded           = errors.New("A2S_INFO: GoldSource modded status read failed")
	ErrInfoGSModData          = errors.New("A2S_INFO: GoldSource mod data read failed")
	ErrInfoGSModLink          = errors.New("A2S_INFO: GoldSource mod link read failed")
	ErrInfoGSModDownloadLink  = errors.New("A2S_INFO: GoldSource mod download link read failed")
	ErrInfoGSModVersion       = errors.New("A2S_INFO: GoldSource mod version read failed")
	ErrInfoGSModSize          = errors.New("A2S_INFO: GoldSource mod size read failed")
	ErrInfoGSModType          = errors.New("A2S_INFO: GoldSource mod type read failed")
	ErrInfoGSModDLL           = errors.New("A2S_INFO: GoldSource mod DLL read failed")

	// A2S_PLAYER errors

	ErrPlayerRead     = errors.New("A2S_PLAYER: failed to read player")
	ErrPlayerCount    = errors.New("A2S_PLAYER: count read failed")
	ErrPlayerIndex    = errors.New("A2S_PLAYER: index read failed")
	ErrPlayerName     = errors.New("A2S_PLAYER: name read failed")
	ErrPlayerScore    = errors.New("A2S_PLAYER: score read failed")
	ErrPlayerDuration = errors.New("A2S_PLAYER: duration read failed")
	ErrPlayerDeaths   = errors.New("A2S_PLAYER: deaths read failed")
	ErrPlayerMoney    = errors.New("A2S_PLAYER: money read failed")

	// A2S_RULES errors

	ErrRuleRead  = errors.New("A2S_RULES: failed to read")
	ErrRuleCount = errors.New("A2S_RULES: count read failed")
	ErrRuleKey   = errors.New("A2S_RULES: key read failed")
	ErrRuleValue = errors.New("A2S_RULES: value read failed")

	// A2A_PING errors

	ErrPingRead    = errors.New("A2S_PING: failed to read")
	ErrPingPayload = errors.New("A2A_PING: payload read failed")

	// A2S_SERVERQUERY_GETCHALLENGE errors

	ErrChallengeRead  = errors.New("A2S_SERVERQUERY_GETCHALLENGE: failed to read")
	ErrChallengeValue = errors.New("A2S_SERVERQUERY_GETCHALLENGE: value read failed")

	// Multi-packet errors

	ErrSinglePacket        = errors.New("received single packet data is too short")
	ErrMultiPacket         = errors.New("received multi packet data is too short")
	ErrWrongByte           = errors.New("unexpected response byte")
	ErrWrongRequest        = errors.New("unsupported request type")
	ErrHeaderWrongRequest  = errors.New("unsupported request type in header")
	ErrInsufficientData    = errors.New("insufficient data length")
	ErrMultiPacketInvalid  = errors.New("received invalid packet identifier in response")
	ErrMultiPacketMismatch = errors.New("mismatched number of packets received")

	// Validator errors

	ErrValidatorHeader    = errors.New("validator: wrong header byte")
	ErrValidatorInfo      = errors.New("validator: wrong A2S_INFO: response")
	ErrValidatorPlayer    = errors.New("validator: wrong A2S_PLAYER: response")
	ErrValidatorRules     = errors.New("validator: wrong A2S_RULES: response")
	ErrValidatorPing      = errors.New("validator: wrong A2A_PING response")
	ErrValidatorChallenge = errors.New("validator: wrong A2S_SERVERQUERY_GETCHALLENGE: response")
	ErrValidatorRequest   = errors.New("validator: wrong request type")

	// Bzip2 errors

	ErrDecompressSize         = errors.New("bz2 decompressed size exceeds limit")
	ErrDecompressFailed       = errors.New("bz2 decompression failed")
	ErrDecompressSizeMismatch = errors.New("bz2 decompressed size mismatch")
	ErrDecompressCRC          = errors.New("bz2 CRC32 checksum mismatch")
)
