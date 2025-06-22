package lang

type Lang interface {
	GetLang() string
	// MAIN MENU
	Start() string

	// Game
	BackToMenu() string
}

type Language uint

const (
	LanguageEN Language = iota
	LanguageID
)

func NewLanguage(lang Language) Lang {
	if lang == LanguageID {
		return &ID{}
	}

	return &EN{}
}
