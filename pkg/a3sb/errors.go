package a3sb

import "errors"

const errorPrefix string = "fail read Arma 3 server browser protocol "

var (
	ErrRules            = errors.New("A2S_RULES: fail read rules")
	ErrRulesDayZ        = errors.New("A2S_RULES: fail parse DayZ rules")
	ErrRulesDataRemains = errors.New("A2S_RULES: not all data was read from the buffer")

	ErrProtoV1     = errors.New("got protocol version v1, this is the oldest version and it is not supported")
	ErrProtoV3     = errors.New("got v3 protocol for DayZ, contact the author on the project issues page to update the library")
	ErrProtoNewest = errors.New("got the latest version of the protocol, contact the author on the project issues page")

	ErrVersion     = errors.New(errorPrefix + "version")
	ErrFlags       = errors.New(errorPrefix + "flags")
	ErrDifficulty  = errors.New(errorPrefix + "difficulty")
	ErrDLC         = errors.New(errorPrefix + "DLC")
	ErrMod         = errors.New(errorPrefix + "mod")
	ErrSignature   = errors.New(errorPrefix + "signature")
	ErrDescription = errors.New(errorPrefix + "description")
)
