package a3sb

import "errors"

const errorPrefix string = "fail read Arma 3 server browser protocol "

var (
	// error read A2S_RULES
	ErrRules = errors.New("A2S_RULES: fail read rules")
	// error parse A2S_RULES
	ErrRulesDayZ = errors.New("A2S_RULES: fail parse DayZ rules")
	// error read A2S_RULES, not all data was read
	ErrRulesDataRemains = errors.New("A2S_RULES: not all data was read from the buffer")
	// error parse DayZ dedicated rule
	ErrRulesDayZDedicated = errors.New("A2S_RULES: invalid DayZ dedicated rule value")

	// error old unsupported protocol v1
	ErrProtoV1 = errors.New("got protocol version v1, this is the oldest version and it is not supported")
	// error v3 proto returned for expected v2 in DayZ
	ErrProtoV3 = errors.New("got v3 protocol for DayZ, contact the author on the project issues page to update the library")
	// error new unsupported protocol v4 or newest
	ErrProtoNewest = errors.New("got the latest version of the protocol, contact the author on the project issues page")

	// error in read a3sb version
	ErrVersion = errors.New(errorPrefix + "version")
	// error in read a3sb flags
	ErrFlags = errors.New(errorPrefix + "flags")
	// error in read a3sb difficulty
	ErrDifficulty = errors.New(errorPrefix + "difficulty")
	// error in read a3sb DLC
	ErrDLC = errors.New(errorPrefix + "DLC")
	// error in read a3sb mod
	ErrMod = errors.New(errorPrefix + "mod")
	// error in read a3sb signature
	ErrSignature = errors.New(errorPrefix + "signature")
	// error in read a3sb description
	ErrDescription = errors.New(errorPrefix + "description")
)
