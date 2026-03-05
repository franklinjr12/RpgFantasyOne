package game

import (
	"math"
	"strings"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/systems"
	"singlefantasy/app/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type RuntimeContext struct {
	Game        *Game
	Player      *gameobjects.Player
	Enemies     *[]*gameobjects.Enemy
	Boss        **gameobjects.Boss
	Projectiles *[]*Projectile
	Dungeon     **world.Dungeon
	CurrentRoom **world.Room
	Camera      **systems.Camera
	Input       *systems.Input
	IsMenuOpen  bool
}

func NewRuntimeContext(game *Game) *RuntimeContext {
	return &RuntimeContext{
		Game:        game,
		Player:      game.Player,
		Enemies:     &game.Enemies,
		Boss:        &game.Boss,
		Projectiles: &game.Projectiles,
		Dungeon:     &game.Dungeon,
		CurrentRoom: &game.CurrentRoom,
		Camera:      &game.Camera,
	}
}

type RuntimeSystem interface {
	Name() string
	Update(ctx *RuntimeContext, dt float32)
}

type RuntimePipeline struct {
	systems []RuntimeSystem
}

func NewRuntimePipeline() *RuntimePipeline {
	return &RuntimePipeline{
		systems: []RuntimeSystem{
			// Required fixed runtime order: Input -> AI -> Casting -> Projectiles -> Movement -> Combat Resolve -> Effects -> Dungeon/Run -> UI/Render Prep.
			&inputSystem{},
			&aiSystem{},
			&castingSystem{},
			&projectilesSystem{},
			&movementSystem{},
			&combatResolveSystem{},
			&effectsSystem{},
			&dungeonRunSystem{},
			&uiRenderPrepSystem{},
		},
	}
}

func (p *RuntimePipeline) Update(ctx *RuntimeContext, dt float32) {
	for _, system := range p.systems {
		system.Update(ctx, dt)
	}
}

func (p *RuntimePipeline) OrderString() string {
	parts := make([]string, 0, len(p.systems))
	for _, system := range p.systems {
		parts = append(parts, system.Name())
	}
	return strings.Join(parts, " -> ")
}

type inputSystem struct{}

func (s *inputSystem) Name() string { return "Input" }

func (s *inputSystem) Update(ctx *RuntimeContext, _ float32) {
	g := ctx.Game
	ctx.Input = systems.UpdateInput(g.Camera)
	ctx.IsMenuOpen = g.IsMenuOpen()

	if g.RoomTransitionTimer > 0 || g.PendingRoomTransition {
		return
	}

	if !ctx.IsMenuOpen && ctx.Input.HasMoveTarget {
		g.PlayerMoveTargetX = ctx.Input.MoveToX
		g.PlayerMoveTargetY = ctx.Input.MoveToY
		g.HasPlayerMoveTarget = true
		g.PlayerAttackTarget = nil
	}

	if !ctx.Input.Attack {
		return
	}

	mouseX, mouseY := systems.GetMousePosition()
	worldX, worldY := systems.ScreenToWorldIso(mouseX, mouseY, g.Camera)

	for _, enemy := range g.Enemies {
		if !enemy.IsAlive() {
			continue
		}

		if worldX >= enemy.PosX && worldX <= enemy.PosX+enemy.Hitbox.Width &&
			worldY >= enemy.PosY && worldY <= enemy.PosY+enemy.Hitbox.Height {
			g.PlayerAttackTarget = enemy
			g.HasPlayerMoveTarget = false
			return
		}
	}

	if g.Boss != nil && g.Boss.IsAlive() {
		if worldX >= g.Boss.PosX && worldX <= g.Boss.PosX+g.Boss.Hitbox.Width &&
			worldY >= g.Boss.PosY && worldY <= g.Boss.PosY+g.Boss.Hitbox.Height {
			g.PlayerAttackTarget = g.Boss
			g.HasPlayerMoveTarget = false
		}
	}
}

type aiSystem struct{}

func (s *aiSystem) Name() string { return "AI" }

