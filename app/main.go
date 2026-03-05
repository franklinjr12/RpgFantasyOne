package main

import (
	"singlefantasy/app/assets"
	"singlefantasy/app/game"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(game.WindowWidth, game.WindowHeight, "Single Fantasy")
	defer func() {
		assets.Get().UnloadAll()
		if rl.IsAudioDeviceReady() {
			rl.CloseAudioDevice()
		}
		rl.CloseWindow()
	}()

	rl.InitAudioDevice()
	rl.SetTargetFPS(60)

	g := game.NewGame()

	accumulator := float32(0)

	for !rl.WindowShouldClose() {
		frameTime := rl.GetFrameTime()
		if frameTime > game.MaxFrameTime {
			frameTime = game.MaxFrameTime
		}

		g.UpdateFrame()

		accumulator += frameTime

		updateSteps := 0
		for accumulator >= game.FixedDeltaTime {
			g.UpdateFixed(game.FixedDeltaTime)
			accumulator -= game.FixedDeltaTime
			updateSteps++

			if updateSteps >= game.MaxUpdateSteps {
				accumulator = 0
				break
			}
		}

		g.SetFrameDiagnostics(frameTime, updateSteps)
		g.Draw()
	}
}
