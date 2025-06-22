package lang

type ID struct {
}

func (l *ID) GetLang() string {
	return "bahasa"
}

func (l *ID) BackToMenu() string {
	return "Tekan 'Esc' untuk kembali"
}

// MAIN MENU
func (l *ID) Start() string {
	return "Mulai"
}

var _ Lang = (*ID)(nil)
