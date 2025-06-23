package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/rizalmf/old-boys/src"
)

func main() {

	g := src.NewGame()

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal("init failed", err)
	}

	// <-make(chan bool)
}
