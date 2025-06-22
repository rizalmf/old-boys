package sounds

import (
	_ "embed"
)

var (
	//go:embed bass.mp3
	Bass_mp3 []byte

	//go:embed guitar.mp3
	Guitar_mp3 []byte

	//go:embed drums.mp3
	Drums_mp3 []byte
)
