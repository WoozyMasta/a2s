package a3sb

import "encoding/json"

// Represent game-server language
type ServerLang uint32

const (
	LangEnglish    ServerLang = 65545 // English
	LangCzech      ServerLang = 65541 // Czech
	LangGerman     ServerLang = 65543 // German
	LangRussian    ServerLang = 65561 // Russian
	LangPolish     ServerLang = 65557 // Polish
	LangHungarian  ServerLang = 65550 // Hungarian
	LangItalian    ServerLang = 65552 // Italian
	LangSpanish    ServerLang = 65546 // Spanish
	LangFrench     ServerLang = 65548 // French
	LangChinese    ServerLang = 65540 // Chinese
	LangJapanese   ServerLang = 65553 // Japanese
	LangPortuguese ServerLang = 65558 // Portuguese
)

// Helper for convert internal type to string in JSON
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
