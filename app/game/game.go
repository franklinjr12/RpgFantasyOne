package game

import (
	"fmt"
	"log"
	"math"

	"singlefantasy/app/assets"
	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/systems"
	"singlefantasy/app/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AppState int

const (
	StateBoot AppState = iota
	StateMainMenu
	StateClassSelect
	StateRun
	StateReward
	StateResults
)

type RunResults struct {
	Victory            bool
	RunDurationSeconds float32
	RoomsCleared       int
	TotalRooms         int
	SelectedClass      gamedata.ClassType
	RewardPicked       string
}

type Game struct {
	State               AppState
	Player              *gameobjects.Player
	Enemies             []*gameobjects.Enemy
	Boss                *gameobjects.Boss
	Dungeon             *world.Dungeon
	Camera              *systems.Camera
	Projectiles         []*Projectile
	CurrentRoom         *world.Room
	SelectedClass       gamedata.ClassType
	LevelUpMenu         bool
	RewardOptions       []*gamedata.Item
	SelectedReward      int
	PlayerMoveTargetX   float32
	PlayerMoveTargetY   float32
	HasPlayerMoveTarget bool
	PlayerAttackTarget  interface{}
	DebugOverlayEnabled bool
	BootCompleted       bool
	LastFrameTime       float32
	LastUpdateSteps     int
	RunElapsed          float32
	Results             RunResults
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
	Skill  *gamedata.Skill
	Caster *gameobjects.Player
}

func NewGame() *Game {
	return &Game{
		State:               StateBoot,
		Player:              nil,
		Enemies:             []*gameobjects.Enemy{},
		Boss:                nil,
		Dungeon:             nil,
		Camera:              systems.NewCamera(),
		Projectiles:         []*Projectile{},
		CurrentRoom:         nil,
		SelectedClass:       gamedata.ClassTypeMelee,
		LevelUpMenu:         false,
		RewardOptions:       []*gamedata.Item{},
		SelectedReward:      0,
		PlayerMoveTargetX:   0,
		PlayerMoveTargetY:   0,
		HasPlayerMoveTarget: false,
		PlayerAttackTarget:  nil,
		DebugOverlayEnabled: false,
		BootCompleted:       false,
		LastFrameTime:       0,
		LastUpdateSteps:     0,
		RunElapsed:          0,
		Results:             RunResults{},
	}
}

func (g *Game) SetFrameDiagnostics(frameTime float32, updateSteps int) {
	g.LastFrameTime = frameTime
	g.LastUpdateSteps = updateSteps
}

func (g *Game) UpdateFrame() {
	if rl.IsKeyPressed(DebugToggleKey) {
		g.DebugOverlayEnabled = !g.DebugOverlayEnabled
	}

	switch g.State {
	case StateBoot:
		g.updateBoot()
	case StateMainMenu:
		g.updateMainMenu()
	case StateClassSelect:
		g.updateClassSelect()
	case StateReward:
		g.updateReward()
	case StateResults:
		g.updateResults()
	}
}

func (g *Game) UpdateFixed(deltaTime float32) {
	if g.State == StateRun {
		g.updateRun(deltaTime)
	}
}

func (g *Game) updateBoot() {
	if !g.BootCompleted {
		manager := assets.Get()
		manager.LoadTexture(
			systems.HumanoidSpriteSheetAssetKey,
			"resources/sprites/Basic Humanoid Sprites 4x.png",
			288,
			288,
			rl.Magenta,
		)
		manager.LoadFont(assets.FontDefault, "")
		manager.LoadSound("sfx.ui.confirm", "resources/audio/ui_confirm.wav")
		manager.LoadMusic("music.menu", "resources/audio/menu.ogg")
		g.BootCompleted = true
	}

	g.EnterMainMenu()
}

func (g *Game) EnterMainMenu() {
	g.ResetState()
	g.State = StateMainMenu
}

func (g *Game) EnterClassSelect() {
	g.State = StateClassSelect
}

func (g *Game) StartRun() {
	g.ResetState()
	g.Dungeon = world.NewDungeon()
	g.CurrentRoom = g.Dungeon.GetCurrentRoom()

	if g.CurrentRoom == nil {
		g.EnterResults(false, "")
		return
	}

	startX := g.CurrentRoom.X + g.CurrentRoom.Width/2
	startY := g.CurrentRoom.Y + g.CurrentRoom.Height/2
	g.Player = gameobjects.NewPlayer(startX, startY, g.SelectedClass)

	g.SpawnRoomEnemies()
	g.RunElapsed = 0
	g.State = StateRun
}

func (g *Game) EnterReward() {
	if g.Player == nil {
		g.EnterResults(false, "")
		return
	}

	g.RewardOptions = gamedata.GenerateRewardOptions(g.Player.Class.Type)
	g.SelectedReward = 0
	g.State = StateReward
}

func (g *Game) EnterResults(victory bool, rewardPicked string) {
	totalRooms := 0
	roomsCleared := 0
	if g.Dungeon != nil {
		totalRooms = len(g.Dungeon.Rooms)
		for _, room := range g.Dungeon.Rooms {
			if room.Completed {
				roomsCleared++
			}
		}
	}

	g.Results = RunResults{
		Victory:            victory,
		RunDurationSeconds: g.RunElapsed,
		RoomsCleared:       roomsCleared,
		TotalRooms:         totalRooms,
		SelectedClass:      g.SelectedClass,
		RewardPicked:       rewardPicked,
	}
	g.State = StateResults
}

func (g *Game) ResetState() {
	g.Player = nil
	g.Enemies = []*gameobjects.Enemy{}
	g.Boss = nil
	g.Dungeon = nil
	g.Projectiles = []*Projectile{}
	g.CurrentRoom = nil
	g.Camera = systems.NewCamera()
	g.LevelUpMenu = false
	g.RewardOptions = []*gamedata.Item{}
	g.SelectedReward = 0
	g.PlayerMoveTargetX = 0
	g.PlayerMoveTargetY = 0
	g.HasPlayerMoveTarget = false
	g.PlayerAttackTarget = nil
	g.RunElapsed = 0
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
		g.EnterReward()
		return
	}

	g.CurrentRoom = g.Dungeon.GetCurrentRoom()
	if g.CurrentRoom == nil || g.Player == nil {
		return
	}

	startX := g.CurrentRoom.X + g.CurrentRoom.Width/2
	startY := g.CurrentRoom.Y + g.CurrentRoom.Height/2
	g.Player.X = startX
	g.Player.Y = startY
	g.Player.Health = g.Player.MaxHealth

	g.SpawnRoomEnemies()
	g.Projectiles = []*Projectile{}
}

