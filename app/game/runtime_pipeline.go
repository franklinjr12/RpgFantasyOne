package game

import (
	"math"
	"sort"
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

	worldX := ctx.Input.CursorWorldX
	worldY := ctx.Input.CursorWorldY

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
	if ctx.IsMenuOpen {
		return
	}

	skillInputs := []bool{ctx.Input.Skill1, ctx.Input.Skill2, ctx.Input.Skill3, ctx.Input.Skill4}
	for i, skillPressed := range skillInputs {
		if !skillPressed || i >= len(g.Player.Skills) {
			continue
		}
		skill := g.Player.Skills[i]
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
	s.updateDelayedSkillEffects(g, dt)
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

func (s *projectilesSystem) updatePlayerProjectiles(g *Game, dt float32) {
	for i := len(g.Projectiles) - 1; i >= 0; i-- {
		proj := g.Projectiles[i]
		if !proj.Alive {
			g.Projectiles = append(g.Projectiles[:i], g.Projectiles[i+1:]...)
			continue
		}

		proj.Lifetime -= dt
		if proj.Lifetime <= 0 {
			proj.Alive = false
			g.Projectiles = append(g.Projectiles[:i], g.Projectiles[i+1:]...)
			continue
		}

		proj.X += proj.VX * dt
		proj.Y += proj.VY * dt

		s.resolvePlayerProjectileHits(g, proj)

		if g.CurrentRoom != nil {
			if proj.X < g.CurrentRoom.X || proj.X > g.CurrentRoom.X+g.CurrentRoom.Width ||
				proj.Y < g.CurrentRoom.Y || proj.Y > g.CurrentRoom.Y+g.CurrentRoom.Height {
				proj.Alive = false
			}
		}

		if !proj.Alive {
			g.Projectiles = append(g.Projectiles[:i], g.Projectiles[i+1:]...)
		}
	}
}

func (s *projectilesSystem) resolvePlayerProjectileHits(g *Game, proj *Projectile) {
	for _, enemy := range g.Enemies {
		if enemy == nil || !enemy.IsAlive() || !proj.Alive {
			continue
		}
		if s.tryHitEnemy(g, proj, enemy) {
			return
		}
	}

	if g.Boss == nil || !g.Boss.IsAlive() || !proj.Alive {
		return
	}
	s.tryHitBoss(g, proj, g.Boss)
}

func (s *projectilesSystem) tryHitEnemy(g *Game, proj *Projectile, enemy *gameobjects.Enemy) bool {
	if wasTargetHitByProjectile(proj, enemy) {
		return false
	}

	enemyX, enemyY := enemy.Center()
	distance := systems.GetDistance(proj.X, proj.Y, enemyX, enemyY)
	if distance > proj.Radius+enemy.Hitbox.Width/2 {
		return false
	}

	wasAlive := enemy.IsAlive()
	s.applyProjectileHit(proj, enemy)
	markProjectileTargetHit(proj, enemy)
	g.spawnSkillImpactVisual(proj.Skill, enemyX, enemyY)
	g.playSkillImpactSFX(proj.Skill)
	if wasAlive && !enemy.IsAlive() {
		g.Player.GainXP(20)
		if g.Player.Class.Type == gamedata.ClassTypeRanged {
			g.Player.Heal(g.Player.Class.KillHealAmount)
		}
	}

	if proj.Pierce > 0 {
		proj.Pierce--
		return false
	}
	proj.Alive = false
	return true
}

func (s *projectilesSystem) tryHitBoss(g *Game, proj *Projectile, boss *gameobjects.Boss) bool {
	if wasTargetHitByProjectile(proj, boss) {
		return false
	}

	bossX, bossY := boss.Center()
	distance := systems.GetDistance(proj.X, proj.Y, bossX, bossY)
	if distance > proj.Radius+boss.Hitbox.Width/2 {
		return false
	}

	wasAlive := boss.IsAlive()
	s.applyProjectileHit(proj, boss)
	markProjectileTargetHit(proj, boss)
	g.spawnSkillImpactVisual(proj.Skill, bossX, bossY)
	g.playSkillImpactSFX(proj.Skill)
	if wasAlive && !boss.IsAlive() {
		g.Player.GainXP(100)
		if g.Player.Class.Type == gamedata.ClassTypeRanged {
			g.Player.Heal(g.Player.Class.KillHealAmount * 5)
		}
	}

	if proj.Pierce > 0 {
		proj.Pierce--
		return false
	}
	proj.Alive = false
	return true
}

func (s *projectilesSystem) applyProjectileHit(proj *Projectile, target interface{}) {
	if proj.Skill != nil && proj.Caster != nil {
		systems.ApplySkill(proj.Caster, proj.Skill, []interface{}{target})
		return
	}

	systems.ApplyCombatHit(systems.CombatHitRequest{
		Caster:             proj.Caster,
		Target:             target,
		BaseDamage:         proj.Damage,
		DamageType:         proj.DamageType,
		CritMultiplier:     1.5,
		ApplyOnHitHooks:    proj.Caster != nil,
		UseSourceModifiers: false,
	})
}

func (s *projectilesSystem) updateDelayedSkillEffects(g *Game, dt float32) {
	for i := len(g.DelayedSkillEffects) - 1; i >= 0; i-- {
		delayed := g.DelayedSkillEffects[i]
		if delayed == nil || !delayed.Alive {
			g.DelayedSkillEffects = append(g.DelayedSkillEffects[:i], g.DelayedSkillEffects[i+1:]...)
			continue
		}

		if !delayed.Active {
			delayed.Delay -= dt
			if delayed.Delay > 0 {
				continue
			}
			delayed.Active = true

			targetsHit := s.applyDelayedSkill(g, delayed)
			if targetsHit > 0 {
				g.spawnSkillImpactVisual(delayed.Skill, delayed.LastAppliedX, delayed.LastAppliedY)
				g.playSkillImpactSFX(delayed.Skill)
			}

			if delayed.ActiveTime <= 0 || delayed.TickRate <= 0 {
				delayed.Alive = false
				g.DelayedSkillEffects = append(g.DelayedSkillEffects[:i], g.DelayedSkillEffects[i+1:]...)
				continue
			}
		}

		if delayed.ActiveTime <= 0 {
			delayed.Alive = false
			g.DelayedSkillEffects = append(g.DelayedSkillEffects[:i], g.DelayedSkillEffects[i+1:]...)
			continue
		}

		delayed.ActiveTime -= dt
		delayed.TickTimer += dt
		for delayed.TickTimer >= delayed.TickRate && delayed.ActiveTime > 0 {
			delayed.TickTimer -= delayed.TickRate
			targetsHit := s.applyDelayedSkill(g, delayed)
			if targetsHit > 0 {
				g.spawnSkillImpactVisual(delayed.Skill, delayed.LastAppliedX, delayed.LastAppliedY)
				g.playSkillImpactSFX(delayed.Skill)
			}
		}

		if delayed.ActiveTime <= 0 {
			delayed.Alive = false
			g.DelayedSkillEffects = append(g.DelayedSkillEffects[:i], g.DelayedSkillEffects[i+1:]...)
		}
	}
}

func (s *projectilesSystem) applyDelayedSkill(g *Game, delayed *DelayedSkillEffect) int {
	if g == nil || delayed == nil || delayed.Skill == nil || delayed.Caster == nil {
		return 0
	}
	targets := s.resolveDelayedTargets(g, delayed)
	systems.ApplySkill(delayed.Caster, delayed.Skill, targets)
	g.applySkillPostCast(delayed.Skill, len(targets))
	delayed.LastAppliedX = delayed.X
	delayed.LastAppliedY = delayed.Y
	return len(targets)
}

func (s *projectilesSystem) resolveDelayedTargets(g *Game, delayed *DelayedSkillEffect) []interface{} {
	if delayed.Skill.Targeting.Type != gamedata.TargetArea {
		return systems.ResolveTargets(delayed.Caster, delayed.Intent, delayed.Skill.Targeting, g.Enemies, g.Boss)
	}

	radius := delayed.Skill.Targeting.Radius
	if radius <= 0 {
		radius = delayed.Radius
	}
	if radius <= 0 {
		return nil
	}

	type candidate struct {
		target    interface{}
		distance2 float32
		order     int
	}

	centerX := delayed.X
	centerY := delayed.Y
	radius2 := radius * radius
	candidates := make([]candidate, 0, len(g.Enemies)+1)
	order := 0

	for _, enemy := range g.Enemies {
		if enemy == nil || !enemy.IsAlive() {
			order++
			continue
		}
		targetX, targetY := enemy.Center()
		dx := targetX - centerX
		dy := targetY - centerY
		distance2 := dx*dx + dy*dy
		if distance2 <= radius2 {
			candidates = append(candidates, candidate{
				target:    enemy,
				distance2: distance2,
				order:     order,
			})
		}
		order++
	}

	if g.Boss != nil && g.Boss.IsAlive() {
		targetX, targetY := g.Boss.Center()
		dx := targetX - centerX
		dy := targetY - centerY
		distance2 := dx*dx + dy*dy
		if distance2 <= radius2 {
			candidates = append(candidates, candidate{
				target:    g.Boss,
				distance2: distance2,
				order:     order,
			})
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].distance2 == candidates[j].distance2 {
			return candidates[i].order < candidates[j].order
		}
		return candidates[i].distance2 < candidates[j].distance2
	})

	limit := len(candidates)
	maxTargets := delayed.Skill.Targeting.MaxTargets
	if maxTargets > 0 && maxTargets < limit {
		limit = maxTargets
	}

	targets := make([]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		targets = append(targets, candidates[i].target)
	}
	return targets
}

func markProjectileTargetHit(proj *Projectile, target interface{}) {
	if proj.HitTargets == nil {
		proj.HitTargets = map[interface{}]struct{}{}
	}
	proj.HitTargets[target] = struct{}{}
}

func wasTargetHitByProjectile(proj *Projectile, target interface{}) bool {
	if proj.HitTargets == nil {
		return false
	}
	_, hit := proj.HitTargets[target]
	return hit
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
	if g == nil {
		return
	}

	g.updateSkillVisualEffects(dt)
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
