package game

import (
	"fmt"
	"math"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/systems"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (g *Game) drawRunHUD() {
	if g.Player == nil || g.Dungeon == nil {
		return
	}

	roomText := fmt.Sprintf("Room: %d/%d", g.Dungeon.CurrentRoom+1, len(g.Dungeon.Rooms))
	rl.DrawText(roomText, 10, 20, 20, rl.Black)

	g.drawBarWithText(10, 48, 280, 18, float32(g.Player.HP)/float32(g.Player.MaxHP), rl.NewColor(60, 20, 20, 255), rl.NewColor(220, 70, 70, 255), fmt.Sprintf("HP %d/%d", g.Player.HP, g.Player.MaxHP))
	if g.Player.MaxMana > 0 {
		g.drawBarWithText(10, 72, 280, 18, float32(g.Player.Mana)/float32(g.Player.MaxMana), rl.NewColor(22, 28, 60, 255), rl.NewColor(90, 150, 240, 255), fmt.Sprintf("Mana %d/%d", g.Player.Mana, g.Player.MaxMana))
	}

	xpRatio := float32(0)
	if g.Player.XPToNext > 0 {
		xpRatio = float32(g.Player.XP) / float32(g.Player.XPToNext)
	}
	g.drawBarWithText(10, 96, 280, 14, xpRatio, rl.NewColor(30, 30, 30, 255), rl.NewColor(215, 180, 80, 255), fmt.Sprintf("Lv %d  XP %d/%d", g.Player.Level, g.Player.XP, g.Player.XPToNext))

	statPointColor := rl.NewColor(90, 90, 90, 255)
	if g.Player.StatPoints > 0 {
		statPointColor = rl.NewColor(26, 132, 56, 255)
	}
	rl.DrawText(fmt.Sprintf("Stat Points: %d", g.Player.StatPoints), 10, 118, 20, statPointColor)

	g.drawMinimap()
	g.drawPlayerEffectsTray()
	g.drawTargetFrame()
}

func (g *Game) drawBarWithText(x, y, width, height, ratio float32, bgColor, fillColor rl.Color, text string) {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	outer := rl.NewRectangle(x, y, width, height)
	inner := rl.NewRectangle(x+1, y+1, (width-2)*ratio, height-2)
	rl.DrawRectangleRec(outer, bgColor)
	rl.DrawRectangleRec(inner, fillColor)
	rl.DrawRectangleLinesEx(outer, 1, rl.Black)
	rl.DrawText(text, int32(x+6), int32(y-1), 16, rl.RayWhite)
}

func (g *Game) drawClassSkillPreview(classType gamedata.ClassType) {
	skills := gamedata.GetClassSkillData(classType)
	if len(skills) == 0 {
		return
	}

	panelX := float32(WindowWidth/2 - 360)
	panelY := float32(WindowHeight/2 + 95)
	panelW := float32(720)
	panelH := float32(168)
	rl.DrawRectangleRec(rl.NewRectangle(panelX, panelY, panelW, panelH), rl.NewColor(0, 0, 0, 100))
	rl.DrawRectangleLinesEx(rl.NewRectangle(panelX, panelY, panelW, panelH), 1, rl.NewColor(60, 60, 60, 200))

	rl.DrawText("Skill Preview", int32(panelX+12), int32(panelY+8), 22, rl.Black)
	for i, skill := range skills {
		if skill == nil {
			continue
		}
		cellX := panelX + 12 + float32(i)*174
		cellY := panelY + 42
		iconRect := rl.NewRectangle(cellX, cellY, 40, 40)
		systems.DrawIconCell(systems.GetSkillIconCell(skill.Type), iconRect, rl.White, rl.NewColor(80, 80, 80, 255))
		rl.DrawRectangleLinesEx(iconRect, 1, rl.NewColor(30, 30, 30, 220))
		rl.DrawText(skill.Name, int32(cellX+48), int32(cellY+6), 18, rl.Black)
		rl.DrawText(fmt.Sprintf("CD %.0fs", math.Ceil(float64(skill.Cooldown))), int32(cellX+48), int32(cellY+24), 16, rl.DarkGray)
	}
}

