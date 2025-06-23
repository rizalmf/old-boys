package src

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rizalmf/old-boys/src/constants"
	"github.com/rizalmf/old-boys/src/scenes"
)

type Game struct {
	sceneMap      map[scenes.SceneId]scenes.Scene
	activeSceneId scenes.SceneId
}

func NewGame() *Game {
	ebiten.SetVsyncEnabled(true)
	ebiten.SetWindowSize(constants.ScreenWidth, constants.ScreenHeight)
	ebiten.SetWindowTitle(constants.GameTitle)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(constants.TPS)

	g := &Game{
		sceneMap: map[scenes.SceneId]scenes.Scene{
			scenes.GameSceneId: scenes.NewGameScene(),
		},
		activeSceneId: scenes.GameSceneId,
	}
	g.sceneMap[g.activeSceneId].FirstLoad()

	return g
}

func (g *Game) Update() error {
	nextSceneId := g.sceneMap[g.activeSceneId].Update()

	// exit
	if nextSceneId == scenes.ExitSceneId {
		g.sceneMap[g.activeSceneId].OnExit()
	}

	// switch scene
	if nextSceneId != g.activeSceneId {
		nextScene := g.sceneMap[nextSceneId]
		// if not loaded? then load in
		if !nextScene.IsLoaded() {
			nextScene.FirstLoad()
		}
		nextScene.OnEnter(g.sceneMap[g.activeSceneId].ExportProperties())

		g.sceneMap[g.activeSceneId].OnExit()
	}

	g.activeSceneId = nextSceneId

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.sceneMap[g.activeSceneId].Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return constants.ScreenWidth, constants.ScreenHeight
}

func (g *Game) ExportProperties() (prop scenes.Properties) {
	return g.sceneMap[g.activeSceneId].ExportProperties()
}