func (s *aiSystem) Update(ctx *RuntimeContext, dt float32) {
	g := ctx.Game
	if ctx.IsMenuOpen || g.Player == nil {
		return
	}

	playerX, playerY := g.Player.Center()
	for _, enemy := range g.Enemies {
		enemy.Update(dt, playerX, playerY)
	}

	if g.Boss == nil {
		return
	}

	g.Boss.Update(dt, playerX, playerY)
	if g.Boss.ShouldSpawnAdds() {
		for i := 0; i < 2; i++ {
			angle := float32(i) * 3.14159 * 2.0 / 2.0
			addX := g.Boss.PosX + g.Boss.Hitbox.Width/2 + float32(math.Cos(float64(angle)))*100
			addY := g.Boss.PosY + g.Boss.Hitbox.Height/2 + float32(math.Sin(float64(angle)))*100
			add := gameobjects.NewEnemy(addX, addY, false)
			g.Enemies = append(g.Enemies, add)
		}
		g.Boss.ResetAddSpawnTimer()
	}
}

type castingSystem struct{}

func (s *castingSystem) Name() string { return "Casting" }

func (s *castingSystem) Update(ctx *RuntimeContext, _ float32) {
	g := ctx.Game
	if g.Player == nil || ctx.Input == nil {
		return
	}

	skillInputs := []bool{ctx.Input.Skill1, ctx.Input.Skill2, ctx.Input.Skill3, ctx.Input.Skill4}
	for i, skillPressed := range skillInputs {
		if !skillPressed || i >= len(g.Player.Skills) {
			continue
		}
		skill := g.Player.Skills[i]
		if !systems.CanCast(g.Player, skill) {
			continue
		}
		g.PlayerAttackTarget = nil
		g.TryCastSkill(skill, ctx.Input)
	}
}

type projectilesSystem struct{}

func (s *projectilesSystem) Name() string { return "Projectiles" }

func (s *projectilesSystem) Update(ctx *RuntimeContext, dt float32) {
	g := ctx.Game
	if ctx.IsMenuOpen || g.Player == nil {
		return
	}

	s.updateBossProjectiles(g, dt)
	s.resolveBossHits(g)
	s.updatePlayerProjectiles(g, dt)
}

func (s *projectilesSystem) updateBossProjectiles(g *Game, dt float32) {
	if g.Boss == nil {
		return
	}

	for i := len(g.Boss.Projectiles) - 1; i >= 0; i-- {
		proj := g.Boss.Projectiles[i]
		if !proj.Alive {
			g.Boss.Projectiles = append(g.Boss.Projectiles[:i], g.Boss.Projectiles[i+1:]...)
			continue
		}

		proj.X += proj.VX * dt
		proj.Y += proj.VY * dt

		playerCenterX, playerCenterY := g.Player.Center()
		distance := systems.GetDistance(proj.X, proj.Y, playerCenterX, playerCenterY)
		if distance <= proj.Radius+g.Player.Hitbox.Width/2 {
			g.ApplyPlayerDirectHit(proj.Damage, proj.X, proj.Y)
			proj.Alive = false
		}

		if g.CurrentRoom != nil {
			if proj.X < g.CurrentRoom.X || proj.X > g.CurrentRoom.X+g.CurrentRoom.Width ||
				proj.Y < g.CurrentRoom.Y || proj.Y > g.CurrentRoom.Y+g.CurrentRoom.Height {
				proj.Alive = false
			}
		}
	}
}

func (s *projectilesSystem) resolveBossHits(g *Game) {
	if g.Boss == nil || !g.Boss.IsAlive() {
		return
	}

	for _, proj := range g.Projectiles {
		if !proj.Alive {
			continue
		}

		bossX, bossY := g.Boss.Center()
		distance := systems.GetDistance(proj.X, proj.Y, bossX, bossY)
		if distance > proj.Radius+g.Boss.Hitbox.Width/2 {
			continue
		}

		wasAlive := g.Boss.IsAlive()
		if proj.Skill != nil && proj.Caster != nil {
			systems.ApplySkill(proj.Caster, proj.Skill, []interface{}{g.Boss})
		} else {
			g.Boss.TakeDamage(proj.Damage)
		}
		proj.Alive = false

		if wasAlive && !g.Boss.IsAlive() {
			g.Player.GainXP(100)
			if g.Player.Class.Type == gamedata.ClassTypeRanged {
				g.Player.Heal(g.Player.Class.KillHealAmount * 5)
			}
		}
		break
	}
}