func (g *Game) drawRewardItemIcon(item *gamedata.Item, x, y float32, selected bool) {
	if item == nil {
		return
	}

	bg := rl.NewColor(40, 40, 40, 255)
	if selected {
		bg = rl.NewColor(90, 70, 10, 255)
	}
	iconRect := rl.NewRectangle(x, y, 44, 44)
	rl.DrawRectangleRec(iconRect, bg)
	systems.DrawIconCell(systems.GetItemIconCell(item), rl.NewRectangle(x+4, y+4, 36, 36), rl.White, rl.NewColor(80, 80, 80, 255))
	rl.DrawRectangleLinesEx(iconRect, 1, rl.NewColor(220, 220, 220, 210))
}

func (g *Game) drawPlayerEffectsTray() {
	if g.Player == nil || len(g.Player.Effects) == 0 {
		return
	}

	startX := float32(10)
	startY := float32(WindowHeight - 160)
	slotSize := float32(36)
	spacing := float32(4)
	rl.DrawText("Effects", int32(startX), int32(startY-22), 18, rl.DarkGray)

	for i, effect := range g.Player.Effects {
		x := startX + float32(i)*(slotSize+spacing)
		rect := rl.NewRectangle(x, startY, slotSize, slotSize)
		rl.DrawRectangleRec(rect, rl.NewColor(20, 20, 20, 220))
		systems.DrawIconCell(systems.GetEffectIconCell(effect.Type), rl.NewRectangle(x+2, startY+2, slotSize-4, slotSize-4), rl.White, rl.NewColor(90, 90, 90, 255))
		rl.DrawRectangleLinesEx(rect, 1, effectBorderColor(effect.Type))
		rl.DrawText(fmt.Sprintf("%.1f", effect.TimeLeft), int32(x), int32(startY+slotSize+1), 14, rl.RayWhite)
	}
}

func effectBorderColor(effectType gamedata.EffectType) rl.Color {
	switch effectType {
	case gamedata.EffectSlow, gamedata.EffectStun, gamedata.EffectFreeze, gamedata.EffectSilence, gamedata.EffectBurn, gamedata.EffectPoison, gamedata.EffectMoveSpeedReduction:
		return rl.NewColor(225, 80, 80, 255)
	default:
		return rl.NewColor(80, 190, 100, 255)
	}
}

func (g *Game) drawTargetFrame() {
	target := g.getHoveredOrLockedTarget()
	if target == nil {
		return
	}

	name := "Enemy"
	isElite := false
	hp := 0
	maxHP := 0
	effects := []gamedata.EffectInstance{}

	switch t := target.(type) {
	case *gameobjects.Enemy:
		if t == nil || !t.IsAlive() {
			return
		}
		name = t.DisplayName()
		if t.IsElite {
			isElite = true
		}
		hp = t.HP
		maxHP = t.MaxHP
		effects = t.Effects
	case *gameobjects.Boss:
		if t == nil || !t.IsAlive() {
			return
		}
		name = "Dungeon Boss"
		hp = t.HP
		maxHP = t.MaxHP
		effects = t.Effects
	default:
		return
	}

	panelX := float32(WindowWidth - 290)
	panelY := float32(160)
	panelRect := rl.NewRectangle(panelX, panelY, 270, 118)
	rl.DrawRectangleRec(panelRect, rl.NewColor(0, 0, 0, 175))
	rl.DrawRectangleLinesEx(panelRect, 1, rl.NewColor(230, 230, 230, 220))

	titleColor := rl.RayWhite
	if isElite {
		titleColor = rl.NewColor(255, 210, 80, 255)
	}
	rl.DrawText(name, int32(panelX+10), int32(panelY+8), 22, titleColor)
	if maxHP > 0 {
		ratio := float32(hp) / float32(maxHP)
		g.drawBarWithText(panelX+10, panelY+40, 250, 16, ratio, rl.NewColor(60, 20, 20, 255), rl.NewColor(230, 70, 70, 255), fmt.Sprintf("%d/%d", hp, maxHP))
	}

	iconX := panelX + 10
	iconY := panelY + 68
	slotSize := float32(30)
	for i, effect := range effects {
		if i >= 8 {
			break
		}
		x := iconX + float32(i)*(slotSize+2)
		rect := rl.NewRectangle(x, iconY, slotSize, slotSize)
		rl.DrawRectangleRec(rect, rl.NewColor(25, 25, 25, 220))
		systems.DrawIconCell(systems.GetEffectIconCell(effect.Type), rl.NewRectangle(x+2, iconY+2, slotSize-4, slotSize-4), rl.White, rl.NewColor(80, 80, 80, 255))
		rl.DrawRectangleLinesEx(rect, 1, rl.NewColor(190, 190, 190, 220))
	}
}

