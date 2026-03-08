package systems

import rl "github.com/gen2brain/raylib-go/raylib"

type Input struct {
	MoveToX       float32
	MoveToY       float32
	HasMoveTarget bool
	CursorWorldX  float32
	CursorWorldY  float32
	Attack        bool
	Skill1        bool
	Skill2        bool
	Skill3        bool
	Skill4        bool
}

func UpdateInput(camera *Camera) *Input {
	mouseX, mouseY := GetMousePosition()
	cursorWorldX, cursorWorldY := ScreenToWorldIso(mouseX, mouseY, camera)

	moveToX := float32(0)
	moveToY := float32(0)
	hasMoveTarget := false

	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		moveToX = cursorWorldX
		moveToY = cursorWorldY
		hasMoveTarget = true
	}

	return &Input{
		MoveToX:       moveToX,
		MoveToY:       moveToY,
		HasMoveTarget: hasMoveTarget,
		CursorWorldX:  cursorWorldX,
		CursorWorldY:  cursorWorldY,
		Attack:        rl.IsMouseButtonPressed(rl.MouseLeftButton),
		Skill1:        rl.IsKeyPressed(rl.KeyQ) || rl.IsKeyPressed(rl.KeyOne),
		Skill2:        rl.IsKeyPressed(rl.KeyW) || rl.IsKeyPressed(rl.KeyTwo),
		Skill3:        rl.IsKeyPressed(rl.KeyE) || rl.IsKeyPressed(rl.KeyThree),
		Skill4:        rl.IsKeyPressed(rl.KeyR) || rl.IsKeyPressed(rl.KeyFour),
	}
}

func GetMousePosition() (float32, float32) {
	pos := rl.GetMousePosition()
	return pos.X, pos.Y
}