func (g *Game) IsMenuOpen() bool {
	return g.LevelUpMenu
}

func (g *Game) updateMainMenu() {
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		g.EnterClassSelect()
	}
}

func (g *Game) updateClassSelect() {
	if rl.IsKeyPressed(rl.KeyOne) {
		g.SelectedClass = gamedata.ClassTypeMelee
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		g.SelectedClass = gamedata.ClassTypeRanged
	}
	if rl.IsKeyPressed(rl.KeyThree) {
		g.SelectedClass = gamedata.ClassTypeCaster
	}
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		g.StartRun()
	}
}

func (g *Game) updateReward() {
	if rl.IsKeyPressed(rl.KeyOne) {
		g.SelectedReward = 0
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		g.SelectedReward = 1
	}
	if rl.IsKeyPressed(rl.KeyThree) {
		g.SelectedReward = 2
	}

	if !rl.IsKeyPressed(rl.KeyEnter) {
		return
	}

	rewardPicked := "None"
	if g.Player != nil && g.SelectedReward >= 0 && g.SelectedReward < len(g.RewardOptions) {
		item := g.RewardOptions[g.SelectedReward]
		if item != nil {
			g.Player.EquipItem(item)
			rewardPicked = item.Name
		}
	}
	g.EnterResults(true, rewardPicked)
}

