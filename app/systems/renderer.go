package systems

import (
	"fmt"
	"math"
	"sync"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	spriteSheet     rl.Texture2D
	spriteSheetOnce sync.Once
	spriteWidth     = float32(72)
	spriteHeight    = float32(72)
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

type uiSkillSlot struct {
	Skill    *gamedata.Skill
	KeyLabel string
	Rect     rl.Rectangle
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

	var sourceRect rl.Rectangle
	switch player.Class.Type {
	case 0: // ClassTypeMelee
		sourceRect = getSpriteSourceRect(1, 2) // row 2, column 3 (0-indexed: row 1, col 2)
	case 1: // ClassTypeRanged
		sourceRect = getSpriteSourceRect(0, 2) // row 1, column 3 (0-indexed: row 0, col 2)
	case 2: // ClassTypeCaster
		sourceRect = getSpriteSourceRect(0, 0) // row 1, column 1 (0-indexed: row 0, col 0)
	default:
		sourceRect = getSpriteSourceRect(0, 0)
	}

	destRect := rl.NewRectangle(screenX, screenY, player.Width, player.Height)

	tint := rl.White
	if player.HitFlashTimer > 0 {
		tint = rl.Red
	}

	rl.DrawTexturePro(GetSpriteSheet(), sourceRect, destRect, rl.NewVector2(0, 0), 0, tint)

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

	sourceRect := getSpriteSourceRect(2, 4) // row 3, column 5 (0-indexed: row 2, col 4)
	destRect := rl.NewRectangle(screenX, screenY, enemy.Width, enemy.Height)

	tint := rl.White
	if enemy.HitFlashTimer > 0 {
		tint = rl.Orange
	} else if enemy.AttackFlashTimer > 0 {
		tint = rl.Yellow
	} else if enemy.IsElite {
		tint = rl.NewColor(255, 200, 0, 255)
	}

	rl.DrawTexturePro(GetSpriteSheet(), sourceRect, destRect, rl.NewVector2(0, 0), 0, tint)

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

	sourceRect := getSpriteSourceRect(2, 4) // row 3, column 5 (0-indexed: row 2, col 4)
	destRect := rl.NewRectangle(screenX, screenY, boss.Width, boss.Height)

	tint := rl.NewColor(136, 0, 255, 255)
	if boss.HitFlashTimer > 0 {
		tint = rl.Orange
	} else if boss.AttackFlashTimer > 0 {
		tint = rl.Yellow
	}

	if boss.TelegraphTimer > 0 {
		rl.DrawCircleLines(int32(screenX+boss.Width/2), int32(screenY+boss.Height/2), 100, rl.Red)
	}

	rl.DrawTexturePro(GetSpriteSheet(), sourceRect, destRect, rl.NewVector2(0, 0), 0, tint)

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

func LoadSpriteSheet() {
	spriteSheetOnce.Do(func() {
		spriteSheet = rl.LoadTexture("resources/sprites/Basic Humanoid Sprites 4x.png")
	})
}

func GetSpriteSheet() rl.Texture2D {
	LoadSpriteSheet()
	return spriteSheet
}

func getSpriteSourceRect(row, col int) rl.Rectangle {
	x := float32(col) * spriteWidth
	y := float32(row) * spriteHeight
	return rl.NewRectangle(x, y, spriteWidth, spriteHeight)
}

func DrawSkillBar(player *gameobjects.Player) {
	if player == nil {
		return
	}

	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	slotCount := 4
	slotWidth := float32(80)
	slotHeight := float32(80)
	slotSpacing := float32(10)
	barPadding := float32(10)

	totalSlotsWidth := float32(slotCount)*slotWidth + float32(slotCount-1)*slotSpacing
	barWidth := totalSlotsWidth + barPadding*2
	barHeight := slotHeight + barPadding*2

	barX := (screenWidth - barWidth) / 2
	barY := screenHeight - barHeight - 20

	barRect := rl.NewRectangle(barX, barY, barWidth, barHeight)
	rl.DrawRectangleRec(barRect, rl.NewColor(0, 0, 0, 180))

	keyLabels := []string{"Q", "W", "E", "R"}

	var slots []uiSkillSlot

	for i := 0; i < slotCount; i++ {
		slotX := barX + barPadding + float32(i)*(slotWidth+slotSpacing)
		slotY := barY + barPadding
		slotRect := rl.NewRectangle(slotX, slotY, slotWidth, slotHeight)

		var skill *gamedata.Skill
		if i < len(player.Skills) {
			skill = player.Skills[i]
		}

		keyLabel := ""
		if i < len(keyLabels) {
			keyLabel = keyLabels[i]
		}

		slots = append(slots, uiSkillSlot{
			Skill:    skill,
			KeyLabel: keyLabel,
			Rect:     slotRect,
		})
	}

	for _, slot := range slots {
		slotX := slot.Rect.X
		slotY := slot.Rect.Y
		slotWidth := slot.Rect.Width
		slotHeight := slot.Rect.Height

		rl.DrawRectangleRec(slot.Rect, rl.NewColor(50, 50, 50, 255))
		rl.DrawRectangleLinesEx(slot.Rect, 2, rl.NewColor(200, 200, 200, 255))

		iconMargin := float32(8)
		iconRect := rl.NewRectangle(slotX+iconMargin, slotY+iconMargin, slotWidth-2*iconMargin, slotHeight-2*iconMargin)
		rl.DrawRectangleRec(iconRect, rl.NewColor(80, 80, 80, 255))

		if slot.KeyLabel != "" {
			textWidth := rl.MeasureText(slot.KeyLabel, 20)
			textX := int32(slotX + 5)
			textY := int32(slotY + slotHeight - 22)
			rl.DrawRectangle(textX-2, textY-2, int32(textWidth)+4, 24, rl.NewColor(0, 0, 0, 180))
			rl.DrawText(slot.KeyLabel, textX, textY, 20, rl.RayWhite)
		}

		if slot.Skill != nil {
			remaining := slot.Skill.RemainingCooldown()
			if remaining > 0 && slot.Skill.Cooldown > 0 {
				ratio := remaining / slot.Skill.Cooldown
				if ratio < 0 {
					ratio = 0
				}
				if ratio > 1 {
					ratio = 1
				}

				overlayHeight := slotHeight * ratio
				overlayRect := rl.NewRectangle(slotX, slotY+slotHeight-overlayHeight, slotWidth, overlayHeight)
				rl.DrawRectangleRec(overlayRect, rl.NewColor(0, 0, 0, 150))

				text := fmt.Sprintf("%.1f", remaining)
				textWidth := rl.MeasureText(text, 18)
				textX := int32(slotX + (slotWidth-float32(textWidth))/2)
				textY := int32(slotY + slotHeight/2 - 9)
				rl.DrawText(text, textX, textY, 18, rl.RayWhite)
			}
		}
	}
}
