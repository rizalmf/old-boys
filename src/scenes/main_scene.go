package scenes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/rizalmf/old-boys/assets/fonts"
	"github.com/rizalmf/old-boys/assets/images"
	"github.com/rizalmf/old-boys/assets/notes"
	"github.com/rizalmf/old-boys/assets/sounds"
	"github.com/rizalmf/old-boys/src/animations"
	"github.com/rizalmf/old-boys/src/constants"
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

type ScoreCriteria struct {
	perfect int
	good    int
	miss    int
}

type Score struct {
	perfectNote uint
	goodNote    uint
	missNote    int
	guitarScore ScoreCriteria
	bassScore   ScoreCriteria
	drumScore   ScoreCriteria
}

type MainScene struct {
	isLoaded    bool
	loadCount   int
	loadTotal   int
	loadPercent int

	// Fonts
	fontSource *text.GoTextFaceSource

	// State
	state       inGameState
	isVeryBegin bool

	// Images
	Man1         entities.Char
	Man2         entities.Char
	Man3         entities.Char
	sky          *ebiten.Image
	skyOffset    float64
	garageAndSky *ebiten.Image
	garage       *ebiten.Image
	garageDoor   *ebiten.Image
	garageInside *ebiten.Image
	// Menu Animation
	garageAnimY      float64
	garageAnimActive bool
	doorAnimY        float64
	doorAnimActive   bool
	// Finish Animation
	finishAnimY      float64
	finishAnimActive bool
	isFinishAnim     bool

	// Audio
	BassAudio    *audio.Player
	GuitarAudio  *audio.Player
	DrumsAudio   *audio.Player
	AudioContext *audio.Context
	GarageSFX    []byte

	// Songs
	// --- State Song ---
	songChart   []*Note // Daftar semua not dalam lagu (beatmap).
	currentTick float64 // Posisi waktu saat ini dalam lagu.
	ticksPerSec float64 // Berapa banyak "tick" yang berlalu per detik.

	// --- State Game ---
	scoreVal  int
	score     Score
	touchIDs  []ebiten.TouchID
	lanes     []Instrument // Konfigurasi untuk setiap lajur.
	noteSpeed float64      // Kecepatan not jatuh ke bawah (pixel per tick).
	hitZoneY  float64      // Posisi Y dari zona penilaian.
	lastFrame time.Time    // Untuk menghitung delta time.

	// --- Visual ---
	markPerfectImage  *ebiten.Image
	markGoodImage     *ebiten.Image
	markMissImage     *ebiten.Image
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
	return &MainScene{
		ticksPerSec: 100.0, // "BPM" virtual.
		noteSpeed:   0.7,   // kecepatan visual not.
		lastFrame:   time.Now(),
		songChart:   make([]*Note, 0),
		hitZoneY:    338,
	}
}

func (g *MainScene) ExportProperties() (prop Properties) {

	return Properties{}
}