func (g *Game) updateResults() {
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		g.EnterMainMenu()
	}
}

func (g *Game) UpdateAutoAttack(_ float32) {
	if g.Player == nil || g.PlayerAttackTarget == nil {
		return
	}

	var targetAlive bool
	var targetX, targetY float32
	var targetWidth, targetHeight float32

	switch t := g.PlayerAttackTarget.(type) {
	case *gameobjects.Enemy:
		if !t.Alive {
			g.PlayerAttackTarget = nil
			return
		}
		targetAlive = true
		targetX = t.X
		targetY = t.Y
		targetWidth = t.Width
		targetHeight = t.Height
	case *gameobjects.Boss:
		if t.Enemy == nil || !t.Alive {
			g.PlayerAttackTarget = nil
			return
		}
		targetAlive = true
		targetX = t.X
		targetY = t.Y
		targetWidth = t.Width
		targetHeight = t.Height
	default:
		g.PlayerAttackTarget = nil
		return
	}

	if !targetAlive {
		g.PlayerAttackTarget = nil
		return
	}

	playerCenterX := g.Player.X + g.Player.Width/2
	playerCenterY := g.Player.Y + g.Player.Height/2
	targetCenterX := targetX + targetWidth/2
	targetCenterY := targetY + targetHeight/2

	distance := systems.GetDistance(playerCenterX, playerCenterY, targetCenterX, targetCenterY)

	if distance > g.Player.AttackRange {
		g.PlayerMoveTargetX = targetCenterX
		g.PlayerMoveTargetY = targetCenterY
		g.HasPlayerMoveTarget = true
		return
	}

	if g.Player.Class.Type == gamedata.ClassTypeRanged {
		g.HasPlayerMoveTarget = false
	}

	if g.Player.CurrentAttackCooldown > 0 {
		return
	}

	damage := g.Player.GetAutoAttackDamage()
	cooldown := g.Player.GetAttackCooldown()

	switch g.Player.Class.Type {
	case gamedata.ClassTypeMelee:
		var wasAlive bool
		switch t := g.PlayerAttackTarget.(type) {
		case *gameobjects.Enemy:
			wasAlive = t.Alive
			t.TakeDamage(damage)
			lifesteal := int(float32(damage) * g.Player.Class.LifestealPercent)
			g.Player.Heal(lifesteal)

			if wasAlive && !t.Alive {
				g.Player.GainXP(20)
				g.PlayerAttackTarget = nil
			}
		case *gameobjects.Boss:
			wasAlive = t.Alive
			t.TakeDamage(damage)
			lifesteal := int(float32(damage) * g.Player.Class.LifestealPercent)
			g.Player.Heal(lifesteal)

			if wasAlive && !t.Alive {
				g.Player.GainXP(100)
				g.PlayerAttackTarget = nil
			}
		}

	case gamedata.ClassTypeRanged:
		dx := targetCenterX - playerCenterX
		dy := targetCenterY - playerCenterY
		dist := systems.GetDistance(0, 0, dx, dy)
		if dist > 0 {
			speed := float32(400)
			proj := &Projectile{
				X:      playerCenterX,
				Y:      playerCenterY,
				VX:     (dx / dist) * speed,
				VY:     (dy / dist) * speed,
				Speed:  speed,
				Damage: damage,
				Radius: 5,
				Alive:  true,
			}
			g.Projectiles = append(g.Projectiles, proj)
		}

	case gamedata.ClassTypeCaster:
		if !g.Player.CanUseMana(g.Player.Class.ManaCost) {
			return
		}

		var wasAlive bool
		switch t := g.PlayerAttackTarget.(type) {
		case *gameobjects.Enemy:
			wasAlive = t.Alive
			t.TakeDamage(damage)
			g.Player.UseMana(g.Player.Class.ManaCost)

			if wasAlive && !t.Alive {
				g.Player.GainXP(20)
				g.PlayerAttackTarget = nil
			}
		case *gameobjects.Boss:
			wasAlive = t.Alive
			t.TakeDamage(damage)
			g.Player.UseMana(g.Player.Class.ManaCost)

			if wasAlive && !t.Alive {
				g.Player.GainXP(100)
				g.PlayerAttackTarget = nil
			}
		}
	}

	g.Player.CurrentAttackCooldown = cooldown
}

