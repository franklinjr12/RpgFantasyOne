package systems

import (
	"fmt"
	"math"

	"singlefantasy/app/assets"
	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	spriteWidth  = float32(72)
	spriteHeight = float32(72)
)

var TerrainColorNormalRGBA = rl.NewColor(110, 122, 126, 255)
var TerrainColorBossRGBA = rl.NewColor(78, 88, 102, 255)
var PlayerColorRGBA = rl.NewColor(0, 0, 255, 255)
var EnemyColorRGBA = rl.NewColor(255, 0, 0, 255)
var EliteColorRGBA = rl.NewColor(255, 136, 0, 255)
var BossColorRGBA = rl.NewColor(136, 0, 255, 255)
var ProjectileColorRGBA = rl.NewColor(255, 255, 0, 255)

type Camera struct {
	X         float32
	Y         float32
	TargetX   float32
	TargetY   float32
	Width     float32
	Height    float32
	Smoothing float32
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

const (
	HumanoidSpriteSheetAssetKey = assets.TextureHumanoidSpriteSheet
)

func NewCamera() *Camera {
	return &Camera{
		X:         0,
		Y:         0,
		TargetX:   0,
		TargetY:   0,
		Width:     CameraWidth,
		Height:    CameraHeight,
		Smoothing: 0.18,
	}
}

func UpdateCamera(camera *Camera, playerX, playerY, worldWidth, worldHeight float32) {
	if camera == nil {
		return
	}

	maxX := worldWidth - camera.Width
	if maxX < 0 {
		maxX = 0
	}
	maxY := worldHeight - camera.Height
	if maxY < 0 {
		maxY = 0
	}

	targetX := clamp(playerX-camera.Width/2, 0, maxX)
	targetY := clamp(playerY-camera.Height/2, 0, maxY)
	camera.TargetX = targetX
	camera.TargetY = targetY

	smoothing := camera.Smoothing
	if smoothing <= 0 || smoothing > 1 {
		smoothing = 1
	}

	camera.X += (targetX - camera.X) * smoothing
	camera.Y += (targetY - camera.Y) * smoothing

	if nearlyEqual(camera.X, targetX, 0.05) {
		camera.X = targetX
	}
	if nearlyEqual(camera.Y, targetY, 0.05) {
		camera.Y = targetY
	}

	camera.X = clamp(camera.X, 0, maxX)
	camera.Y = clamp(camera.Y, 0, maxY)
}

func DrawPlayer(player *gameobjects.Player, camera *Camera) {
	screenX, screenY := actorScreenRect(player.PosX, player.PosY, player.Hitbox.Width, player.Hitbox.Height, camera)

	var sourceRect rl.Rectangle
	switch player.Class.Type {
	case 0:
		sourceRect = getSpriteSourceRect(1, 2)
	case 1:
		sourceRect = getSpriteSourceRect(0, 2)
	case 2:
		sourceRect = getSpriteSourceRect(0, 0)
	default:
		sourceRect = getSpriteSourceRect(0, 0)
	}

	destRect := rl.NewRectangle(screenX, screenY, player.Hitbox.Width, player.Hitbox.Height)

	tint := rl.White
	if player.HitFlashTimer > 0 {
		tint = rl.Red
	}

	drawTextureOrRect(GetSpriteSheet(), sourceRect, destRect, tint, rl.Blue)
	drawHealthBar(destRect, float32(player.HP)/float32(player.MaxHP), 5)
}

func DrawEnemy(enemy *gameobjects.Enemy, camera *Camera) {
	if !enemy.IsAlive() {
		return
	}

	screenX, screenY := actorScreenRect(enemy.PosX, enemy.PosY, enemy.Hitbox.Width, enemy.Hitbox.Height, camera)

	sourceRect := getSpriteSourceRect(2, 4)
	destRect := rl.NewRectangle(screenX, screenY, enemy.Hitbox.Width, enemy.Hitbox.Height)

	tint := rl.White
	if enemy.HitFlashTimer > 0 {
		tint = rl.Orange
	} else if enemy.AttackFlashTimer > 0 {
		tint = rl.Yellow
	} else if enemy.IsElite {
		tint = rl.NewColor(255, 200, 0, 255)
	}

	drawTextureOrRect(GetSpriteSheet(), sourceRect, destRect, tint, rl.Red)
	drawHealthBar(destRect, float32(enemy.HP)/float32(enemy.MaxHP), 5)
}

func DrawRoom(room *world.Room, camera *Camera) {
	if room == nil {
		return
	}

	corners := projectAABBBase(world.AABB{X: room.X, Y: room.Y, Width: room.Width, Height: room.Height}, camera)

	floorColor := TerrainColorNormalRGBA
	if room.IsBoss() {
		floorColor = TerrainColorBossRGBA
	}
	wallColor := shadeColor(floorColor, -26)
	outlineColor := shadeColor(floorColor, -65)

	drawIsoQuad(corners, floorColor)

	wallHeight := float32(46)
	northTop := liftIso(corners[0], wallHeight)
	eastTop := liftIso(corners[1], wallHeight)
	southTop := liftIso(corners[2], wallHeight)
	westTop := liftIso(corners[3], wallHeight)

	drawIsoQuad([4]rl.Vector2{eastTop, southTop, corners[2], corners[1]}, wallColor)
	drawIsoQuad([4]rl.Vector2{westTop, southTop, corners[2], corners[3]}, shadeColor(wallColor, -12))
	drawIsoOutline(corners, outlineColor)

	for _, obstacle := range room.Obstacles {
		DrawObstacle(obstacle, camera)
	}
	for _, door := range room.Doors {
		DrawDoor(door, camera)
	}

	drawIsoOutline([4]rl.Vector2{northTop, eastTop, southTop, westTop}, shadeColor(outlineColor, 10))
}

func DrawObstacle(obstacle world.AABB, camera *Camera) {
	base := projectAABBBase(obstacle, camera)
	lift := float32(28)
	top := [4]rl.Vector2{liftIso(base[0], lift), liftIso(base[1], lift), liftIso(base[2], lift), liftIso(base[3], lift)}

	topColor := rl.NewColor(123, 123, 123, 255)
	rightColor := rl.NewColor(98, 98, 98, 255)
	leftColor := rl.NewColor(84, 84, 84, 255)
	outline := rl.NewColor(30, 30, 30, 255)

	drawIsoQuad([4]rl.Vector2{top[1], top[2], base[2], base[1]}, rightColor)
	drawIsoQuad([4]rl.Vector2{top[2], top[3], base[3], base[2]}, leftColor)
	drawIsoQuad(top, topColor)
	drawIsoOutline(top, outline)
}

func DrawDoor(door *world.Door, camera *Camera) {
	if door == nil {
		return
	}

	base := projectAABBBase(door.Bounds, camera)
	lift := float32(36)
	top := [4]rl.Vector2{liftIso(base[0], lift), liftIso(base[1], lift), liftIso(base[2], lift), liftIso(base[3], lift)}

	topColor := rl.NewColor(48, 168, 102, 255)
	sideColor := rl.NewColor(38, 132, 80, 255)
	if door.Locked {
		topColor = rl.NewColor(168, 82, 68, 255)
		sideColor = rl.NewColor(130, 62, 52, 255)
	}

	drawIsoQuad([4]rl.Vector2{top[1], top[2], base[2], base[1]}, sideColor)
	drawIsoQuad([4]rl.Vector2{top[2], top[3], base[3], base[2]}, shadeColor(sideColor, -12))
	drawIsoQuad(top, topColor)
	drawIsoOutline(top, shadeColor(topColor, -40))
}

func DrawProjectile(x, y, radius float32, camera *Camera) {
	screenX, screenY := WorldToScreenIso(x, y, camera)
	rl.DrawCircle(int32(screenX), int32(screenY), radius, ProjectileColorRGBA)
}

func DrawBoss(boss *gameobjects.Boss, camera *Camera) {
	if !boss.IsAlive() {
		return
	}

	screenX, screenY := actorScreenRect(boss.PosX, boss.PosY, boss.Hitbox.Width, boss.Hitbox.Height, camera)

	sourceRect := getSpriteSourceRect(2, 4)
	destRect := rl.NewRectangle(screenX, screenY, boss.Hitbox.Width, boss.Hitbox.Height)

	tint := rl.NewColor(136, 0, 255, 255)
	if boss.HitFlashTimer > 0 {
		tint = rl.Orange
	} else if boss.AttackFlashTimer > 0 {
		tint = rl.Yellow
	}

	if boss.TelegraphTimer > 0 {
		centerX, centerY := boss.Center()
		tx, ty := WorldToScreenIso(centerX, centerY, camera)
		rl.DrawCircleLines(int32(tx), int32(ty), 70, rl.Red)
	}

	drawTextureOrRect(GetSpriteSheet(), sourceRect, destRect, tint, rl.Purple)
	drawHealthBar(destRect, float32(boss.HP)/float32(boss.MaxHP), 8)
}

func DrawBossProjectile(x, y, radius float32, camera *Camera) {
	screenX, screenY := WorldToScreenIso(x, y, camera)
	rl.DrawCircle(int32(screenX), int32(screenY), radius, rl.Purple)
}

func GetDistance(x1, y1, x2, y2 float32) float32 {
	dx := x2 - x1
	dy := y2 - y1
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

func GetSpriteSheet() rl.Texture2D {
	return assets.Get().GetTexture(HumanoidSpriteSheetAssetKey)
}

func getSpriteSourceRect(row, col int) rl.Rectangle {
	x := float32(col) * spriteWidth
	y := float32(row) * spriteHeight
	return rl.NewRectangle(x, y, spriteWidth, spriteHeight)
}

func drawTextureOrRect(texture rl.Texture2D, sourceRect, destRect rl.Rectangle, tint rl.Color, fallbackColor rl.Color) {
	if texture.ID != 0 {
		rl.DrawTexturePro(texture, sourceRect, destRect, rl.NewVector2(0, 0), 0, tint)
		return
	}

	rl.DrawRectangleRec(destRect, fallbackColor)
	rl.DrawRectangleLinesEx(destRect, 1, rl.Black)
}

func drawHealthBar(destRect rl.Rectangle, healthPercent float32, height float32) {
	if healthPercent < 0 {
		healthPercent = 0
	}
	if healthPercent > 1 {
		healthPercent = 1
	}

	healthBarWidth := destRect.Width
	healthBarX := destRect.X
	healthBarY := destRect.Y - height - 3

	rl.DrawRectangleRec(rl.NewRectangle(healthBarX, healthBarY, healthBarWidth, height), rl.Red)
	rl.DrawRectangleRec(rl.NewRectangle(healthBarX, healthBarY, healthBarWidth*healthPercent, height), rl.Green)
}

func actorScreenRect(worldX, worldY, width, height float32, camera *Camera) (float32, float32) {
	anchorX := worldX + width/2
	anchorY := worldY + height
	screenX, screenY := WorldToScreenIso(anchorX, anchorY, camera)
	return screenX - width/2, screenY - height
}

func projectAABBBase(rect world.AABB, camera *Camera) [4]rl.Vector2 {
	x1, y1 := WorldToScreenIso(rect.X, rect.Y, camera)
	x2, y2 := WorldToScreenIso(rect.X+rect.Width, rect.Y, camera)
	x3, y3 := WorldToScreenIso(rect.X+rect.Width, rect.Y+rect.Height, camera)
	x4, y4 := WorldToScreenIso(rect.X, rect.Y+rect.Height, camera)

	return [4]rl.Vector2{
		rl.NewVector2(x1, y1),
		rl.NewVector2(x2, y2),
		rl.NewVector2(x3, y3),
		rl.NewVector2(x4, y4),
	}
}

func drawIsoQuad(points [4]rl.Vector2, color rl.Color) {
	rl.DrawTriangle(points[0], points[1], points[2], color)
	rl.DrawTriangle(points[0], points[2], points[3], color)
}

func drawIsoOutline(points [4]rl.Vector2, color rl.Color) {
	rl.DrawLineV(points[0], points[1], color)
	rl.DrawLineV(points[1], points[2], color)
	rl.DrawLineV(points[2], points[3], color)
	rl.DrawLineV(points[3], points[0], color)
}

func liftIso(point rl.Vector2, amount float32) rl.Vector2 {
	return rl.NewVector2(point.X, point.Y-amount)
}

func shadeColor(base rl.Color, delta int32) rl.Color {
	r := int32(base.R) + delta
	g := int32(base.G) + delta
	b := int32(base.B) + delta

	if r < 0 {
		r = 0
	}
	if g < 0 {
		g = 0
	}
	if b < 0 {
		b = 0
	}
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	return rl.NewColor(uint8(r), uint8(g), uint8(b), base.A)
}

func DrawSkillBar(player *gameobjects.Player, keyLabels []string) {
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

	if len(keyLabels) == 0 {
		keyLabels = []string{"Q", "W", "E", "R"}
	}

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

func DrawDebugOverlay(lines []string) {
	if len(lines) == 0 {
		return
	}

	padding := int32(10)
	lineHeight := int32(20)
	maxWidth := int32(0)
	for _, line := range lines {
		width := int32(rl.MeasureText(line, 18))
		if width > maxWidth {
			maxWidth = width
		}
	}

	panelWidth := maxWidth + padding*2
	panelHeight := int32(len(lines))*lineHeight + padding*2
	panelX := int32(10)
	panelY := int32(10)

	rl.DrawRectangle(panelX, panelY, panelWidth, panelHeight, rl.NewColor(0, 0, 0, 170))
	rl.DrawRectangleLines(panelX, panelY, panelWidth, panelHeight, rl.NewColor(220, 220, 220, 220))

	for i, line := range lines {
		x := panelX + padding
		y := panelY + padding + int32(i)*lineHeight
		rl.DrawText(line, x, y, 18, rl.RayWhite)
	}
}