func (g *MainScene) FirstLoad() {

	g.isVeryBegin = true
	g.state = inGameFinish
	g.loadCount = 0
	g.loadTotal = 25
	g.score = Score{
		perfectNote: 100,
		goodNote:    50,
		missNote:    0, // tanpa penalty
	}

	var err error

	// fonts
	g.loadCount++
	g.fontSource, err = text.NewGoTextFaceSource(bytes.NewReader(fonts.Font_otf))
	if err != nil {
		log.Fatal(err)
	}

	// note
	g.loadCount++
	img, _, err := image.Decode(bytes.NewReader(images.Man1_ico_png))
	if err != nil {
		log.Fatal(err)
	}
	g.noteMan1Image = ebiten.NewImageFromImage(img)

	g.loadCount++
	img, _, err = image.Decode(bytes.NewReader(images.Man2_ico_png))
	if err != nil {
		log.Fatal(err)
	}
	g.noteMan2Image = ebiten.NewImageFromImage(img)

	g.loadCount++
	img, _, err = image.Decode(bytes.NewReader(images.Man3_ico_png))
	if err != nil {
		log.Fatal(err)
	}
	g.noteMan3Image = ebiten.NewImageFromImage(img)

	g.loadCount++
	g.bgNoteImage = ebiten.NewImage(noteLineWidth*3, NoteHeight)
	g.bgNoteImage.Fill(color.Black)

	g.loadCount++
	g.noteImage = ebiten.NewImage(int(noteLineWidth), 8)
	g.noteImage.Fill(color.White)

	g.loadCount++
	g.hitZoneLine = ebiten.NewImage(noteLineWidth, 4)
	g.hitZoneLine.Fill(color.RGBA{255, 255, 255, 128})

	g.loadCount++
	g.lanes = []Instrument{
		{Key: ebiten.KeyLeft, Color: color.RGBA{150, 75, 0, 255},
			TouchRange: image.Rect(500, 348, 536, 374),
		}, // Soklat
		{Key: ebiten.KeyDown, Color: color.RGBA{255, 255, 255, 255},
			TouchRange: image.Rect(547, 348, 581, 374),
		}, // Putih
		{Key: ebiten.KeyRight, Color: color.RGBA{100, 255, 100, 255},
			TouchRange: image.Rect(586, 348, 621, 374),
		}, // Hijau
	}

	g.loadCount++
	err = json.Unmarshal(notes.Note_json, &g.songChart)
	if err != nil {
		log.Fatal(err)
	}

	// audio
	g.loadCount++
	if g.AudioContext == nil {
		if audio.CurrentContext() != nil {
			g.AudioContext = audio.CurrentContext()
		} else {
			g.AudioContext = audio.NewContext(sounds.Rates)
		}
	}

	g.loadCount++
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

	g.loadCount++
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

	g.loadCount++
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

	g.loadCount++
	grMp3, err := mp3.DecodeF32(bytes.NewReader(sounds.Garage_mp3))
	if err != nil {
		log.Fatal(err)
	}
	g.GarageSFX, err = io.ReadAll(grMp3)
	if err != nil {
		log.Fatal(err)
	}

	// images
	g.loadCount++
	img, _, err = image.Decode(bytes.NewReader(images.Sky_png))
	if err != nil {
		log.Fatal(err)
	}
	g.sky = ebiten.NewImageFromImage(img)

	g.loadCount++
	img, _, err = image.Decode(bytes.NewReader(images.Garage_png))
	if err != nil {
		log.Fatal(err)
	}
	g.garage = ebiten.NewImageFromImage(img)

	g.loadCount++
	img, _, err = image.Decode(bytes.NewReader(images.GarageAndSky_png))
	if err != nil {
		log.Fatal(err)
	}
	g.garageAndSky = ebiten.NewImageFromImage(img)

	g.loadCount++
	img, _, err = image.Decode(bytes.NewReader(images.Door_png))
	if err != nil {
		log.Fatal(err)
	}
	g.garageDoor = ebiten.NewImageFromImage(img)

	g.loadCount++
	img, _, err = image.Decode(bytes.NewReader(images.Inside_png))
	if err != nil {
		log.Fatal(err)
	}
	g.garageInside = ebiten.NewImageFromImage(img)

	g.loadCount++
	markTime := 1 * 60
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
		MarkTime:   markTime,
	}

	g.loadCount++
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
		MarkTime:   markTime,
	}

	g.loadCount++
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
		MarkTime:   markTime,
	}

	g.loadCount++
	img, _, err = image.Decode(bytes.NewReader(images.MarkPerfect_png))
	if err != nil {
		log.Fatal(err)
	}
	g.markPerfectImage = ebiten.NewImageFromImage(img)

	g.loadCount++
	img, _, err = image.Decode(bytes.NewReader(images.MarkGood_png))
	if err != nil {
		log.Fatal(err)
	}
	g.markGoodImage = ebiten.NewImageFromImage(img)

	g.loadCount++
	img, _, err = image.Decode(bytes.NewReader(images.MarkMiss_png))
	if err != nil {
		log.Fatal(err)
	}
	g.markMissImage = ebiten.NewImageFromImage(img)

	g.garageAnimY = float64(constants.ScreenHeight)
	g.garageAnimActive = false

	g.doorAnimY = 0
	g.doorAnimActive = false

	// Initialize finish animation
	g.finishAnimY = 0
	g.finishAnimActive = true
	g.isFinishAnim = false

	// g.state = inGamePlay
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

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		cX, cY := ebiten.CursorPosition()
		fmt.Println(cX, cY)
	}

	if !g.isLoaded {
		g.loadPercent = int(float64(g.loadCount) / float64(g.loadTotal) * 100)
		return GameSceneId
	}

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

	if !g.garageAnimActive && g.isVeryBegin {
		g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) || len(g.touchIDs) > 0 {

			g.garageAnimActive = true
			g.isVeryBegin = false
		}
	}

	if g.garageAnimActive {
		speed := 15.0
		g.garageAnimY -= speed
		if g.garageAnimY <= 0 {
			g.garageAnimY = 0
			g.garageAnimActive = false
			g.doorAnimActive = true

			p := g.AudioContext.NewPlayerF32FromBytes(g.GarageSFX)
			p.SetVolume(1)
			p.Play()
		}
	}

	if g.doorAnimActive {
		if g.doorAnimY == 0 {
			g.lastFrame = time.Now().Add(-time.Duration(2.5 * float64(time.Second)))
		}
		speed := 1.0
		g.doorAnimY -= speed
		if g.doorAnimY < (-constants.ScreenHeight / 2) {
			g.doorAnimY = -constants.ScreenHeight / 2
			g.state = inGamePlay
			g.doorAnimActive = false
			g.BassAudio.Rewind()
			g.GuitarAudio.Rewind()
			g.DrumsAudio.Rewind()
		}
	}

}
func (g *MainScene) UpdateInGamePlay() {
	if g.Man1.IsMark {
		g.Man1.CurrentMarkTime++
		if g.Man1.CurrentMarkTime > g.Man1.MarkTime {
			g.Man1.CurrentMarkTime = 0
			g.Man1.IsMark = false
		}
	}
	if g.Man2.IsMark {
		g.Man2.CurrentMarkTime++
		if g.Man2.CurrentMarkTime > g.Man2.MarkTime {
			g.Man2.CurrentMarkTime = 0
			g.Man2.IsMark = false
		}
	}
	if g.Man3.IsMark {
		g.Man3.CurrentMarkTime++
		if g.Man3.CurrentMarkTime > g.Man3.MarkTime {
			g.Man3.CurrentMarkTime = 0
			g.Man3.IsMark = false
		}
	}

	g.Man1.Animations.Update()
	g.Man2.Animations.Update()
	g.Man3.Animations.Update()

	// Hitung delta time untuk pergerakan yang konsisten.
	dt := time.Since(g.lastFrame).Seconds()
	g.lastFrame = time.Now()

	g.skyOffset -= 20 * dt
	if g.skyOffset <= -constants.ScreenWidth {
		g.skyOffset += constants.ScreenWidth
	}

	// Majukan posisi waktu lagu.
	g.currentTick += g.ticksPerSec * dt
	// Perbarui posisi Y setiap not dan cek jika terlewat.
	highestTick := 0.0
	for _, note := range g.songChart {
		if note.Tick > highestTick {
			highestTick = note.Tick
		}
		if !note.IsActive {
			continue
		}

		// Hitung posisi Y berdasarkan seberapa jauh not dari waktu saat ini.
		// Not akan berada di hitZoneY saat note.Tick == g.currentTick.
		tickDifference := note.Tick - g.currentTick
		note.YPosition = g.hitZoneY - (tickDifference * g.noteSpeed)

		// Cek jika not terlewat (sudah melewati zona penilaian).
		if note.YPosition > (NoteY + NoteHeight) {
			note.IsActive = false

			g.scoreVal += int(g.score.missNote)
			switch note.Lane {
			case GuitarLaneId:
				g.score.guitarScore.miss += 1
				g.GuitarAudio.SetVolume(0)
				g.Man1.MarkImage = g.markMissImage
				g.Man1.IsMark = true
				g.Man1.CurrentMarkTime = 0
			case DrumsLaneId:
				g.score.drumScore.miss += 1
				g.DrumsAudio.SetVolume(0)
				g.Man3.MarkImage = g.markMissImage
				g.Man3.IsMark = true
				g.Man3.CurrentMarkTime = 0
			case BassLaneId:
				g.score.bassScore.miss += 1
				g.BassAudio.SetVolume(0)
				g.Man2.MarkImage = g.markMissImage
				g.Man2.IsMark = true
				g.Man2.CurrentMarkTime = 0
			}
		}
	}

	if highestTick+300 < g.currentTick {
		g.state = inGameFinish
	}

	// Handle input
	// Toleransi waktu untuk penilaian (dalam tick).
	perfectWindow := 10.0
	goodWindow := 20.0

	cX, cY := 0, 0
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		cX, cY = ebiten.CursorPosition()
	}
	g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
	if len(g.touchIDs) > 0 {
		cX, cY = ebiten.TouchPosition(g.touchIDs[0])
	}
	cs := image.Rect(cX, cY, cX+5, cY+5)

	g.isNoteMan1Pressed = false
	g.isNoteMan2Pressed = false
	g.isNoteMan3Pressed = false
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || cs.In(g.lanes[0].TouchRange) {
		g.isNoteMan1Pressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || cs.In(g.lanes[2].TouchRange) {
		g.isNoteMan2Pressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || cs.In(g.lanes[1].TouchRange) {
		g.isNoteMan3Pressed = true
	}

	for i, lane := range g.lanes {
		// Cek jika tombol untuk lajur ini baru saja ditekan.
		if inpututil.IsKeyJustPressed(lane.Key) || cs.In(lane.TouchRange) {
			var bestNote *Note
			minTickDiff := math.Inf(1)

			// Cari not aktif terdekat di lajur yang ditekan.
			for _, note := range g.songChart {
				if note.IsActive && int(note.Lane) == i {
					tickDiff := math.Abs(note.Tick - g.currentTick)
					if tickDiff < minTickDiff {
						minTickDiff = tickDiff
						bestNote = note
						break
					}
				}
			}

			// Jika ada not yang ditemukan dalam jangkauan.
			if bestNote != nil {
				if minTickDiff <= perfectWindow {
					g.scoreVal += int(g.score.perfectNote)
					bestNote.IsActive = false
					switch bestNote.Lane {
					case GuitarLaneId:
						g.score.guitarScore.perfect += 1
						g.GuitarAudio.SetVolume(1)
						g.Man1.MarkImage = g.markPerfectImage
						g.Man1.IsMark = true
						g.Man1.CurrentMarkTime = 0
					case DrumsLaneId:
						g.score.drumScore.perfect += 1
						g.DrumsAudio.SetVolume(1)
						g.Man3.MarkImage = g.markPerfectImage
						g.Man3.IsMark = true
						g.Man3.CurrentMarkTime = 0
					case BassLaneId:
						g.score.bassScore.perfect += 1
						g.BassAudio.SetVolume(1)
						g.Man2.MarkImage = g.markPerfectImage
						g.Man2.IsMark = true
						g.Man2.CurrentMarkTime = 0
					}
				} else if minTickDiff <= goodWindow {
					g.scoreVal += int(g.score.goodNote)
					bestNote.IsActive = false
					switch bestNote.Lane {
					case GuitarLaneId:
						g.score.guitarScore.good += 1
						g.GuitarAudio.SetVolume(1)
						g.Man1.MarkImage = g.markGoodImage
						g.Man1.IsMark = true
						g.Man1.CurrentMarkTime = 0
					case DrumsLaneId:
						g.score.drumScore.good += 1
						g.DrumsAudio.SetVolume(1)
						g.Man3.MarkImage = g.markGoodImage
						g.Man3.IsMark = true
						g.Man3.CurrentMarkTime = 0
					case BassLaneId:
						g.score.bassScore.good += 1
						g.BassAudio.SetVolume(1)
						g.Man2.MarkImage = g.markGoodImage
						g.Man2.IsMark = true
						g.Man2.CurrentMarkTime = 0
					}
				}
			}
		}
	}

	if !g.BassAudio.IsPlaying() {
		g.BassAudio.Play()
		g.GuitarAudio.Play()
		g.DrumsAudio.Play()
		g.BassAudio.SetVolume(1)
		g.GuitarAudio.SetVolume(1)
		g.DrumsAudio.SetVolume(1)
	}

}
func (g *MainScene) UpdateInGameFinish() {
	g.Man1.Animations.Update()
	g.Man2.Animations.Update()
	g.Man3.Animations.Update()
	dt := time.Since(g.lastFrame).Seconds()
	g.lastFrame = time.Now()

	g.skyOffset -= 20 * dt
	if g.skyOffset <= -constants.ScreenWidth {
		g.skyOffset += constants.ScreenWidth
	}

	// Update finish animation
	if g.finishAnimActive && !g.isFinishAnim {
		speed := 1.0
		g.finishAnimY += speed
		// Stop animation when elements reach bottom of screen
		if g.finishAnimY >= float64(constants.ScreenHeight) {
			g.finishAnimY = float64(constants.ScreenHeight)
			g.finishAnimActive = false
			g.isFinishAnim = true
		}
	}
}
func (g *MainScene) Draw(screen *ebiten.Image) {

	if !g.isLoaded {
		x := (constants.ScreenWidth - 70) / 2
		y := (constants.ScreenHeight - 30) / 2
		l := 70
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Loading %d", g.loadPercent)+"%", x, y)
		vector.StrokeLine(screen, float32(x), float32(y)+20, float32(x)+(float32(l)*float32(g.loadPercent)/100), float32(y)+20, 7, color.White, false)
		return
	}

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

	off := g.skyOffset
	for i := range 2 {
		op.GeoM.Translate(off+float64(i*constants.ScreenWidth), 0)
		screen.DrawImage(g.sky, op)
		op.GeoM.Reset()
	}

	y := g.garageAnimY
	op.GeoM.Reset()
	op.GeoM.Translate(0, y)
	screen.DrawImage(g.garageInside, op)
	op.GeoM.Reset()

	if g.doorAnimActive {
		// boys
		op.GeoM.Translate(g.Man3.X, g.Man3.Y)
		Frame, _ := g.Man3.Animations.Frame()
		screen.DrawImage(g.Man3.Image.SubImage(
			g.Man3.Sheet.HorizontalRect(Frame),
		).(*ebiten.Image), op)
		op.GeoM.Reset()
		if g.Man3.IsMark {
			op.GeoM.Translate(g.Man3.X+18, g.Man3.Y-30)
			screen.DrawImage(g.Man3.MarkImage, op)
			op.GeoM.Reset()
		}

		op.GeoM.Translate(g.Man1.X, g.Man1.Y)
		Frame, _ = g.Man1.Animations.Frame()
		screen.DrawImage(g.Man1.Image.SubImage(
			g.Man1.Sheet.HorizontalRect(Frame),
		).(*ebiten.Image), op)
		op.GeoM.Reset()
		if g.Man1.IsMark {
			op.GeoM.Translate(g.Man1.X+8, g.Man1.Y-15)
			screen.DrawImage(g.Man1.MarkImage, op)
			op.GeoM.Reset()
		}

		op.GeoM.Translate(g.Man2.X, g.Man2.Y)
		Frame, _ = g.Man2.Animations.Frame()
		screen.DrawImage(g.Man2.Image.SubImage(
			g.Man2.Sheet.HorizontalRect(Frame),
		).(*ebiten.Image), op)
		op.GeoM.Reset()
		if g.Man2.IsMark {
			op.GeoM.Translate(g.Man2.X+13, g.Man2.Y-15)
			screen.DrawImage(g.Man2.MarkImage, op)
			op.GeoM.Reset()
		}
	}

	if !g.doorAnimActive {
		op.GeoM.Translate(0, y)
	} else {
		op.GeoM.Translate(0, g.doorAnimY)
	}
	screen.DrawImage(g.garageDoor, op)
	op.GeoM.Reset()

	op.GeoM.Translate(0, y)
	screen.DrawImage(g.garageAndSky, op)
	op.GeoM.Reset()

	if g.isVeryBegin {
		titleTexts := constants.GameTitle
		var titleFontSize float64 = 70
		opt := &text.DrawOptions{}
		opt.GeoM.Translate(constants.ScreenWidth/2, 80)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = titleFontSize
		opt.PrimaryAlign = text.AlignCenter
		text.Draw(screen, titleTexts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   titleFontSize,
		}, opt)

		opt.GeoM.Reset()

		var fontSize float64 = 32
		texts := "Press Enter/Click/Touch"
		opt.GeoM.Translate(constants.ScreenWidth/2, 220)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignCenter
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()
	}

}