func (g *Game) updateRun(deltaTime float32) {
	if g.Player == nil {
		return
	}

	input := systems.UpdateInput(g.Camera)
	isMenuOpen := g.IsMenuOpen()

	skillInputs := []bool{input.Skill1, input.Skill2, input.Skill3, input.Skill4}
	for i, skillPressed := range skillInputs {
		if skillPressed && i < len(g.Player.Skills) {
			skill := g.Player.Skills[i]
			if systems.CanCast(g.Player, skill) {
				g.PlayerAttackTarget = nil
				g.TryCastSkill(skill, input)
			}
		}
	}

	if !isMenuOpen {
		if input.HasMoveTarget {
			g.PlayerMoveTargetX = input.MoveToX
			g.PlayerMoveTargetY = input.MoveToY
			g.HasPlayerMoveTarget = true
			g.PlayerAttackTarget = nil
		}

		g.UpdateAutoAttack(deltaTime)

		if g.HasPlayerMoveTarget {
			playerCenterX := g.Player.X + g.Player.Width/2
			playerCenterY := g.Player.Y + g.Player.Height/2
			dx := g.PlayerMoveTargetX - playerCenterX
			dy := g.PlayerMoveTargetY - playerCenterY
			distance := systems.GetDistance(0, 0, dx, dy)

			if distance > 5 {
				moveSpeed := g.GetPlayerMoveSpeed()
				moveDistance := moveSpeed * deltaTime
				if moveDistance > distance {
					moveDistance = distance
				}
				g.Player.X += (dx / distance) * moveDistance
				g.Player.Y += (dy / distance) * moveDistance
			} else {
				g.HasPlayerMoveTarget = false
			}
		}
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

		targetFound := false

		for _, enemy := range g.Enemies {
			if !enemy.Alive {
				continue
			}

			if worldX >= enemy.X && worldX <= enemy.X+enemy.Width &&
				worldY >= enemy.Y && worldY <= enemy.Y+enemy.Height {
				g.PlayerAttackTarget = enemy
				g.HasPlayerMoveTarget = false
				targetFound = true
				break
			}
		}

		if !targetFound && g.Boss != nil && g.Boss.Alive {
			if worldX >= g.Boss.X && worldX <= g.Boss.X+g.Boss.Width &&
				worldY >= g.Boss.Y && worldY <= g.Boss.Y+g.Boss.Height {
				g.PlayerAttackTarget = g.Boss
				g.HasPlayerMoveTarget = false
			}
		}
	}

	if !isMenuOpen {
		g.RunElapsed += deltaTime
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
					if proj.Skill != nil && proj.Caster != nil {
						systems.ApplySkill(proj.Caster, proj.Skill, []interface{}{g.Boss})
					} else {
						g.Boss.TakeDamage(proj.Damage)
					}
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
		}
	}

	if !isMenuOpen {
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
					if proj.Skill != nil && proj.Caster != nil {
						systems.ApplySkill(proj.Caster, proj.Skill, []interface{}{enemy})
					} else {
						enemy.TakeDamage(proj.Damage)
					}
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
	} else if g.Player.StatPoints > 0 {
		g.LevelUpMenu = true
	}

	if !g.Player.IsAlive() {
		g.EnterResults(false, "")
		return
	}

	if g.CheckRoomCompletion() {
		if g.CurrentRoom != nil && g.CurrentRoom.IsBoss() {
			g.EnterReward()
		} else {
			g.AdvanceToNextRoom()
		}
	}

	if g.Dungeon != nil {
		worldWidth, worldHeight := g.Dungeon.GetWorldBounds()
		systems.UpdateCamera(
			g.Camera,
			g.Player.X+g.Player.Width/2,
			g.Player.Y+g.Player.Height/2,
			worldWidth,
			worldHeight,
		)
	}
}