func (s *projectilesSystem) updatePlayerProjectiles(g *Game, dt float32) {
	for i := len(g.Projectiles) - 1; i >= 0; i-- {
		proj := g.Projectiles[i]
		if !proj.Alive {
			g.Projectiles = append(g.Projectiles[:i], g.Projectiles[i+1:]...)
			continue
		}

		proj.X += proj.VX * dt
		proj.Y += proj.VY * dt

		for _, enemy := range g.Enemies {
			if !enemy.IsAlive() {
				continue
			}

			enemyX, enemyY := enemy.Center()
			distance := systems.GetDistance(proj.X, proj.Y, enemyX, enemyY)
			if distance > proj.Radius+enemy.Hitbox.Width/2 {
				continue
			}

			wasAlive := enemy.IsAlive()
			if proj.Skill != nil && proj.Caster != nil {
				systems.ApplySkill(proj.Caster, proj.Skill, []interface{}{enemy})
			} else {
				enemy.TakeDamage(proj.Damage)
			}
			proj.Alive = false

			if wasAlive && !enemy.IsAlive() {
				g.Player.GainXP(20)
				if g.Player.Class.Type == gamedata.ClassTypeRanged {
					g.Player.Heal(g.Player.Class.KillHealAmount)
				}
			}
			break
		}

		if g.CurrentRoom != nil {
			if proj.X < g.CurrentRoom.X || proj.X > g.CurrentRoom.X+g.CurrentRoom.Width ||
				proj.Y < g.CurrentRoom.Y || proj.Y > g.CurrentRoom.Y+g.CurrentRoom.Height {
				proj.Alive = false
			}
		}
	}
}

type movementSystem struct{}

func (s *movementSystem) Name() string { return "Movement" }

func (s *movementSystem) Update(ctx *RuntimeContext, dt float32) {
	g := ctx.Game
	if ctx.IsMenuOpen || g.Player == nil {
		return
	}
	if g.RoomTransitionTimer > 0 || g.PendingRoomTransition {
		return
	}

	g.UpdateAutoAttack(dt)

	desiredVelX := float32(0)
	desiredVelY := float32(0)

	if g.HasPlayerMoveTarget {
		playerCenterX, playerCenterY := g.Player.Center()
		dx := g.PlayerMoveTargetX - playerCenterX
		dy := g.PlayerMoveTargetY - playerCenterY
		distance := systems.GetDistance(0, 0, dx, dy)

		if distance > PlayerMoveTargetStopDistance {
			moveSpeed := g.GetPlayerMoveSpeed()
			if distance < PlayerMoveTargetSlowRadius {
				moveSpeed *= distance / PlayerMoveTargetSlowRadius
			}
			desiredVelX = (dx / distance) * moveSpeed
			desiredVelY = (dy / distance) * moveSpeed
		} else {
			g.HasPlayerMoveTarget = false
		}
	}

	g.Player.MoveVelocityX = smoothAxisVelocity(g.Player.MoveVelocityX, desiredVelX, PlayerMoveAcceleration, PlayerMoveDeceleration, dt)
	g.Player.MoveVelocityY = smoothAxisVelocity(g.Player.MoveVelocityY, desiredVelY, PlayerMoveAcceleration, PlayerMoveDeceleration, dt)
	g.Player.KnockbackVelX = decayAxisVelocity(g.Player.KnockbackVelX, PlayerKnockbackDecayPerSecond, dt)
	g.Player.KnockbackVelY = decayAxisVelocity(g.Player.KnockbackVelY, PlayerKnockbackDecayPerSecond, dt)

	totalVelX := g.Player.MoveVelocityX + g.Player.KnockbackVelX
	totalVelY := g.Player.MoveVelocityY + g.Player.KnockbackVelY
	s.updatePlayerFacing(g, totalVelX)

	moveDeltaX := totalVelX * dt
	moveDeltaY := totalVelY * dt

	if g.CurrentRoom == nil {
		g.Player.PosX += moveDeltaX
		g.Player.PosY += moveDeltaY
		return
	}

	startX := g.Player.PosX
	startY := g.Player.PosY
	newX, newY := systems.ResolvePlayerMovement(
		startX,
		startY,
		g.Player.Hitbox.Width,
		g.Player.Hitbox.Height,
		moveDeltaX,
		moveDeltaY,
		g.CurrentRoom,
	)

	appliedDeltaX := newX - startX
	appliedDeltaY := newY - startY
	g.Player.MoveVelocityX, g.Player.KnockbackVelX = dampBlockedAxis(g.Player.MoveVelocityX, g.Player.KnockbackVelX, moveDeltaX, appliedDeltaX)
	g.Player.MoveVelocityY, g.Player.KnockbackVelY = dampBlockedAxis(g.Player.MoveVelocityY, g.Player.KnockbackVelY, moveDeltaY, appliedDeltaY)

	g.Player.PosX = newX
	g.Player.PosY = newY
}

