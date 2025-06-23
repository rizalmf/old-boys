package animations

import (
	"image"
)

// doesnt have to associate with ebiten.Image
type SpriteSheet struct {
	HeightInTiles int
	TileSize      int
}

func (s *SpriteSheet) HorizontalRect(index int) image.Rectangle {
	x := index * s.TileSize

	return image.Rect(x, 0, x+s.TileSize, s.HeightInTiles)
}

func NewSpriteSheet(heightInTiles, tileSize int) *SpriteSheet {
	return &SpriteSheet{
		HeightInTiles: heightInTiles,
		TileSize:      tileSize,
	}
}
