package scenes

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/rizalmf/old-boys/assets/images"
	"github.com/rizalmf/old-boys/assets/sounds"
	"github.com/rizalmf/old-boys/src/animations"
	"github.com/rizalmf/old-boys/src/entities"
)

type inGameState int8

const (
	inGameMenu inGameState = iota
	inGamePlay
	inGameFinish
)

const (
	noteLineWidth = 40
	firstNoteX    = 505
	NoteY         = 200
	NoteHeight    = 145
)

type MainScene struct {
	isLoaded bool
	// state
	state inGameState

	// Images
	Man1         entities.Char
	Man2         entities.Char
	Man3         entities.Char
	garage       *ebiten.Image
	garageDoor   *ebiten.Image
	garageInside *ebiten.Image

	// Audio
	BassAudio    *audio.Player
	GuitarAudio  *audio.Player
	DrumsAudio   *audio.Player
	AudioContext *audio.Context

	// Songs
	// --- State Song ---
	songChart   []*Note // Daftar semua not dalam lagu (beatmap).
	currentTick float64 // Posisi waktu saat ini dalam lagu.
	ticksPerSec float64 // Berapa banyak "tick" yang berlalu per detik.

	// --- State Game ---
	score     int
	lanes     []Instrument // Konfigurasi untuk setiap lajur.
	noteSpeed float64      // Kecepatan not jatuh ke bawah (pixel per tick).
	hitZoneY  float64      // Posisi Y dari zona penilaian.
	lastFrame time.Time    // Untuk menghitung delta time.

	// --- Visual ---
	isNoteMan1Pressed bool
	isNoteMan2Pressed bool
	isNoteMan3Pressed bool
	noteMan1Image     *ebiten.Image
	noteMan2Image     *ebiten.Image
	noteMan3Image     *ebiten.Image
	bgNoteImage       *ebiten.Image
	noteImage         *ebiten.Image // Gambar untuk setiap not.
	hitZoneLine       *ebiten.Image // Gambar untuk garis zona penilaian.
}

func NewGameScene() *MainScene {
	return &MainScene{}
}

func (g *MainScene) ExportProperties() (prop Properties) {

	return Properties{}
}

