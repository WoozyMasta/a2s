package a3sb

import "errors"

// errorPrefix prefixes errors raised while decoding A3SB fields.
const errorPrefix string = "fail read Arma 3 server browser protocol "

var (
	// ErrRules indicates a failure while reading the A2S_RULES payload.
	ErrRules = errors.New("A2S_RULES: fail read rules")
	// ErrRulesDayZ indicates a failure while parsing DayZ rules.
	ErrRulesDayZ = errors.New("A2S_RULES: fail parse DayZ rules")
	// ErrRulesDataRemains indicates that unconsumed rule data remains.
	ErrRulesDataRemains = errors.New("A2S_RULES: not all data was read from the buffer")
	// ErrRulesPageMetadata indicates invalid or inconsistent A3SB page metadata.
	ErrRulesPageMetadata = errors.New("A2S_RULES: invalid page metadata")
	// ErrRulesPageConflict indicates different payloads for one A3SB page.
	ErrRulesPageConflict = errors.New("A2S_RULES: conflicting page payload")
	// ErrRulesPageMissing indicates that an advertised A3SB page was not received.
	ErrRulesPageMissing = errors.New("A2S_RULES: missing page")
	// ErrRulesDayZDedicated indicates an invalid DayZ dedicated value.
	ErrRulesDayZDedicated = errors.New("A2S_RULES: invalid DayZ dedicated rule value")

	// ErrEncode indicates that an A3SB binary response cannot be encoded.
	ErrEncode = errors.New("A3SB: encode failed")
	// ErrEncodeCreatorDLC indicates that the Creator DLC wire form is not supported by the encoder.
	ErrEncodeCreatorDLC = errors.New("A3SB: Creator DLC encoding is unsupported")

	// ErrProtoV1 indicates the unsupported legacy protocol version 1.
	ErrProtoV1 = errors.New("got protocol version v1, this is the oldest version and it is not supported")
	// ErrProtoV3 indicates a v3 response where DayZ v2 was expected.
	ErrProtoV3 = errors.New("got v3 protocol for DayZ, contact the author on the project issues page to update the library")
	// ErrProtoNewest indicates an unknown newer protocol version.
	ErrProtoNewest = errors.New("got the latest version of the protocol, contact the author on the project issues page")

	// ErrVersion indicates a failure while reading the A3SB version.
	ErrVersion = errors.New(errorPrefix + "version")
	// ErrFlags indicates a failure while reading A3SB flags.
	ErrFlags = errors.New(errorPrefix + "flags")
	// ErrDifficulty indicates a failure while reading A3SB difficulty.
	ErrDifficulty = errors.New(errorPrefix + "difficulty")
	// ErrDLC indicates a failure while reading A3SB DLC data.
	ErrDLC = errors.New(errorPrefix + "DLC")
	// ErrMod indicates a failure while reading A3SB mod data.
	ErrMod = errors.New(errorPrefix + "mod")
	// ErrSignature indicates a failure while reading A3SB signatures.
	ErrSignature = errors.New(errorPrefix + "signature")
	// ErrDescription indicates a failure while reading the DayZ description.
	ErrDescription = errors.New(errorPrefix + "description")
)
