package scenes

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type LaneId uint

const (
	GuitarLaneId LaneId = iota
	DrumsLaneId
	BassLaneId
)

type Note struct {
	Lane      LaneId  // Lajur tempat not ini berada (0 hingga laneCount-1).
	Tick      float64 // Waktu (dalam "tick") kapan not ini harusnya ditekan.
	IsActive  bool    // Status apakah not ini masih dalam permainan (belum ditekan atau terlewat).
	YPosition float64 // Posisi Y not di layar saat ini.
}

type Instrument struct {
	Key        ebiten.Key      // Keyboard
	Color      color.Color     // Warna Instrument
	TouchRange image.Rectangle // Range mouse/touchscreen
}