func (g *MainScene) FirstLoad() {

	g.state = inGamePlay

	// audio
	if g.AudioContext == nil {
		if audio.CurrentContext() != nil {
			g.AudioContext = audio.CurrentContext()
		} else {
			g.AudioContext = audio.NewContext(sounds.Rates)
		}
	}

	gMp3, err := mp3.DecodeF32(bytes.NewReader(sounds.Guitar_mp3))
	if err != nil {
		log.Fatal(err)
	}
	sfx, err := io.ReadAll(gMp3)
	if err != nil {
		log.Fatal(err)
	}
	g.GuitarAudio = g.AudioContext.NewPlayerF32FromBytes(sfx)
	g.GuitarAudio.SetVolume(0)

	dMp3, err := mp3.DecodeF32(bytes.NewReader(sounds.Drums_mp3))
	if err != nil {
		log.Fatal(err)
	}
	sfx, err = io.ReadAll(dMp3)
	if err != nil {
		log.Fatal(err)
	}
	g.DrumsAudio = g.AudioContext.NewPlayerF32FromBytes(sfx)
	g.DrumsAudio.SetVolume(0)

	bMp3, err := mp3.DecodeF32(bytes.NewReader(sounds.Bass_mp3))
	if err != nil {
		log.Fatal(err)
	}
	sfx, err = io.ReadAll(bMp3)
	if err != nil {
		log.Fatal(err)
	}
	g.BassAudio = g.AudioContext.NewPlayerF32FromBytes(sfx)
	g.BassAudio.SetVolume(0)

	// images
	img, _, err := image.Decode(bytes.NewReader(images.Garage_png))
	if err != nil {
		log.Fatal(err)
	}
	g.garage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.Door_png))
	if err != nil {
		log.Fatal(err)
	}
	g.garageDoor = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.Inside_png))
	if err != nil {
		log.Fatal(err)
	}
	g.garageInside = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.Man1_png))
	if err != nil {
		log.Fatal(err)
	}
	g.Man1 = entities.Char{
		Sprite: &entities.Sprite{
			Image: ebiten.NewImageFromImage(img),
			X:     125,
			Y:     217,
		},
		Sheet:      animations.NewSpriteSheet(176, 111),
		Animations: animations.NewAnimationHorizotal(0, 1, 32),
	}

	img, _, err = image.Decode(bytes.NewReader(images.Man2_png))
	if err != nil {
		log.Fatal(err)
	}
	g.Man2 = entities.Char{
		Sprite: &entities.Sprite{
			Image: ebiten.NewImageFromImage(img),
			X:     290,
			Y:     215,
		},
		Sheet:      animations.NewSpriteSheet(176, 111),
		Animations: animations.NewAnimationHorizotal(0, 1, 32),
	}

	img, _, err = image.Decode(bytes.NewReader(images.Man3_png))
	if err != nil {
		log.Fatal(err)
	}
	g.Man3 = entities.Char{
		Sprite: &entities.Sprite{
			Image: ebiten.NewImageFromImage(img),
			X:     203,
			Y:     218,
		},
		Sheet:      animations.NewSpriteSheet(128, 112),
		Animations: animations.NewAnimationHorizotal(0, 1, 32),
	}

	// note
	img, _, err = image.Decode(bytes.NewReader(images.Man1_ico_png))
	if err != nil {
		log.Fatal(err)
	}
	g.noteMan1Image = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.Man2_ico_png))
	if err != nil {
		log.Fatal(err)
	}
	g.noteMan2Image = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(images.Man3_ico_png))
	if err != nil {
		log.Fatal(err)
	}
	g.noteMan3Image = ebiten.NewImageFromImage(img)

	g.bgNoteImage = ebiten.NewImage(noteLineWidth*3, NoteHeight)
	g.bgNoteImage.Fill(color.Black)

	g.isLoaded = true
}

func (g *MainScene) IsLoaded() bool {
	return g.isLoaded
}

func (g *MainScene) OnEnter(prop Properties) {

}

func (g *MainScene) OnExit() {

}

func (g *MainScene) Update() SceneId {

	switch g.state {
	case inGameMenu:
		g.UpdateInGameMenu()
	case inGamePlay:
		g.UpdateInGamePlay()
	case inGameFinish:
		g.UpdateInGameFinish()
	}

	return GameSceneId
}

func (g *MainScene) UpdateInGameMenu() {
	g.BassAudio.Pause()
	g.GuitarAudio.Pause()
	g.DrumsAudio.Pause()
	g.BassAudio.Rewind()
	g.GuitarAudio.Rewind()
	g.DrumsAudio.Rewind()

}
func (g *MainScene) UpdateInGamePlay() {
	g.Man1.Animations.Update()
	g.Man2.Animations.Update()
	g.Man3.Animations.Update()

	if !g.BassAudio.IsPlaying() {
		// g.BassAudio.Rewind()
		g.BassAudio.Play()
		g.GuitarAudio.Play()
		g.DrumsAudio.Play()
		g.BassAudio.SetVolume(1)
		g.GuitarAudio.SetVolume(1)
		g.DrumsAudio.SetVolume(1)
	}

	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.GuitarAudio.SetVolume(0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.GuitarAudio.SetVolume(1)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cX, cY := ebiten.CursorPosition()
		fmt.Println(cX, cY)
	}

	g.isNoteMan1Pressed = false
	g.isNoteMan2Pressed = false
	g.isNoteMan3Pressed = false
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.isNoteMan1Pressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.isNoteMan2Pressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.isNoteMan3Pressed = true
	}

}
func (g *MainScene) UpdateInGameFinish() {

}
func (g *MainScene) Draw(screen *ebiten.Image) {

	switch g.state {
	case inGameMenu:
		g.DrawInGameMenu(screen)
	case inGamePlay:
		g.DrawInGamePlay(screen)
	case inGameFinish:
		g.DrawInGameFinish(screen)
	}
}