func (s *movementSystem) updatePlayerFacing(g *Game, horizontalVelocity float32) {
	if horizontalVelocity > PlayerFacingDeadzone {
		g.Player.FacingRight = true
		return
	}
	if horizontalVelocity < -PlayerFacingDeadzone {
		g.Player.FacingRight = false
		return
	}

	valid, targetX, _, targetWidth, _ := g.getPlayerAttackTargetBounds()
	if !valid {
		return
	}

	playerCenterX, _ := g.Player.Center()
	targetCenterX := targetX + targetWidth/2
	dx := targetCenterX - playerCenterX
	if dx > PlayerFacingDeadzone {
		g.Player.FacingRight = true
	}
	if dx < -PlayerFacingDeadzone {
		g.Player.FacingRight = false
	}
}

type combatResolveSystem struct{}

func (s *combatResolveSystem) Name() string { return "Combat Resolve" }

func (s *combatResolveSystem) Update(ctx *RuntimeContext, _ float32) {
	g := ctx.Game
	if ctx.IsMenuOpen || g.Player == nil {
		return
	}

	for _, enemy := range g.Enemies {
		hit, damage, sourceX, sourceY := enemy.Attack()
		if hit {
			g.ApplyPlayerDirectHit(damage, sourceX, sourceY)
		}
	}

	if g.Boss != nil {
		hit, damage, sourceX, sourceY := g.Boss.Attack()
		if hit {
			g.ApplyPlayerDirectHit(damage, sourceX, sourceY)
		}
	}
}

type effectsSystem struct{}

func (s *effectsSystem) Name() string { return "Effects" }

func (s *effectsSystem) Update(ctx *RuntimeContext, dt float32) {
	g := ctx.Game
	if ctx.IsMenuOpen || g.Player == nil {
		return
	}

	g.RunElapsed += dt
	g.Player.Update(dt)
}

type dungeonRunSystem struct{}

func (s *dungeonRunSystem) Name() string { return "Dungeon/Run" }

func (s *dungeonRunSystem) Update(ctx *RuntimeContext, dt float32) {
	g := ctx.Game
	if g.Player == nil {
		return
	}

	if g.RoomTransitionTimer > 0 {
		g.RoomTransitionTimer -= dt
		if g.RoomTransitionTimer <= 0 {
			g.RoomTransitionTimer = 0
			if g.PendingRoomTransition {
				g.PendingRoomTransition = false
				g.AdvanceToNextRoom()
			}
		}
	}

	if !g.Player.IsAlive() {
		g.EnterResults(false, "")
		return
	}

	if g.CurrentRoom != nil {
		roomCleared := g.CheckRoomCompletion()
		if g.CurrentRoom.IsBoss() && roomCleared {
			g.EnterReward()
			return
		}

		if !g.CurrentRoom.IsBoss() {
			g.CurrentRoom.SetDoorsLocked(!roomCleared)
		}

		if roomCleared && !g.CurrentRoom.IsBoss() && !g.PendingRoomTransition && g.RoomTransitionTimer == 0 {
			playerBounds := world.AABB{
				X:      g.Player.PosX,
				Y:      g.Player.PosY,
				Width:  g.Player.Hitbox.Width,
				Height: g.Player.Hitbox.Height,
			}

			for _, door := range g.CurrentRoom.Doors {
				if door == nil || door.Locked {
					continue
				}
				if !systems.AABBOverlap(playerBounds, door.Bounds) {
					continue
				}

				g.PendingRoomTransition = true
				g.RoomTransitionTimer = g.RoomTransitionDuration
				g.HasPlayerMoveTarget = false
				g.PlayerAttackTarget = nil
				break
			}
		}
	}

	if g.Dungeon != nil {
		worldWidth, worldHeight := g.Dungeon.GetWorldBounds()
		playerCenterX, playerCenterY := g.Player.Center()
		systems.UpdateCamera(g.Camera, playerCenterX, playerCenterY, worldWidth, worldHeight)
	}
}

type uiRenderPrepSystem struct{}

func (s *uiRenderPrepSystem) Name() string { return "UI/Render Prep" }

func (s *uiRenderPrepSystem) Update(ctx *RuntimeContext, _ float32) {
	g := ctx.Game
	if g.Player == nil {
		return
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
}
