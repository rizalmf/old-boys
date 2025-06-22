package scenes

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Instrument mengasosiasikan sebuah lajur dengan tombol keyboard dan warna.
type Instrument struct {
	Key   ebiten.Key
	Color color.Color
}