func (g *Game) Draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	switch g.State {
	case StateBoot:
		g.drawBootScreen()
	case StateMainMenu:
		g.drawMainMenu()
	case StateClassSelect:
		g.drawClassSelect()
	case StateRun:
		g.drawRun()
	case StateReward:
		g.drawRewardSelection()
	case StateResults:
		g.drawResults()
	}

	if g.DebugOverlayEnabled {
		systems.DrawDebugOverlay(g.GetDebugLines())
	}

	rl.EndDrawing()
}

func (g *Game) drawBootScreen() {
	rl.DrawText("Booting...", WindowWidth/2-80, WindowHeight/2-20, 40, rl.Black)
}

func (g *Game) drawMainMenu() {
	rl.DrawText("Single Fantasy", WindowWidth/2-140, WindowHeight/2-120, 48, rl.Black)
	rl.DrawText("Press ENTER or SPACE to Start", WindowWidth/2-170, WindowHeight/2-20, 24, rl.DarkGray)
	rl.DrawText("F3 toggles debug overlay", WindowWidth/2-130, WindowHeight/2+20, 20, rl.Gray)
}

func (g *Game) drawClassSelect() {
	rl.DrawText("Select Class", WindowWidth/2-110, WindowHeight/2-140, 40, rl.Black)
	classNames := []string{"1) Warrior", "2) Ranger", "3) Mage"}
	for i, name := range classNames {
		color := rl.Black
		if int(g.SelectedClass) == i {
			color = rl.Blue
		}
		rl.DrawText(name, WindowWidth/2-80, WindowHeight/2-60+int32(i*35), 26, color)
	}
	rl.DrawText("Press ENTER or SPACE to Confirm", WindowWidth/2-180, WindowHeight/2+90, 22, rl.DarkGray)
}

func (g *Game) drawRun() {
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
		systems.DrawSkillBar(g.Player)
	}

	if g.Player == nil || g.Dungeon == nil {
		return
	}

	roomText := fmt.Sprintf("Room: %d/%d", g.Dungeon.CurrentRoom+1, len(g.Dungeon.Rooms))
	healthText := fmt.Sprintf("Health: %d/%d", g.Player.Health, g.Player.MaxHealth)
	manaText := fmt.Sprintf("Mana: %d/%d", g.Player.Mana, g.Player.MaxMana)
	levelText := fmt.Sprintf("Level: %d | XP: %d/%d", g.Player.Level, g.Player.XP, g.Player.XPToNext)
	rl.DrawText(roomText, 10, 40, 20, rl.Black)
	rl.DrawText(healthText, 10, 65, 20, rl.Black)
	if g.Player.Class.Type == gamedata.ClassTypeCaster {
		rl.DrawText(manaText, 10, 90, 20, rl.Black)
	}
	rl.DrawText(levelText, 10, 115, 20, rl.Black)

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
}

func (g *Game) drawRewardSelection() {
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

	rl.DrawText("Press ENTER to confirm, 1-3 to choose", WindowWidth/2-190, WindowHeight/2+180, 18, rl.White)
}

