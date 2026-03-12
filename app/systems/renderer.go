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
	FloorAtlasAssetKey          = "texture.atlas.floor"
	WallsHighAtlasAssetKey      = "texture.atlas.walls.high"
)

const (
	floorAtlasTileWidth    = 16
	floorAtlasTileHeight   = 16
	wallsAtlasTileWidth    = 16
	wallsAtlasTileHeight   = 32
	doorAtlasTileWidth     = 32
	doorAtlasTileHeight    = 64
	facingWallTileX        = 0
	facingWallTileY        = 3
	perpendicularWallTileX = 0
	perpendicularWallTileY = 2
	doorTileX              = 16
	doorTileY              = 2
	wallHeightTiles        = 2.0
	doorHeightTiles        = 2.0
)

var floorTileVariants = [][2]int{
	{0, 0},
	{0, 1},
	{0, 2},
	{1, 0},
	{1, 1},
	{1, 2},
}

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
	if !player.FacingRight {
		sourceRect.X += sourceRect.Width
		sourceRect.Width = -sourceRect.Width
	}

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

	sourceRect := enemySpriteSourceRect(enemy)
	destRect := rl.NewRectangle(screenX, screenY, enemy.Hitbox.Width, enemy.Hitbox.Height)

	tint := enemyArchetypeTint(enemy)
	if enemy.HitFlashTimer > 0 {
		tint = rl.Orange
	} else if enemy.AttackFlashTimer > 0 {
		tint = rl.Yellow
	} else if enemy.IsElite {
		tint = blendColor(tint, rl.NewColor(255, 210, 96, 255), 0.55)
	}

	drawTextureOrRect(GetSpriteSheet(), sourceRect, destRect, tint, rl.Red)
	drawHealthBar(destRect, float32(enemy.HP)/float32(enemy.MaxHP), 5)
}

