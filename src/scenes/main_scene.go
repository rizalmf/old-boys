package scenes

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type MainScene struct {
	isLoaded bool

	// --- State Lagu ---
	songChart   []*Note // Daftar semua not dalam lagu (beatmap).
	currentTick float64 // Posisi waktu saat ini dalam lagu.
	ticksPerSec float64 // Berapa banyak "tick" yang berlalu per detik.

	// --- State Permainan ---
	score     int
	lanes     []Instrument // Konfigurasi untuk setiap lajur.
	noteSpeed float64      // Kecepatan not jatuh ke bawah (pixel per tick).
	hitZoneY  float64      // Posisi Y dari zona penilaian.
	lastFrame time.Time    // Untuk menghitung delta time.

	// --- Visual ---
	noteImage   *ebiten.Image // Gambar untuk setiap not.
	hitZoneLine *ebiten.Image // Gambar untuk garis zona penilaian.
}

func NewGameScene() *MainScene {
	return &MainScene{}
}

func (g *MainScene) Draw(screen *ebiten.Image) {

}

func (g *MainScene) ExportProperties() (prop Properties) {

	return Properties{}
}

func (g *MainScene) FirstLoad() {

}

func (g *MainScene) IsLoaded() bool {
	return g.isLoaded
}

func (g *MainScene) OnEnter(prop Properties) {

}

func (g *MainScene) OnExit() {

}

func (g *MainScene) Update() SceneId {

	return GameSceneId
}

var _ Scene = (*MainScene)(nil)
