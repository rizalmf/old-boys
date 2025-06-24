package entities

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rizalmf/old-boys/src/animations"
)

type Sprite struct {
	Image        *ebiten.Image
	X, Y, Dx, Dy float64
}

type Char struct {
	*Sprite
	Sheet            *animations.SpriteSheet
	Animations       *animations.Animation // single state animation
	RunSpeed         float64
	CenterX, CenterY float64

	MarkImage       *ebiten.Image
	IsMark          bool
	MarkTime        int
	CurrentMarkTime int
}