func (g *Game) getHoveredOrLockedTarget() interface{} {
	if g.PlayerAttackTarget != nil {
		switch t := g.PlayerAttackTarget.(type) {
		case *gameobjects.Enemy:
			if t != nil && t.IsAlive() {
				return t
			}
		case *gameobjects.Boss:
			if t != nil && t.IsAlive() {
				return t
			}
		}
	}

	if g.Camera == nil {
		return nil
	}

	mouseX, mouseY := systems.GetMousePosition()
	worldX, worldY := systems.ScreenToWorldIso(mouseX, mouseY, g.Camera)

	if g.Boss != nil && g.Boss.IsAlive() && pointInRect(worldX, worldY, g.Boss.PosX, g.Boss.PosY, g.Boss.Hitbox.Width, g.Boss.Hitbox.Height) {
		return g.Boss
	}
	for _, enemy := range g.Enemies {
		if enemy == nil || !enemy.IsAlive() {
			continue
		}
		if pointInRect(worldX, worldY, enemy.PosX, enemy.PosY, enemy.Hitbox.Width, enemy.Hitbox.Height) {
			return enemy
		}
	}

	return nil
}

func pointInRect(x, y, minX, minY, width, height float32) bool {
	return x >= minX && x <= minX+width && y >= minY && y <= minY+height
}

func (g *Game) drawMinimap() {
	if g.Dungeon == nil || len(g.Dungeon.Rooms) == 0 {
		return
	}

	panelX := float32(WindowWidth - 220)
	panelY := float32(20)
	panelW := float32(200)
	panelH := float32(130)
	panelRect := rl.NewRectangle(panelX, panelY, panelW, panelH)
	rl.DrawRectangleRec(panelRect, rl.NewColor(0, 0, 0, 145))
	rl.DrawRectangleLinesEx(panelRect, 1, rl.NewColor(220, 220, 220, 200))
	rl.DrawText("Minimap", int32(panelX+10), int32(panelY+6), 18, rl.RayWhite)

	minX := g.Dungeon.Rooms[0].X
	minY := g.Dungeon.Rooms[0].Y
	maxX := g.Dungeon.Rooms[0].X + g.Dungeon.Rooms[0].Width
	maxY := g.Dungeon.Rooms[0].Y + g.Dungeon.Rooms[0].Height
	for _, room := range g.Dungeon.Rooms {
		if room.X < minX {
			minX = room.X
		}
		if room.Y < minY {
			minY = room.Y
		}
		if room.X+room.Width > maxX {
			maxX = room.X + room.Width
		}
		if room.Y+room.Height > maxY {
			maxY = room.Y + room.Height
		}
	}

	spanX := maxX - minX
	spanY := maxY - minY
	if spanX <= 0 {
		spanX = 1
	}
	if spanY <= 0 {
		spanY = 1
	}

	mapX := panelX + 10
	mapY := panelY + 32
	mapW := panelW - 20
	mapH := panelH - 42

	for i, room := range g.Dungeon.Rooms {
		rx := mapX + ((room.X-minX)/spanX)*mapW
		ry := mapY + ((room.Y-minY)/spanY)*mapH
		rw := (room.Width / spanX) * mapW
		rh := (room.Height / spanY) * mapH
		if rw < 8 {
			rw = 8
		}
		if rh < 6 {
			rh = 6
		}

		color := rl.NewColor(95, 95, 95, 255)
		if room.Completed {
			color = rl.NewColor(70, 170, 85, 255)
		}
		if room.IsBoss() {
			color = rl.NewColor(140, 90, 190, 255)
		}
		if i == g.Dungeon.CurrentRoom {
			color = rl.NewColor(250, 225, 90, 255)
		}
		rl.DrawRectangleRec(rl.NewRectangle(rx, ry, rw, rh), color)
		rl.DrawRectangleLinesEx(rl.NewRectangle(rx, ry, rw, rh), 1, rl.NewColor(15, 15, 15, 220))
	}
}
