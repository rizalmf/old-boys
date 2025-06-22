package entities

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Sprite struct {
	Image        *ebiten.Image
	X, Y, Dx, Dy float64
}

type Char struct {
	*Sprite
	RunSpeed         float64
	CenterX, CenterY float64
}