func (g *MainScene) DrawInGameMenu(screen *ebiten.Image) {

	op := &ebiten.DrawImageOptions{}

	screen.DrawImage(g.garageDoor, op)
	op.GeoM.Reset()

}

func (g *MainScene) DrawInGamePlay(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	// garage
	screen.DrawImage(g.garageInside, op)
	op.GeoM.Reset()

	screen.DrawImage(g.garage, op)
	op.GeoM.Reset()

	// boys
	op.GeoM.Translate(g.Man3.X, g.Man3.Y)
	Frame, _ := g.Man3.Animations.Frame()
	screen.DrawImage(g.Man3.Image.SubImage(
		g.Man3.Sheet.HorizontalRect(Frame),
	).(*ebiten.Image), op)
	op.GeoM.Reset()

	op.GeoM.Translate(g.Man1.X, g.Man1.Y)
	Frame, _ = g.Man1.Animations.Frame()
	screen.DrawImage(g.Man1.Image.SubImage(
		g.Man1.Sheet.HorizontalRect(Frame),
	).(*ebiten.Image), op)
	op.GeoM.Reset()

	op.GeoM.Translate(g.Man2.X, g.Man2.Y)
	Frame, _ = g.Man2.Animations.Frame()
	screen.DrawImage(g.Man2.Image.SubImage(
		g.Man2.Sheet.HorizontalRect(Frame),
	).(*ebiten.Image), op)
	op.GeoM.Reset()

	// note
	x := float32(firstNoteX)
	yh := NoteY + NoteHeight

	op.GeoM.Translate(float64(x), float64(NoteY))
	op.ColorScale.ScaleAlpha(0.65)
	screen.DrawImage(g.bgNoteImage, op)
	op.GeoM.Reset()

	vector.StrokeLine(screen, x, float32(NoteY), x, float32(yh), 2, color.RGBA{24, 24, 24, 255}, false)
	x += noteLineWidth
	vector.StrokeLine(screen, x, float32(NoteY), x, float32(yh), 2, color.RGBA{24, 24, 24, 255}, false)
	x += noteLineWidth
	vector.StrokeLine(screen, x, float32(NoteY), x, float32(yh), 2, color.RGBA{24, 24, 24, 255}, false)
	x += noteLineWidth
	vector.StrokeLine(screen, x, float32(NoteY), x, float32(yh), 2, color.RGBA{24, 24, 24, 255}, false)

	// rmv debris
	op = &ebiten.DrawImageOptions{}

	x = float32(firstNoteX)
	yManSm := 5
	scale := float64(1) / 3
	pressedScale := float64(1) / 3.2

	op.GeoM.Translate(0, 0)
	if g.isNoteMan1Pressed {
		op.GeoM.Scale(pressedScale, pressedScale)
	} else {
		op.GeoM.Scale(scale, scale)
	}
	op.GeoM.Translate(float64(x), float64(yh-yManSm))
	screen.DrawImage(g.noteMan1Image, op)
	op.GeoM.Reset()

	x += noteLineWidth
	op.GeoM.Translate(0, 0)
	if g.isNoteMan3Pressed {
		op.GeoM.Scale(pressedScale, pressedScale)
	} else {
		op.GeoM.Scale(scale, scale)
	}
	op.GeoM.Translate(float64(x), float64(yh-yManSm))
	screen.DrawImage(g.noteMan3Image, op)
	op.GeoM.Reset()

	x += noteLineWidth
	op.GeoM.Translate(0, 0)
	if g.isNoteMan2Pressed {
		op.GeoM.Scale(pressedScale, pressedScale)
	} else {
		op.GeoM.Scale(scale, scale)
	}
	op.GeoM.Translate(float64(x), float64(yh-yManSm))
	screen.DrawImage(g.noteMan2Image, op)
	op.GeoM.Reset()

}
func (g *MainScene) DrawInGameFinish(screen *ebiten.Image) {

}

var _ Scene = (*MainScene)(nil)
