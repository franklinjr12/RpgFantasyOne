package game

import (
	"fmt"
	"math"
	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/systems"
	"singlefantasy/app/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type RunState int

const (
	RunStateMenu RunState = iota
	RunStateInGame
	RunStateVictory
	RunStateDefeat
	RunStateRewardSelection
)

type Game struct {
	State          RunState
	Player         *gameobjects.Player
	Enemies        []*gameobjects.Enemy
	Boss           *gameobjects.Boss
	Dungeon        *world.Dungeon
	Camera         *systems.Camera
	Projectiles    []*Projectile
	CurrentRoom    *world.Room
	SelectedClass  gamedata.ClassType
	LevelUpMenu    bool
	RewardOptions  []*gamedata.Item
	SelectedReward int
}

type Projectile struct {
	X      float32
	Y      float32
	VX     float32
	VY     float32
	Speed  float32
	Damage int
	Radius float32
	Alive  bool
}

func NewGame() *Game {
	return &Game{
		State:          RunStateMenu,
		Player:         nil,
		Enemies:        []*gameobjects.Enemy{},
		Boss:           nil,
		Dungeon:        nil,
		Camera:         systems.NewCamera(),
		Projectiles:    []*Projectile{},
		CurrentRoom:    nil,
		SelectedClass:  gamedata.ClassTypeMelee,
		LevelUpMenu:    false,
		RewardOptions:  []*gamedata.Item{},
		SelectedReward: 0,
	}
}

func (g *Game) StartRun() {
	g.ResetState()
	g.Dungeon = world.NewDungeon()
	g.CurrentRoom = g.Dungeon.GetCurrentRoom()

	startX := g.CurrentRoom.X + g.CurrentRoom.Width/2
	startY := g.CurrentRoom.Y + g.CurrentRoom.Height/2
	g.Player = gameobjects.NewPlayer(startX, startY, g.SelectedClass)

	g.SpawnRoomEnemies()
	g.State = RunStateInGame
}

func (g *Game) EndRun(victory bool) {
	if victory {
		g.RewardOptions = gamedata.GenerateRewardOptions(g.Player.Class.Type)
		g.SelectedReward = 0
		g.State = RunStateRewardSelection
	} else {
		g.State = RunStateDefeat
	}
}

func (g *Game) ResetState() {
	g.Player = nil
	g.Enemies = []*gameobjects.Enemy{}
	g.Boss = nil
	g.Dungeon = nil
	g.Projectiles = []*Projectile{}
	g.CurrentRoom = nil
	g.Camera = systems.NewCamera()
}

func (g *Game) SpawnRoomEnemies() {
	if g.CurrentRoom == nil {
		return
	}

	g.Enemies = []*gameobjects.Enemy{}
	g.Boss = nil

	if g.CurrentRoom.IsBoss() {
		bossX := g.CurrentRoom.X + g.CurrentRoom.Width/2
		bossY := g.CurrentRoom.Y + g.CurrentRoom.Height/2
		g.Boss = gameobjects.NewBoss(bossX, bossY)
	} else {
		for _, enemyRef := range g.CurrentRoom.Enemies {
			enemy := gameobjects.NewEnemy(enemyRef.X, enemyRef.Y, enemyRef.IsElite)
			g.Enemies = append(g.Enemies, enemy)
		}
	}
}

func (g *Game) CheckRoomCompletion() bool {
	if g.CurrentRoom == nil {
		return false
	}

	if g.CurrentRoom.IsBoss() {
		if g.Boss != nil && !g.Boss.Alive {
			g.CurrentRoom.Completed = true
			return true
		}
		return false
	}

	for _, enemy := range g.Enemies {
		if enemy.Alive {
			return false
		}
	}

	g.CurrentRoom.Completed = true
	return true
}

func (g *Game) AdvanceToNextRoom() {
	if g.Dungeon == nil {
		return
	}

	g.Dungeon.CurrentRoom++
	if g.Dungeon.CurrentRoom >= len(g.Dungeon.Rooms) {
		g.EndRun(true)
		return
	}

	g.CurrentRoom = g.Dungeon.GetCurrentRoom()
	startX := g.CurrentRoom.X + g.CurrentRoom.Width/2
	startY := g.CurrentRoom.Y + g.CurrentRoom.Height/2
	g.Player.X = startX
	g.Player.Y = startY
	g.Player.Health = g.Player.MaxHealth

	g.SpawnRoomEnemies()
	g.Projectiles = []*Projectile{}
}

func (g *Game) Update(deltaTime float32) {
	if g.State != RunStateInGame {
		return
	}

	if g.Player == nil {
		return
	}

	input := systems.UpdateInput()

	if input.Skill1 && len(g.Player.Skills) > 0 && g.Player.Skills[0].CanUse() {
		skill := g.Player.Skills[0]
		switch skill.Type {
		case gamedata.SkillTypeDash:
			mouseX, mouseY := systems.GetMousePosition()
			worldX, worldY := systems.ScreenToWorld(mouseX, mouseY, g.Camera)
			playerCenterX := g.Player.X + g.Player.Width/2
			playerCenterY := g.Player.Y + g.Player.Height/2
			dx := worldX - playerCenterX
			dy := worldY - playerCenterY
			distance := systems.GetDistance(0, 0, dx, dy)
			if distance > 0 {
				dashDistance := float32(100)
				g.Player.X += (dx / distance) * dashDistance
				g.Player.Y += (dy / distance) * dashDistance
			}
			skill.Use()
		case gamedata.SkillTypeMultiShot:
			if g.Player.Class.Type == gamedata.ClassTypeRanged {
				mouseX, mouseY := systems.GetMousePosition()
				worldX, worldY := systems.ScreenToWorld(mouseX, mouseY, g.Camera)
				playerCenterX := g.Player.X + g.Player.Width/2
				playerCenterY := g.Player.Y + g.Player.Height/2
				dx := worldX - playerCenterX
				dy := worldY - playerCenterY
				distance := systems.GetDistance(0, 0, dx, dy)
				if distance > 0 {
					speed := float32(400)
					angles := []float32{-0.3, 0, 0.3}
					for _, angle := range angles {
						cos := float32(1.0)
						sin := angle
						projDx := (dx*cos - dy*sin) / distance
						projDy := (dx*sin + dy*cos) / distance
						proj := &Projectile{
							X:      playerCenterX,
							Y:      playerCenterY,
							VX:     projDx * speed,
							VY:     projDy * speed,
							Speed:  speed,
							Damage: g.Player.AttackDamage,
							Radius: 5,
							Alive:  true,
						}
						g.Projectiles = append(g.Projectiles, proj)
					}
				}
			}
			skill.Use()
		case gamedata.SkillTypeManaShield:
			if g.Player.CanUseMana(skill.ManaCost) {
				g.Player.UseMana(skill.ManaCost)
				g.Player.ManaShieldActive = true
				g.Player.ManaShieldAmount = g.Player.Mana / 2
				skill.Use()
			}
		}
	}

	if input.MoveUp {
		g.Player.Y -= g.Player.MoveSpeed * deltaTime
	}
	if input.MoveDown {
		g.Player.Y += g.Player.MoveSpeed * deltaTime
	}
	if input.MoveLeft {
		g.Player.X -= g.Player.MoveSpeed * deltaTime
	}
	if input.MoveRight {
		g.Player.X += g.Player.MoveSpeed * deltaTime
	}

	if g.CurrentRoom != nil {
		if g.Player.X < g.CurrentRoom.X {
			g.Player.X = g.CurrentRoom.X
		}
		if g.Player.X+g.Player.Width > g.CurrentRoom.X+g.CurrentRoom.Width {
			g.Player.X = g.CurrentRoom.X + g.CurrentRoom.Width - g.Player.Width
		}
		if g.Player.Y < g.CurrentRoom.Y {
			g.Player.Y = g.CurrentRoom.Y
		}
		if g.Player.Y+g.Player.Height > g.CurrentRoom.Y+g.CurrentRoom.Height {
			g.Player.Y = g.CurrentRoom.Y + g.CurrentRoom.Height - g.Player.Height
		}
	}

	if input.Attack {
		mouseX, mouseY := systems.GetMousePosition()
		worldX, worldY := systems.ScreenToWorld(mouseX, mouseY, g.Camera)

		switch g.Player.Class.Type {
		case gamedata.ClassTypeMelee:
			for _, enemy := range g.Enemies {
				if !enemy.Alive {
					continue
				}

				if worldX >= enemy.X && worldX <= enemy.X+enemy.Width &&
					worldY >= enemy.Y && worldY <= enemy.Y+enemy.Height {

					playerCenterX := g.Player.X + g.Player.Width/2
					playerCenterY := g.Player.Y + g.Player.Height/2
					enemyCenterX := enemy.X + enemy.Width/2
					enemyCenterY := enemy.Y + enemy.Height/2

					distance := systems.GetDistance(playerCenterX, playerCenterY, enemyCenterX, enemyCenterY)

					if distance <= g.Player.AttackRange {
						wasAlive := enemy.Alive
						damage := g.Player.AttackDamage
						enemy.TakeDamage(damage)
						lifesteal := int(float32(damage) * g.Player.Class.LifestealPercent)
						g.Player.Heal(lifesteal)

						if wasAlive && !enemy.Alive {
							g.Player.GainXP(20)
						}
						break
					}
				}
			}
		case gamedata.ClassTypeRanged:
			playerCenterX := g.Player.X + g.Player.Width/2
			playerCenterY := g.Player.Y + g.Player.Height/2
			dx := worldX - playerCenterX
			dy := worldY - playerCenterY
			distance := systems.GetDistance(0, 0, dx, dy)
			if distance > 0 {
				speed := float32(400)
				proj := &Projectile{
					X:      playerCenterX,
					Y:      playerCenterY,
					VX:     (dx / distance) * speed,
					VY:     (dy / distance) * speed,
					Speed:  speed,
					Damage: g.Player.AttackDamage,
					Radius: 5,
					Alive:  true,
				}
				g.Projectiles = append(g.Projectiles, proj)
			}
		case gamedata.ClassTypeCaster:
			if g.Player.CanUseMana(g.Player.Class.ManaCost) {
				for _, enemy := range g.Enemies {
					if !enemy.Alive {
						continue
					}

					if worldX >= enemy.X && worldX <= enemy.X+enemy.Width &&
						worldY >= enemy.Y && worldY <= enemy.Y+enemy.Height {

						playerCenterX := g.Player.X + g.Player.Width/2
						playerCenterY := g.Player.Y + g.Player.Height/2
						enemyCenterX := enemy.X + enemy.Width/2
						enemyCenterY := enemy.Y + enemy.Height/2

						distance := systems.GetDistance(playerCenterX, playerCenterY, enemyCenterX, enemyCenterY)

						if distance <= g.Player.AttackRange {
							wasAlive := enemy.Alive
							enemy.TakeDamage(g.Player.AttackDamage)
							g.Player.UseMana(g.Player.Class.ManaCost)

							if wasAlive && !enemy.Alive {
								g.Player.GainXP(20)
							}
						}
						break
					}
				}
			}
		}
	}

	g.Player.Update(deltaTime)

	for _, enemy := range g.Enemies {
		enemy.Update(deltaTime, g.Player.X+g.Player.Width/2, g.Player.Y+g.Player.Height/2)
		enemy.Attack(g.Player)
	}

	if g.Boss != nil {
		g.Boss.Update(deltaTime, g.Player.X+g.Player.Width/2, g.Player.Y+g.Player.Height/2)
		g.Boss.Attack(g.Player)

		if g.Boss.ShouldSpawnAdds() {
			for i := 0; i < 2; i++ {
				angle := float32(i) * 3.14159 * 2.0 / 2.0
				addX := g.Boss.X + g.Boss.Width/2 + float32(math.Cos(float64(angle)))*100
				addY := g.Boss.Y + g.Boss.Height/2 + float32(math.Sin(float64(angle)))*100
				add := gameobjects.NewEnemy(addX, addY, false)
				g.Enemies = append(g.Enemies, add)
			}
			g.Boss.ResetAddSpawnTimer()
		}

		for i := len(g.Boss.Projectiles) - 1; i >= 0; i-- {
			proj := g.Boss.Projectiles[i]
			if !proj.Alive {
				g.Boss.Projectiles = append(g.Boss.Projectiles[:i], g.Boss.Projectiles[i+1:]...)
				continue
			}

			proj.X += proj.VX * deltaTime
			proj.Y += proj.VY * deltaTime

			playerCenterX := g.Player.X + g.Player.Width/2
			playerCenterY := g.Player.Y + g.Player.Height/2
			distance := systems.GetDistance(proj.X, proj.Y, playerCenterX, playerCenterY)

			if distance <= proj.Radius+g.Player.Width/2 {
				g.Player.TakeDamage(proj.Damage)
				proj.Alive = false
			}

			if g.CurrentRoom != nil {
				if proj.X < g.CurrentRoom.X || proj.X > g.CurrentRoom.X+g.CurrentRoom.Width ||
					proj.Y < g.CurrentRoom.Y || proj.Y > g.CurrentRoom.Y+g.CurrentRoom.Height {
					proj.Alive = false
				}
			}
		}

		for _, proj := range g.Projectiles {
			if !proj.Alive {
				continue
			}

			bossCenterX := g.Boss.X + g.Boss.Width/2
			bossCenterY := g.Boss.Y + g.Boss.Height/2
			distance := systems.GetDistance(proj.X, proj.Y, bossCenterX, bossCenterY)

			if distance <= proj.Radius+g.Boss.Width/2 {
				wasAlive := g.Boss.Alive
				g.Boss.TakeDamage(proj.Damage)
				proj.Alive = false

				if wasAlive && !g.Boss.Alive {
					g.Player.GainXP(100)
					if g.Player.Class.Type == gamedata.ClassTypeRanged {
						g.Player.Heal(g.Player.Class.KillHealAmount * 5)
					}
				}
				break
			}
		}

		if input.Attack {
			mouseX, mouseY := systems.GetMousePosition()
			worldX, worldY := systems.ScreenToWorld(mouseX, mouseY, g.Camera)

			if worldX >= g.Boss.X && worldX <= g.Boss.X+g.Boss.Width &&
				worldY >= g.Boss.Y && worldY <= g.Boss.Y+g.Boss.Height {

				playerCenterX := g.Player.X + g.Player.Width/2
				playerCenterY := g.Player.Y + g.Player.Height/2
				bossCenterX := g.Boss.X + g.Boss.Width/2
				bossCenterY := g.Boss.Y + g.Boss.Height/2

				distance := systems.GetDistance(playerCenterX, playerCenterY, bossCenterX, bossCenterY)

				switch g.Player.Class.Type {
				case gamedata.ClassTypeMelee:
					if distance <= g.Player.AttackRange {
						wasAlive := g.Boss.Alive
						damage := g.Player.AttackDamage
						g.Boss.TakeDamage(damage)
						lifesteal := int(float32(damage) * g.Player.Class.LifestealPercent)
						g.Player.Heal(lifesteal)

						if wasAlive && !g.Boss.Alive {
							g.Player.GainXP(100)
						}
					}
				case gamedata.ClassTypeCaster:
					if distance <= g.Player.AttackRange && g.Player.CanUseMana(g.Player.Class.ManaCost) {
						wasAlive := g.Boss.Alive
						g.Boss.TakeDamage(g.Player.AttackDamage)
						g.Player.UseMana(g.Player.Class.ManaCost)

						if wasAlive && !g.Boss.Alive {
							g.Player.GainXP(100)
						}
					}
				}
			}
		}
	}

	for i := len(g.Projectiles) - 1; i >= 0; i-- {
		proj := g.Projectiles[i]
		if !proj.Alive {
			g.Projectiles = append(g.Projectiles[:i], g.Projectiles[i+1:]...)
			continue
		}

		proj.X += proj.VX * deltaTime
		proj.Y += proj.VY * deltaTime

		for _, enemy := range g.Enemies {
			if !enemy.Alive {
				continue
			}

			enemyCenterX := enemy.X + enemy.Width/2
			enemyCenterY := enemy.Y + enemy.Height/2
			distance := systems.GetDistance(proj.X, proj.Y, enemyCenterX, enemyCenterY)

			if distance <= proj.Radius+enemy.Width/2 {
				wasAlive := enemy.Alive
				enemy.TakeDamage(proj.Damage)
				proj.Alive = false

				if wasAlive && !enemy.Alive {
					g.Player.GainXP(20)
					if g.Player.Class.Type == gamedata.ClassTypeRanged {
						g.Player.Heal(g.Player.Class.KillHealAmount)
					}
				}
				break
			}
		}

		if g.CurrentRoom != nil {
			if proj.X < g.CurrentRoom.X || proj.X > g.CurrentRoom.X+g.CurrentRoom.Width ||
				proj.Y < g.CurrentRoom.Y || proj.Y > g.CurrentRoom.Y+g.CurrentRoom.Height {
				proj.Alive = false
			}
		}
	}

	if g.LevelUpMenu {
		if rl.IsKeyPressed(rl.KeyOne) {
			g.Player.AddStatPoint(gamedata.StatTypeSTR)
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			g.Player.AddStatPoint(gamedata.StatTypeAGI)
		}
		if rl.IsKeyPressed(rl.KeyThree) {
			g.Player.AddStatPoint(gamedata.StatTypeVIT)
		}
		if rl.IsKeyPressed(rl.KeyFour) {
			g.Player.AddStatPoint(gamedata.StatTypeINT)
		}
		if rl.IsKeyPressed(rl.KeyFive) {
			g.Player.AddStatPoint(gamedata.StatTypeDEX)
		}
		if rl.IsKeyPressed(rl.KeySix) {
			g.Player.AddStatPoint(gamedata.StatTypeLUK)
		}
		if g.Player.StatPoints == 0 {
			g.LevelUpMenu = false
		}
	} else {
		if g.Player.StatPoints > 0 {
			g.LevelUpMenu = true
		}
	}

	if !g.Player.IsAlive() {
		g.EndRun(false)
		return
	}

	if g.CheckRoomCompletion() {
		g.AdvanceToNextRoom()
	}

	worldWidth, worldHeight := g.Dungeon.GetWorldBounds()
	systems.UpdateCamera(g.Camera, g.Player.X+g.Player.Width/2, g.Player.Y+g.Player.Height/2, worldWidth, worldHeight)
}

func (g *Game) Draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	if g.State == RunStateInGame {
		if g.Dungeon != nil {
			for _, room := range g.Dungeon.Rooms {
				systems.DrawRoom(room.X, room.Y, room.Width, room.Height, room.IsBoss(), g.Camera)
			}
		}

		for _, proj := range g.Projectiles {
			if proj.Alive {
				systems.DrawProjectile(proj.X, proj.Y, proj.Radius, g.Camera)
			}
		}

		for _, enemy := range g.Enemies {
			systems.DrawEnemy(enemy, g.Camera)
		}

		if g.Boss != nil {
			systems.DrawBoss(g.Boss, g.Camera)
			for _, proj := range g.Boss.Projectiles {
				if proj.Alive {
					systems.DrawBossProjectile(proj.X, proj.Y, proj.Radius, g.Camera)
				}
			}
		}

		if g.Player != nil {
			systems.DrawPlayer(g.Player, g.Camera)
		}

		roomText := fmt.Sprintf("Room: %d/%d", g.Dungeon.CurrentRoom+1, len(g.Dungeon.Rooms))
		healthText := fmt.Sprintf("Health: %d/%d", g.Player.Health, g.Player.MaxHealth)
		manaText := fmt.Sprintf("Mana: %d/%d", g.Player.Mana, g.Player.MaxMana)
		levelText := fmt.Sprintf("Level: %d | XP: %d/%d", g.Player.Level, g.Player.XP, g.Player.XPToNext)
		rl.DrawText(roomText, 10, 10, 20, rl.Black)
		rl.DrawText(healthText, 10, 35, 20, rl.Black)
		if g.Player.Class.Type == gamedata.ClassTypeCaster {
			rl.DrawText(manaText, 10, 60, 20, rl.Black)
		}
		rl.DrawText(levelText, 10, 85, 20, rl.Black)

		if g.LevelUpMenu {
			rl.DrawRectangle(WindowWidth/2-200, WindowHeight/2-150, 400, 300, rl.NewColor(0, 0, 0, 200))
			rl.DrawText("Level Up! Allocate Stat Points", WindowWidth/2-180, WindowHeight/2-120, 24, rl.White)
			rl.DrawText(fmt.Sprintf("Points: %d", g.Player.StatPoints), WindowWidth/2-180, WindowHeight/2-90, 20, rl.White)
			stats := []string{"1: STR", "2: AGI", "3: VIT", "4: INT", "5: DEX", "6: LUK"}
			statValues := []int{g.Player.Stats.STR, g.Player.Stats.AGI, g.Player.Stats.VIT, g.Player.Stats.INT, g.Player.Stats.DEX, g.Player.Stats.LUK}
			for i, stat := range stats {
				text := fmt.Sprintf("%s: %d", stat, statValues[i])
				rl.DrawText(text, WindowWidth/2-180, WindowHeight/2-60+int32(i*25), 18, rl.White)
			}
		}
	} else if g.State == RunStateMenu {
		rl.DrawText("Single Fantasy", WindowWidth/2-100, WindowHeight/2-100, 40, rl.Black)
		classNames := []string{"Warrior (1)", "Ranger (2)", "Mage (3)"}
		for i, name := range classNames {
			color := rl.Black
			if int(g.SelectedClass) == i {
				color = rl.Blue
			}
			rl.DrawText(name, WindowWidth/2-80, WindowHeight/2-40+int32(i*30), 20, color)
		}
		rl.DrawText("Press SPACE to start", WindowWidth/2-120, WindowHeight/2+60, 20, rl.Black)
	} else if g.State == RunStateRewardSelection {
		rl.DrawRectangle(WindowWidth/2-300, WindowHeight/2-200, 600, 400, rl.NewColor(0, 0, 0, 220))
		rl.DrawText("Select a Reward", WindowWidth/2-100, WindowHeight/2-180, 30, rl.White)

		for i, item := range g.RewardOptions {
			if item == nil {
				continue
			}

			y := WindowHeight/2 - 120 + int32(i*100)
			color := rl.White
			if g.SelectedReward == i {
				color = rl.Yellow
				rl.DrawRectangle(WindowWidth/2-290, y-5, 580, 90, rl.NewColor(255, 255, 0, 50))
			}

			rl.DrawText(fmt.Sprintf("%d: %s", i+1, item.Name), WindowWidth/2-280, y, 24, color)
			rl.DrawText(item.Description, WindowWidth/2-280, y+30, 18, rl.Gray)

			bonusText := "Bonuses: "
			first := true
			for statType, bonus := range item.StatBonuses {
				if !first {
					bonusText += ", "
				}
				statName := ""
				switch statType {
				case gamedata.StatTypeSTR:
					statName = "STR"
				case gamedata.StatTypeAGI:
					statName = "AGI"
				case gamedata.StatTypeVIT:
					statName = "VIT"
				case gamedata.StatTypeINT:
					statName = "INT"
				case gamedata.StatTypeDEX:
					statName = "DEX"
				case gamedata.StatTypeLUK:
					statName = "LUK"
				}
				bonusText += fmt.Sprintf("%s +%d", statName, bonus)
				first = false
			}
			rl.DrawText(bonusText, WindowWidth/2-280, y+55, 16, rl.LightGray)
		}

		rl.DrawText("Press ENTER to select, 1-3 to choose", WindowWidth/2-180, WindowHeight/2+180, 18, rl.White)
	} else if g.State == RunStateVictory {
		rl.DrawText("Victory!", WindowWidth/2-60, WindowHeight/2-20, 40, rl.Green)
		rl.DrawText("Press SPACE to restart", WindowWidth/2-120, WindowHeight/2+30, 20, rl.Black)
	} else if g.State == RunStateDefeat {
		rl.DrawText("Defeat!", WindowWidth/2-60, WindowHeight/2-20, 40, rl.Red)
		rl.DrawText("Press SPACE to restart", WindowWidth/2-120, WindowHeight/2+30, 20, rl.Black)
	}

	rl.EndDrawing()
}

func (g *Game) HandleMenuInput() {
	if rl.IsKeyPressed(rl.KeyOne) {
		g.SelectedClass = gamedata.ClassTypeMelee
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		g.SelectedClass = gamedata.ClassTypeRanged
	}
	if rl.IsKeyPressed(rl.KeyThree) {
		g.SelectedClass = gamedata.ClassTypeCaster
	}
	if rl.IsKeyPressed(rl.KeySpace) {
		g.StartRun()
	}
}

func (g *Game) HandleGameOverInput() {
	if rl.IsKeyPressed(rl.KeySpace) {
		g.State = RunStateMenu
		g.ResetState()
	}
}

func (g *Game) HandleRewardSelectionInput() {
	if rl.IsKeyPressed(rl.KeyOne) {
		g.SelectedReward = 0
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		g.SelectedReward = 1
	}
	if rl.IsKeyPressed(rl.KeyThree) {
		g.SelectedReward = 2
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		if g.SelectedReward >= 0 && g.SelectedReward < len(g.RewardOptions) {
			item := g.RewardOptions[g.SelectedReward]
			if item != nil {
				g.Player.EquipItem(item)
			}
		}
		g.State = RunStateMenu
		g.ResetState()
	}
}