func DrawRoom(room *world.Room, camera *Camera) {
	if room == nil {
		return
	}

	if hasTemplateTileGrid(room) {
		drawTemplateRoom(room, camera)
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
	screenX, screenY := WorldToScreenIso(obstacle.X, obstacle.Y, camera)
	rl.DrawRectangleRec(
		rl.NewRectangle(screenX, screenY, obstacle.Width, obstacle.Height),
		rl.NewColor(92, 92, 92, 255),
	)
	rl.DrawRectangleLinesEx(
		rl.NewRectangle(screenX, screenY, obstacle.Width, obstacle.Height),
		1,
		rl.NewColor(35, 35, 35, 255),
	)
}

func DrawDoor(door *world.Door, camera *Camera) {
	if door == nil {
		return
	}

	wallsAtlas := assets.Get().GetTexture(WallsHighAtlasAssetKey)
	if wallsAtlas.ID != 0 {
		drawDoorAtlasSprite(door, camera, wallsAtlas)
		return
	}

	screenX, screenY := WorldToScreenIso(door.Bounds.X, door.Bounds.Y, camera)
	color := rl.NewColor(48, 168, 102, 255)
	if door.Locked {
		color = rl.NewColor(168, 82, 68, 255)
	}
	rl.DrawRectangleRec(rl.NewRectangle(screenX, screenY, door.Bounds.Width, door.Bounds.Height), color)
	rl.DrawRectangleLinesEx(rl.NewRectangle(screenX, screenY, door.Bounds.Width, door.Bounds.Height), 1, rl.Black)
}

func hasTemplateTileGrid(room *world.Room) bool {
	return room != nil && len(room.Tiles) > 0 && len(room.Tiles[0]) > 0
}

func drawTemplateRoom(room *world.Room, camera *Camera) {
	floorAtlas := assets.Get().GetTexture(FloorAtlasAssetKey)
	wallsAtlas := assets.Get().GetTexture(WallsHighAtlasAssetKey)
	tileSize := float32(world.RoomTemplateTileSize)

	// Draw ground first so walls and doors can overlay it.
	for y, row := range room.Tiles {
		for x, tile := range row {
			worldX := room.X + float32(x)*world.RoomTemplateTileSize
			worldY := room.Y + float32(y)*world.RoomTemplateTileSize

			switch tile {
			case world.TileFloor, world.TileDoor, world.TileHazard, world.TileTrap, world.TileWall:
				if floorAtlas.ID != 0 {
					floorTileX, floorTileY := floorVariantForCell(room, x, y)
					drawAtlasTileAtCell(
						floorAtlas,
						floorTileX,
						floorTileY,
						floorAtlasTileWidth,
						floorAtlasTileHeight,
						worldX,
						worldY,
						tileSize,
						tileSize,
						camera,
						rl.White,
					)
				} else {
					screenX, screenY := WorldToScreenIso(worldX, worldY, camera)
					rl.DrawRectangleRec(rl.NewRectangle(screenX, screenY, tileSize, tileSize), TerrainColorNormalRGBA)
				}
			default:
			}
		}
	}

	// Draw wall facades in a separate pass to maintain clean layering.
	for y, row := range room.Tiles {
		for x, tile := range row {
			if tile != world.TileWall {
				continue
			}

			worldX := room.X + float32(x)*world.RoomTemplateTileSize
			worldY := room.Y + float32(y)*world.RoomTemplateTileSize

			if wallsAtlas.ID != 0 {
				wallTileX, wallTileY, ok := wallAtlasTileForPosition(room, x, y)
				if !ok {
					continue
				}
				drawAtlasTileBottomAnchored(
					wallsAtlas,
					wallTileX,
					wallTileY,
					wallsAtlasTileWidth,
					wallsAtlasTileHeight,
					worldX,
					worldY,
					tileSize,
					tileSize*wallHeightTiles,
					camera,
					rl.White,
				)
				continue
			}

			DrawObstacle(world.AABB{
				X:      worldX,
				Y:      worldY,
				Width:  tileSize,
				Height: tileSize,
			}, camera)
		}
	}

	for _, door := range room.Doors {
		DrawDoor(door, camera)
	}
}

func wallAtlasTileForPosition(room *world.Room, x, y int) (int, int, bool) {
	hasWalkableAbove := tileIsWalkable(room, x, y-1)
	hasWalkableBelow := tileIsWalkable(room, x, y+1)
	hasWalkableLeft := tileIsWalkable(room, x-1, y)
	hasWalkableRight := tileIsWalkable(room, x+1, y)

	if hasWalkableAbove || hasWalkableBelow {
		return facingWallTileX, facingWallTileY, true
	}
	if hasWalkableLeft || hasWalkableRight {
		return perpendicularWallTileX, perpendicularWallTileY, true
	}
	return 0, 0, false
}

func tileIsWalkable(room *world.Room, x, y int) bool {
	if room == nil || y < 0 || y >= len(room.Tiles) || x < 0 || x >= len(room.Tiles[y]) {
		return false
	}
	tile := room.Tiles[y][x]
	return tile == world.TileFloor || tile == world.TileDoor || tile == world.TileHazard || tile == world.TileTrap
}

func floorVariantForCell(room *world.Room, x, y int) (int, int) {
	hash := (room.ProgressionIndex+1)*73856093 + (x+1)*19349663 + (y+1)*83492791
	if hash < 0 {
		hash = -hash
	}
	index := hash % len(floorTileVariants)
	return floorTileVariants[index][0], floorTileVariants[index][1]
}

func drawAtlasTileAtCell(texture rl.Texture2D, atlasX, atlasY, srcTileWidth, srcTileHeight int, worldX, worldY, destWidth, destHeight float32, camera *Camera, tint rl.Color) {
	source := rl.NewRectangle(
		float32(atlasX*srcTileWidth),
		float32(atlasY*srcTileHeight),
		float32(srcTileWidth),
		float32(srcTileHeight),
	)

	screenX, screenY := WorldToScreenIso(worldX, worldY, camera)
	dest := rl.NewRectangle(screenX, screenY, destWidth, destHeight)
	rl.DrawTexturePro(texture, source, dest, rl.NewVector2(0, 0), 0, tint)
}

func drawAtlasTileBottomAnchored(texture rl.Texture2D, atlasX, atlasY, srcTileWidth, srcTileHeight int, worldX, worldY, destWidth, destHeight float32, camera *Camera, tint rl.Color) {
	source := rl.NewRectangle(
		float32(atlasX*srcTileWidth),
		float32(atlasY*srcTileHeight),
		float32(srcTileWidth),
		float32(srcTileHeight),
	)

	screenX, screenY := WorldToScreenIso(worldX, worldY, camera)
	dest := rl.NewRectangle(screenX, screenY+destWidth-destHeight, destWidth, destHeight)
	rl.DrawTexturePro(texture, source, dest, rl.NewVector2(0, 0), 0, tint)
}

func drawDoorAtlasSprite(door *world.Door, camera *Camera, texture rl.Texture2D) {
	if door == nil {
		return
	}
	tint := rl.White
	if door.Locked {
		tint = rl.NewColor(230, 210, 210, 255)
	}

	drawAtlasTileBottomAnchored(
		texture,
		doorTileX,
		doorTileY,
		doorAtlasTileWidth,
		doorAtlasTileHeight,
		door.Bounds.X,
		door.Bounds.Y,
		door.Bounds.Width,
		door.Bounds.Height*doorHeightTiles,
		camera,
		tint,
	)
}

func DrawProjectile(x, y, radius float32, camera *Camera) {
	screenX, screenY := WorldToScreenIso(x, y, camera)
	rl.DrawCircle(int32(screenX), int32(screenY), radius, ProjectileColorRGBA)
}

func DrawSkillProjectile(x, y, radius float32, skill *gamedata.Skill, camera *Camera) {
	screenX, screenY := WorldToScreenIso(x, y, camera)
	visualColor, visualRadius := skillVisualStyle(skill)
	if radius <= 0 {
		radius = visualRadius
	}
	rl.DrawCircle(int32(screenX), int32(screenY), radius, visualColor)
}

func DrawDelayedTelegraph(x, y, radius float32, skill *gamedata.Skill, camera *Camera) {
	screenX, screenY := WorldToScreenIso(x, y, camera)
	visualColor, _ := skillVisualStyle(skill)
	telegraphFill := rl.NewColor(visualColor.R, visualColor.G, visualColor.B, 40)
	telegraphOutline := rl.NewColor(visualColor.R, visualColor.G, visualColor.B, 220)
	rl.DrawCircle(int32(screenX), int32(screenY), radius, telegraphFill)
	rl.DrawCircleLines(int32(screenX), int32(screenY), radius, telegraphOutline)
}

func DrawActiveSkillZone(x, y, radius float32, skill *gamedata.Skill, camera *Camera) {
	screenX, screenY := WorldToScreenIso(x, y, camera)
	visualColor, _ := skillVisualStyle(skill)
	zoneFill := rl.NewColor(visualColor.R, visualColor.G, visualColor.B, 70)
	zoneOutline := rl.NewColor(visualColor.R, visualColor.G, visualColor.B, 255)
	rl.DrawCircle(int32(screenX), int32(screenY), radius, zoneFill)
	rl.DrawCircleLines(int32(screenX), int32(screenY), radius, zoneOutline)
}

func DrawSkillCastPulse(x, y, radius, remainingRatio float32, skill *gamedata.Skill, filled bool, camera *Camera) {
	if remainingRatio < 0 {
		remainingRatio = 0
	}
	if remainingRatio > 1 {
		remainingRatio = 1
	}
	screenX, screenY := WorldToScreenIso(x, y, camera)
	visualColor, _ := skillVisualStyle(skill)
	pulseRadius := radius * (1 + (1-remainingRatio)*0.25)
	if filled {
		fill := rl.NewColor(visualColor.R, visualColor.G, visualColor.B, 60)
		rl.DrawCircle(int32(screenX), int32(screenY), pulseRadius, fill)
	}
	outlineAlpha := uint8(120 + 135*remainingRatio)
	outline := rl.NewColor(visualColor.R, visualColor.G, visualColor.B, outlineAlpha)
	rl.DrawCircleLines(int32(screenX), int32(screenY), pulseRadius, outline)
}

func DrawBoss(boss *gameobjects.Boss, camera *Camera) {
	if !boss.IsAlive() {
		return
	}

	for _, zone := range boss.ActiveAreaZones() {
		screenX, screenY := WorldToScreenIso(zone.X, zone.Y, camera)
		if zone.Active {
			fill := rl.NewColor(208, 54, 54, 88)
			outline := rl.NewColor(255, 96, 96, 230)
			rl.DrawCircle(int32(screenX), int32(screenY), zone.Radius, fill)
			rl.DrawCircleLines(int32(screenX), int32(screenY), zone.Radius, outline)
			continue
		}

		warningRatio := float32(1)
		if zone.WarningDuration > 0 {
			warningRatio = zone.WarningTimeLeft / zone.WarningDuration
		}
		if warningRatio < 0 {
			warningRatio = 0
		}
		if warningRatio > 1 {
			warningRatio = 1
		}
		alpha := uint8(40 + (1-warningRatio)*90)
		fill := rl.NewColor(255, 192, 84, alpha)
		outline := rl.NewColor(255, 226, 148, 220)
		rl.DrawCircle(int32(screenX), int32(screenY), zone.Radius, fill)
		rl.DrawCircleLines(int32(screenX), int32(screenY), zone.Radius, outline)
	}

	if telegraph, ok := boss.ActiveHeavyTelegraph(); ok {
		screenX, screenY := WorldToScreenIso(telegraph.X, telegraph.Y, camera)
		rl.DrawCircleLines(int32(screenX), int32(screenY), telegraph.Radius, rl.NewColor(255, 88, 88, 255))
		rl.DrawCircleLines(int32(screenX), int32(screenY), telegraph.Radius+3, rl.NewColor(255, 166, 120, 220))
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

	drawTextureOrRect(GetSpriteSheet(), sourceRect, destRect, tint, rl.Purple)
	drawHealthBar(destRect, float32(boss.HP)/float32(boss.MaxHP), 8)
}

func DrawBossProjectile(x, y, radius float32, camera *Camera) {
	screenX, screenY := WorldToScreenIso(x, y, camera)
	rl.DrawCircle(int32(screenX), int32(screenY), radius, rl.Purple)
}

func DrawEnemyProjectile(x, y, radius float32, camera *Camera) {
	screenX, screenY := WorldToScreenIso(x, y, camera)
	rl.DrawCircle(int32(screenX), int32(screenY), radius, rl.NewColor(255, 140, 60, 255))
}

func enemySpriteSourceRect(enemy *gameobjects.Enemy) rl.Rectangle {
	if enemy == nil {
		return getSpriteSourceRect(2, 4)
	}

	switch enemy.Archetype {
	case gamedata.EnemyArchetypeRaider:
		return getSpriteSourceRect(2, 3)
	case gamedata.EnemyArchetypePikeman:
		return getSpriteSourceRect(2, 2)
	case gamedata.EnemyArchetypeArcher:
		return getSpriteSourceRect(2, 1)
	case gamedata.EnemyArchetypeHexCaller:
		return getSpriteSourceRect(1, 2)
	case gamedata.EnemyArchetypeBrute:
		return getSpriteSourceRect(2, 0)
	case gamedata.EnemyArchetypeSwarmling:
		return getSpriteSourceRect(1, 1)
	default:
		return getSpriteSourceRect(2, 3)
	}
}

func enemyArchetypeTint(enemy *gameobjects.Enemy) rl.Color {
	if enemy == nil {
		return rl.White
	}

	switch enemy.Archetype {
	case gamedata.EnemyArchetypeRaider:
		return rl.NewColor(235, 132, 112, 255)
	case gamedata.EnemyArchetypePikeman:
		return rl.NewColor(210, 86, 86, 255)
	case gamedata.EnemyArchetypeArcher:
		return rl.NewColor(128, 210, 124, 255)
	case gamedata.EnemyArchetypeHexCaller:
		return rl.NewColor(122, 152, 246, 255)
	case gamedata.EnemyArchetypeBrute:
		return rl.NewColor(182, 132, 96, 255)
	case gamedata.EnemyArchetypeSwarmling:
		return rl.NewColor(248, 220, 118, 255)
	default:
		return rl.White
	}
}

func blendColor(base, overlay rl.Color, strength float32) rl.Color {
	if strength < 0 {
		strength = 0
	}
	if strength > 1 {
		strength = 1
	}
	inv := 1 - strength
	return rl.NewColor(
		uint8(float32(base.R)*inv+float32(overlay.R)*strength),
		uint8(float32(base.G)*inv+float32(overlay.G)*strength),
		uint8(float32(base.B)*inv+float32(overlay.B)*strength),
		base.A,
	)
}

func skillVisualStyle(skill *gamedata.Skill) (rl.Color, float32) {
	if skill == nil {
		return ProjectileColorRGBA, 5
	}

	switch skill.Type {
	case gamedata.SkillTypePowerStrike:
		return rl.NewColor(232, 112, 44, 255), 9
	case gamedata.SkillTypeGuardStance:
		return rl.NewColor(80, 140, 210, 255), 10
	case gamedata.SkillTypeBloodOath:
		return rl.NewColor(180, 36, 36, 255), 9
	case gamedata.SkillTypeShockwaveSlam:
		return rl.NewColor(255, 165, 0, 255), 11
	case gamedata.SkillTypeQuickShot:
		return rl.NewColor(255, 232, 90, 255), 6
	case gamedata.SkillTypeRetreatRoll:
		return rl.NewColor(132, 212, 168, 255), 9
	case gamedata.SkillTypeFocusedAim:
		return rl.NewColor(255, 190, 56, 255), 9
	case gamedata.SkillTypePoisonTip:
		return rl.NewColor(92, 220, 96, 255), 7
	case gamedata.SkillTypeArcaneBolt:
		return rl.NewColor(86, 156, 255, 255), 8
	case gamedata.SkillTypeManaShield:
		return rl.NewColor(96, 220, 255, 255), 10
	case gamedata.SkillTypeFrostField:
		return rl.NewColor(120, 180, 255, 255), 10
	case gamedata.SkillTypeArcaneDrain:
		return rl.NewColor(180, 106, 255, 255), 10
	default:
		return ProjectileColorRGBA, 5
	}
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
		if slot.Skill != nil {
			DrawIconCell(GetSkillIconCell(slot.Skill.Type), iconRect, rl.White, rl.NewColor(80, 80, 80, 255))
		} else {
			rl.DrawRectangleRec(iconRect, rl.NewColor(80, 80, 80, 255))
		}

		if slot.KeyLabel != "" {
			textWidth := rl.MeasureText(slot.KeyLabel, 20)
			textX := int32(slotX + 5)
			textY := int32(slotY + slotHeight - 22)
			rl.DrawRectangle(textX-2, textY-2, int32(textWidth)+4, 24, rl.NewColor(0, 0, 0, 180))
			rl.DrawText(slot.KeyLabel, textX, textY, 20, rl.RayWhite)
		}

		if slot.Skill != nil {
			if slot.Skill.ManaCost > 0 {
				costText := fmt.Sprintf("%d", slot.Skill.ManaCost)
				costWidth := rl.MeasureText(costText, 16)
				costX := int32(slotX + slotWidth - float32(costWidth) - 8)
				costY := int32(slotY + 4)
				costColor := rl.NewColor(120, 210, 255, 255)
				if !player.CanUseMana(slot.Skill.ManaCost) {
					costColor = rl.NewColor(255, 90, 90, 255)
				}
				rl.DrawRectangle(costX-2, costY-1, int32(costWidth)+4, 18, rl.NewColor(0, 0, 0, 200))
				rl.DrawText(costText, costX, costY, 16, costColor)
			}

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
