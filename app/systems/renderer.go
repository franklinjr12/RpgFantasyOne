package systems

import (
	"math"

	"singlefantasy/app/gameobjects"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var TerrainColorNormalRGBA = rl.NewColor(128, 128, 128, 255)
var TerrainColorBossRGBA = rl.NewColor(64, 64, 64, 255)
var PlayerColorRGBA = rl.NewColor(0, 0, 255, 255)
var EnemyColorRGBA = rl.NewColor(255, 0, 0, 255)
var EliteColorRGBA = rl.NewColor(255, 136, 0, 255)
var BossColorRGBA = rl.NewColor(136, 0, 255, 255)
var ProjectileColorRGBA = rl.NewColor(255, 255, 0, 255)

type Camera struct {
	X      float32
	Y      float32
	Width  float32
	Height float32
}

const (
	CameraWidth  = 1600
	CameraHeight = 900
)

func NewCamera() *Camera {
	return &Camera{
		X:      0,
		Y:      0,
		Width:  CameraWidth,
		Height: CameraHeight,
	}
}

func WorldToScreen(worldX, worldY float32, camera *Camera) (float32, float32) {
	screenX := worldX - camera.X
	screenY := worldY - camera.Y
	return screenX, screenY
}

func ScreenToWorld(screenX, screenY float32, camera *Camera) (float32, float32) {
	worldX := screenX + camera.X
	worldY := screenY + camera.Y
	return worldX, worldY
}

func IsometricToScreen(isoX, isoY float32) (float32, float32) {
	screenX := (isoX - isoY) * 0.5
	screenY := (isoX + isoY) * 0.25
	return screenX, screenY
}

func ScreenToIsometric(screenX, screenY float32) (float32, float32) {
	isoX := screenX + screenY*2
	isoY := -screenX + screenY*2
	return isoX, isoY
}

func UpdateCamera(camera *Camera, playerX, playerY, worldWidth, worldHeight float32) {
	targetX := playerX - camera.Width/2
	targetY := playerY - camera.Height/2

	if targetX < 0 {
		targetX = 0
	}
	if targetY < 0 {
		targetY = 0
	}
	if targetX+camera.Width > worldWidth {
		targetX = worldWidth - camera.Width
	}
	if targetY+camera.Height > worldHeight {
		targetY = worldHeight - camera.Height
	}

	camera.X = targetX
	camera.Y = targetY
}

func DrawPlayer(player *gameobjects.Player, camera *Camera) {
	screenX, screenY := WorldToScreen(player.X, player.Y, camera)

	color := PlayerColorRGBA
	if player.HitFlashTimer > 0 {
		color = rl.Red
	}

	rl.DrawRectangleRec(rl.NewRectangle(screenX, screenY, player.Width, player.Height), color)

	healthBarWidth := player.Width
	healthBarHeight := float32(5)
	healthBarX := screenX
	healthBarY := screenY - healthBarHeight - 3
	healthPercent := float32(player.Health) / float32(player.MaxHealth)
	if healthPercent < 0 {
		healthPercent = 0
	}
	if healthPercent > 1 {
		healthPercent = 1
	}

	rl.DrawRectangleRec(rl.NewRectangle(healthBarX, healthBarY, healthBarWidth, healthBarHeight), rl.Red)
	rl.DrawRectangleRec(rl.NewRectangle(healthBarX, healthBarY, healthBarWidth*healthPercent, healthBarHeight), rl.Green)
}

func DrawEnemy(enemy *gameobjects.Enemy, camera *Camera) {
	if !enemy.Alive {
		return
	}

	screenX, screenY := WorldToScreen(enemy.X, enemy.Y, camera)

	var color rl.Color
	if enemy.IsElite {
		color = EliteColorRGBA
	} else {
		color = EnemyColorRGBA
	}

	if enemy.HitFlashTimer > 0 {
		color = rl.Orange
	} else if enemy.AttackFlashTimer > 0 {
		color = rl.Yellow
	}

	rl.DrawRectangleRec(rl.NewRectangle(screenX, screenY, enemy.Width, enemy.Height), color)

	healthBarWidth := enemy.Width
	healthBarHeight := float32(5)
	healthBarX := screenX
	healthBarY := screenY - healthBarHeight - 3
	healthPercent := float32(enemy.Health) / float32(enemy.MaxHealth)
	if healthPercent < 0 {
		healthPercent = 0
	}
	if healthPercent > 1 {
		healthPercent = 1
	}

	rl.DrawRectangleRec(rl.NewRectangle(healthBarX, healthBarY, healthBarWidth, healthBarHeight), rl.Red)
	rl.DrawRectangleRec(rl.NewRectangle(healthBarX, healthBarY, healthBarWidth*healthPercent, healthBarHeight), rl.Green)
}

func DrawRoom(x, y, width, height float32, isBoss bool, camera *Camera) {
	screenX, screenY := WorldToScreen(x, y, camera)

	var color rl.Color
	if isBoss {
		color = TerrainColorBossRGBA
	} else {
		color = TerrainColorNormalRGBA
	}

	rl.DrawRectangleRec(rl.NewRectangle(screenX, screenY, width, height), color)
	rl.DrawRectangleLinesEx(rl.NewRectangle(screenX, screenY, width, height), 2, rl.Black)
}

func DrawProjectile(x, y, radius float32, camera *Camera) {
	screenX, screenY := WorldToScreen(x, y, camera)
	rl.DrawCircle(int32(screenX+radius), int32(screenY+radius), radius, ProjectileColorRGBA)
}

func DrawBoss(boss *gameobjects.Boss, camera *Camera) {
	if !boss.Alive {
		return
	}

	screenX, screenY := WorldToScreen(boss.X, boss.Y, camera)

	color := BossColorRGBA
	if boss.HitFlashTimer > 0 {
		color = rl.Orange
	} else if boss.AttackFlashTimer > 0 {
		color = rl.Yellow
	}

	if boss.TelegraphTimer > 0 {
		rl.DrawCircleLines(int32(screenX+boss.Width/2), int32(screenY+boss.Height/2), 100, rl.Red)
	}

	rl.DrawRectangleRec(rl.NewRectangle(screenX, screenY, boss.Width, boss.Height), color)

	healthBarWidth := boss.Width
	healthBarHeight := float32(8)
	healthBarX := screenX
	healthBarY := screenY - healthBarHeight - 5
	healthPercent := float32(boss.Health) / float32(boss.MaxHealth)
	if healthPercent < 0 {
		healthPercent = 0
	}
	if healthPercent > 1 {
		healthPercent = 1
	}

	rl.DrawRectangleRec(rl.NewRectangle(healthBarX, healthBarY, healthBarWidth, healthBarHeight), rl.Red)
	rl.DrawRectangleRec(rl.NewRectangle(healthBarX, healthBarY, healthBarWidth*healthPercent, healthBarHeight), rl.Green)
}

func DrawBossProjectile(x, y, radius float32, camera *Camera) {
	screenX, screenY := WorldToScreen(x, y, camera)
	rl.DrawCircle(int32(screenX+radius), int32(screenY+radius), radius, rl.Purple)
}

func GetDistance(x1, y1, x2, y2 float32) float32 {
	dx := x2 - x1
	dy := y2 - y1
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}