func (g *Game) drawResults() {
	title := "Run Complete"
	titleColor := rl.DarkGray
	if g.Results.Victory {
		title = "Victory!"
		titleColor = rl.Green
	} else {
		title = "Defeat!"
		titleColor = rl.Red
	}

	rl.DrawText(title, WindowWidth/2-90, WindowHeight/2-180, 48, titleColor)

	className := "Warrior"
	switch g.Results.SelectedClass {
	case gamedata.ClassTypeRanged:
		className = "Ranger"
	case gamedata.ClassTypeCaster:
		className = "Mage"
	}

	rl.DrawText(fmt.Sprintf("Class: %s", className), WindowWidth/2-150, WindowHeight/2-90, 26, rl.Black)
	rl.DrawText(fmt.Sprintf("Rooms Cleared: %d/%d", g.Results.RoomsCleared, g.Results.TotalRooms), WindowWidth/2-150, WindowHeight/2-55, 26, rl.Black)
	rl.DrawText(fmt.Sprintf("Run Time: %.1fs", g.Results.RunDurationSeconds), WindowWidth/2-150, WindowHeight/2-20, 26, rl.Black)
	rl.DrawText(fmt.Sprintf("Reward Picked: %s", g.Results.RewardPicked), WindowWidth/2-150, WindowHeight/2+15, 26, rl.Black)
	rl.DrawText("Press ENTER or SPACE to return to Main Menu", WindowWidth/2-230, WindowHeight/2+90, 24, rl.DarkGray)
}

func (g *Game) GetStateName() string {
	switch g.State {
	case StateBoot:
		return "Boot"
	case StateMainMenu:
		return "MainMenu"
	case StateClassSelect:
		return "ClassSelect"
	case StateRun:
		return "Run"
	case StateReward:
		return "Reward"
	case StateResults:
		return "Results"
	default:
		return "Unknown"
	}
}

func (g *Game) GetDebugLines() []string {
	lines := []string{
		fmt.Sprintf("FPS: %d", rl.GetFPS()),
		fmt.Sprintf("State: %s", g.GetStateName()),
		fmt.Sprintf("Fixed updates/frame: %d", g.LastUpdateSteps),
		fmt.Sprintf("Frame time (clamped): %.3f s", g.LastFrameTime),
	}

	if g.State == StateRun {
		roomIndex := 0
		totalRooms := 0
		if g.Dungeon != nil {
			roomIndex = g.Dungeon.CurrentRoom + 1
			totalRooms = len(g.Dungeon.Rooms)
		}

		aliveEnemies := 0
		for _, enemy := range g.Enemies {
			if enemy.Alive {
				aliveEnemies++
			}
		}

		bossProjectiles := 0
		if g.Boss != nil {
			for _, proj := range g.Boss.Projectiles {
				if proj.Alive {
					bossProjectiles++
				}
			}
		}

		activeProjectiles := 0
		for _, proj := range g.Projectiles {
			if proj.Alive {
				activeProjectiles++
			}
		}

		lines = append(lines, fmt.Sprintf("Room: %d/%d", roomIndex, totalRooms))
		lines = append(lines, fmt.Sprintf("Enemies (alive/total): %d/%d", aliveEnemies, len(g.Enemies)))
		lines = append(lines, fmt.Sprintf("Projectiles (player/boss): %d/%d", activeProjectiles, bossProjectiles))
	}

	return lines
}

func (g *Game) GetPlayerMoveSpeed() float32 {
	if g.Player == nil {
		return 0
	}

	speed := g.Player.MoveSpeed

	if gamedata.HasEffect(&g.Player.Effects, gamedata.EffectSlow) {
		magnitude := gamedata.GetEffectMagnitude(&g.Player.Effects, gamedata.EffectSlow)
		speed *= (1.0 - magnitude)
	}
	if gamedata.HasEffect(&g.Player.Effects, gamedata.EffectFreeze) {
		return 0
	}
	if gamedata.HasEffect(&g.Player.Effects, gamedata.EffectStun) {
		return 0
	}
	if gamedata.HasEffect(&g.Player.Effects, gamedata.EffectMoveSpeedReduction) {
		magnitude := gamedata.GetEffectMagnitude(&g.Player.Effects, gamedata.EffectMoveSpeedReduction)
		speed *= (1.0 - magnitude)
	}
	if gamedata.HasEffect(&g.Player.Effects, gamedata.EffectMoveSpeedBoost) {
		magnitude := gamedata.GetEffectMagnitude(&g.Player.Effects, gamedata.EffectMoveSpeedBoost)
		speed *= (1.0 + magnitude)
	}

	return speed
}
