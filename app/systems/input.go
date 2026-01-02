package systems

import rl "github.com/gen2brain/raylib-go/raylib"

type Input struct {
	MoveUp    bool
	MoveDown  bool
	MoveLeft  bool
	MoveRight bool
	Attack    bool
	Skill1    bool
	Skill2    bool
	Skill3    bool
}

func UpdateInput() *Input {
	return &Input{
		MoveUp:    rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp),
		MoveDown:  rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown),
		MoveLeft:  rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft),
		MoveRight: rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight),
		Attack:    rl.IsMouseButtonPressed(rl.MouseLeftButton),
		Skill1:    rl.IsKeyPressed(rl.KeyQ),
		Skill2:    rl.IsKeyPressed(rl.KeyE),
		Skill3:    rl.IsKeyPressed(rl.KeyR),
	}
}

func GetMousePosition() (float32, float32) {
	pos := rl.GetMousePosition()
	return pos.X, pos.Y
}