func (g *MainScene) DrawInGamePlay(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	off := g.skyOffset
	for i := range 2 {
		op.GeoM.Translate(off+float64(i*constants.ScreenWidth), 0)
		screen.DrawImage(g.sky, op)
		op.GeoM.Reset()
	}

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
	if g.Man3.IsMark {
		op.GeoM.Translate(g.Man3.X+18, g.Man3.Y-30)
		screen.DrawImage(g.Man3.MarkImage, op)
		op.GeoM.Reset()
	}

	op.GeoM.Translate(g.Man1.X, g.Man1.Y)
	Frame, _ = g.Man1.Animations.Frame()
	screen.DrawImage(g.Man1.Image.SubImage(
		g.Man1.Sheet.HorizontalRect(Frame),
	).(*ebiten.Image), op)
	op.GeoM.Reset()
	if g.Man1.IsMark {
		op.GeoM.Translate(g.Man1.X+8, g.Man1.Y-15)
		screen.DrawImage(g.Man1.MarkImage, op)
		op.GeoM.Reset()
	}

	op.GeoM.Translate(g.Man2.X, g.Man2.Y)
	Frame, _ = g.Man2.Animations.Frame()
	screen.DrawImage(g.Man2.Image.SubImage(
		g.Man2.Sheet.HorizontalRect(Frame),
	).(*ebiten.Image), op)
	op.GeoM.Reset()
	if g.Man2.IsMark {
		op.GeoM.Translate(g.Man2.X+13, g.Man2.Y-15)
		screen.DrawImage(g.Man2.MarkImage, op)
		op.GeoM.Reset()
	}

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

	// Gambar tombol statis di zona penilaian.
	for i, lane := range g.lanes {
		x := firstNoteX + (noteLineWidth * i)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), g.hitZoneY-float64(g.noteImage.Bounds().Dy())/2)
		op.ColorScale.ScaleWithColor(lane.Color)
		switch i {
		case int(GuitarLaneId):
			if g.isNoteMan1Pressed {
				op.ColorScale.ScaleAlpha(0.8)
			} else {
				op.ColorScale.ScaleAlpha(0.4) // Buat lebih transparan
			}
		case int(BassLaneId):
			if g.isNoteMan2Pressed {
				op.ColorScale.ScaleAlpha(0.8)
			} else {
				op.ColorScale.ScaleAlpha(0.4) // Buat lebih transparan
			}
		case int(DrumsLaneId):
			if g.isNoteMan3Pressed {
				op.ColorScale.ScaleAlpha(0.8)
			} else {
				op.ColorScale.ScaleAlpha(0.4) // Buat lebih transparan
			}
		}
		screen.DrawImage(g.noteImage, op)
	}

	// Gambar setiap not yang masih aktif.
	for _, note := range g.songChart {
		if !note.IsActive {
			continue
		}
		if note.YPosition < NoteY {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		// Hitung posisi X berdasarkan lajur not.
		noteX := firstNoteX + (noteLineWidth * float64(note.Lane))
		// Pusatkan not di tengah lajur.
		op.GeoM.Translate(noteX, note.YPosition-float64(g.noteImage.Bounds().Dy())/2)

		// Beri warna not sesuai dengan lajurnya.
		op.ColorScale.ScaleWithColor(g.lanes[note.Lane].Color)

		screen.DrawImage(g.noteImage, op)
	}
}
func (g *MainScene) DrawInGameFinish(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}

	off := g.skyOffset
	for i := range 2 {
		op.GeoM.Translate(off+float64(i*constants.ScreenWidth), 0)
		screen.DrawImage(g.sky, op)
		op.GeoM.Reset()
	}

	// garage with animation offset
	op.GeoM.Translate(0, g.finishAnimY)
	screen.DrawImage(g.garageInside, op)
	op.GeoM.Reset()

	op.GeoM.Translate(0, g.finishAnimY)
	screen.DrawImage(g.garage, op)
	op.GeoM.Reset()

	// boys with animation offset
	op.GeoM.Translate(g.Man3.X, g.Man3.Y+g.finishAnimY)
	Frame, _ := g.Man3.Animations.Frame()
	op.ColorScale.Reset()
	screen.DrawImage(g.Man3.Image.SubImage(
		g.Man3.Sheet.HorizontalRect(Frame),
	).(*ebiten.Image), op)
	op.GeoM.Reset()

	op.GeoM.Translate(g.Man1.X, g.Man1.Y+g.finishAnimY)
	Frame, _ = g.Man1.Animations.Frame()
	op.ColorScale.Reset()
	screen.DrawImage(g.Man1.Image.SubImage(
		g.Man1.Sheet.HorizontalRect(Frame),
	).(*ebiten.Image), op)
	op.GeoM.Reset()

	op.GeoM.Translate(g.Man2.X, g.Man2.Y+g.finishAnimY)
	Frame, _ = g.Man2.Animations.Frame()
	op.ColorScale.Reset()
	screen.DrawImage(g.Man2.Image.SubImage(
		g.Man2.Sheet.HorizontalRect(Frame),
	).(*ebiten.Image), op)
	op.GeoM.Reset()

	// panel
	if g.isFinishAnim {
		x := 180.0
		y := 70.0
		scale := 1.0 / 1.3

		// instrument
		op.GeoM.Translate(0, 0)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x, y)
		screen.DrawImage(g.noteMan1Image, op)
		op.GeoM.Reset()

		x += 135
		op.GeoM.Translate(0, 0)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x, y)
		screen.DrawImage(g.noteMan3Image, op)
		op.GeoM.Reset()

		x += 145
		op.GeoM.Translate(0, 0)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x, y)
		screen.DrawImage(g.noteMan2Image, op)
		op.GeoM.Reset()

		// mark
		x = 70
		y = 155
		op.GeoM.Translate(x, y)
		screen.DrawImage(g.markPerfectImage, op)
		op.GeoM.Reset()

		y += 65
		op.GeoM.Translate(x, y)
		screen.DrawImage(g.markGoodImage, op)
		op.GeoM.Reset()

		y += 65
		op.GeoM.Translate(x, y)
		screen.DrawImage(g.markMissImage, op)
		op.GeoM.Reset()

		// scoring
		x1 := 200.0
		x2 := 335.0
		x3 := 475.0
		y1 := 160.0
		y2 := 225.0
		y3 := 285.0
		var fontSize float64 = 32
		opt := &text.DrawOptions{}

		// guitar
		texts := fmt.Sprintf("X%d", g.score.guitarScore.perfect)
		opt.GeoM.Translate(x1, y1)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignStart
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()
		texts = fmt.Sprintf("X%d", g.score.guitarScore.good)
		opt.GeoM.Translate(x1, y2)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignStart
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()
		texts = fmt.Sprintf("X%d", g.score.guitarScore.miss)
		opt.GeoM.Translate(x1, y3)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignStart
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()

		// drums
		texts = fmt.Sprintf("X%d", g.score.drumScore.perfect)
		opt.GeoM.Translate(x2, y1)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignStart
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()
		texts = fmt.Sprintf("X%d", g.score.drumScore.good)
		opt.GeoM.Translate(x2, y2)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignStart
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()
		texts = fmt.Sprintf("X%d", g.score.drumScore.miss)
		opt.GeoM.Translate(x2, y3)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignStart
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()

		// bass
		texts = fmt.Sprintf("X%d", g.score.bassScore.perfect)
		opt.GeoM.Translate(x3, y1)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignStart
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()
		texts = fmt.Sprintf("X%d", g.score.bassScore.good)
		opt.GeoM.Translate(x3, y2)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignStart
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()
		texts = fmt.Sprintf("X%d", g.score.bassScore.miss)
		opt.GeoM.Translate(x3, y3)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignStart
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()

		opt.GeoM.Reset()
		texts = fmt.Sprintf("X%d", g.score.bassScore.miss)
		opt.GeoM.Translate(x3, y3)
		opt.ColorScale.ScaleWithColor(color.Black)
		opt.LineSpacing = fontSize * 1.5
		opt.PrimaryAlign = text.AlignStart
		text.Draw(screen, texts, &text.GoTextFace{
			Source: g.fontSource,
			Size:   fontSize,
		}, opt)
		opt.GeoM.Reset()
	}
}

var _ Scene = (*MainScene)(nil)
