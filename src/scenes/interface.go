package scenes

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rizalmf/old-boys/src/lang"
)

type SceneId uint

const (
	GameSceneId SceneId = iota
	SplashSceneId
	MenuSceneId
	ExitSceneId
)

type Scene interface {
	Update() SceneId
	Draw(screen *ebiten.Image)
	FirstLoad()
	ExportProperties() (prop Properties)
	OnEnter(prop Properties)
	OnExit()
	IsLoaded() bool
}

type Properties struct {
	Vehicle string
	Lang    lang.Lang
}
