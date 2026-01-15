package systems

import rl "github.com/gen2brain/raylib-go/raylib"

type Input struct {
	MoveToX       float32
	MoveToY       float32
	HasMoveTarget bool
	Attack        bool
	Skill1        bool
	Skill2        bool
	Skill3        bool
	Skill4        bool
}

func UpdateInput(camera *Camera) *Input {
	moveToX := float32(0)
	moveToY := float32(0)
	hasMoveTarget := false

	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		mouseX, mouseY := GetMousePosition()
		moveToX, moveToY = ScreenToWorld(mouseX, mouseY, camera)
		hasMoveTarget = true
	}

	return &Input{
		MoveToX:       moveToX,
		MoveToY:       moveToY,
		HasMoveTarget: hasMoveTarget,
		Attack:        rl.IsMouseButtonPressed(rl.MouseLeftButton),
		Skill1:        rl.IsKeyPressed(rl.KeyQ),
		Skill2:        rl.IsKeyPressed(rl.KeyW),
		Skill3:        rl.IsKeyPressed(rl.KeyE),
		Skill4:        rl.IsKeyPressed(rl.KeyR),
	}
}

func GetMousePosition() (float32, float32) {
	pos := rl.GetMousePosition()
	return pos.X, pos.Y
}
