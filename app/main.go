package main

import (
	"singlefantasy/app/game"
	"singlefantasy/app/systems"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(game.WindowWidth, game.WindowHeight, "Single Fantasy")
	defer rl.CloseWindow()

	rl.SetTargetFPS(game.TargetFPS)

	systems.LoadSpriteSheet()

	g := game.NewGame()

	for !rl.WindowShouldClose() {
		deltaTime := rl.GetFrameTime()

		if g.State == game.RunStateMenu {
			g.HandleMenuInput()
		} else if g.State == game.RunStateRewardSelection {
			g.HandleRewardSelectionInput()
		} else if g.State == game.RunStateVictory || g.State == game.RunStateDefeat {
			g.HandleGameOverInput()
		} else {
			g.Update(deltaTime)
		}

		g.Draw()
	}
}
