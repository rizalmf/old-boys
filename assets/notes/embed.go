package notes

import (
	_ "embed"
)

var (
	//go:embed note.json
	Note_json []byte
)
