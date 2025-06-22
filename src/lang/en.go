package lang

type EN struct {
}

func (l *EN) GetLang() string {
	return "english"
}

func (l *EN) BackToMenu() string {
	return "Press Esc to back main menu"
}

// MAIN MENU
func (l *EN) Start() string {
	return "Start"
}

var _ Lang = (*EN)(nil)
