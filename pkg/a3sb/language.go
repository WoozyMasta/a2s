package a3sb

import "encoding/json"

type ServerLang uint32

const (
	LangEnglish    ServerLang = 65545
	LangCzech      ServerLang = 65541
	LangGerman     ServerLang = 65543
	LangRussian    ServerLang = 65561
	LangPolish     ServerLang = 65557
	LangHungarian  ServerLang = 65550
	LangItalian    ServerLang = 65552
	LangSpanish    ServerLang = 65546
	LangFrench     ServerLang = 65548
	LangChinese    ServerLang = 65540
	LangJapanese   ServerLang = 65553
	LangPortuguese ServerLang = 65558
)

func (l ServerLang) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// Return string represent of uint32 value in language keyword in A2S_RULES for DayZ
func (l ServerLang) String() string {
	switch l {
	case LangEnglish:
		return "English"
	case LangCzech:
		return "Czech"
	case LangGerman:
		return "German"
	case LangRussian:
		return "Russian"
	case LangPolish:
		return "Polish"
	case LangHungarian:
		return "Hungarian"
	case LangItalian:
		return "Italian"
	case LangSpanish:
		return "Spanish"
	case LangFrench:
		return "French"
	case LangChinese:
		return "Chinese"
	case LangJapanese:
		return "Japanese"
	case LangPortuguese:
		return "Portuguese"
	default:
		return "English"
	}
}
