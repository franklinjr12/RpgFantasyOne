package game

import (
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
		if enemy == nil {
			continue
		}
		enemy.Update(dt)
		gameobjects.ResolveEnemyIntent(enemy, playerX, playerY)
	}

	if g.Boss == nil {
		return
	}

	g.Boss.Update(dt, playerX, playerY)
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
	s.updateEnemyProjectiles(g, dt)
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

func (s *projectilesSystem) updateEnemyProjectiles(g *Game, dt float32) {
	for i := len(g.EnemyProjectiles) - 1; i >= 0; i-- {
		proj := g.EnemyProjectiles[i]
		if proj == nil || !proj.Alive {
			g.EnemyProjectiles = append(g.EnemyProjectiles[:i], g.EnemyProjectiles[i+1:]...)
			continue
		}

		proj.Lifetime -= dt
		if proj.Lifetime <= 0 {
			proj.Alive = false
			g.EnemyProjectiles = append(g.EnemyProjectiles[:i], g.EnemyProjectiles[i+1:]...)
			continue
		}

		proj.X += proj.VX * dt
		proj.Y += proj.VY * dt

		playerCenterX, playerCenterY := g.Player.Center()
		distance := systems.GetDistance(proj.X, proj.Y, playerCenterX, playerCenterY)
		if distance <= proj.Radius+g.Player.Hitbox.Width/2 {
			g.ApplyPlayerCombatHit(proj.Damage, proj.DamageType, proj.X, proj.Y, proj.Effects)
			proj.Alive = false
		}

		if g.CurrentRoom != nil {
			if proj.X < g.CurrentRoom.X || proj.X > g.CurrentRoom.X+g.CurrentRoom.Width ||
				proj.Y < g.CurrentRoom.Y || proj.Y > g.CurrentRoom.Y+g.CurrentRoom.Height {
				proj.Alive = false
			}
		}

		if !proj.Alive {
			g.EnemyProjectiles = append(g.EnemyProjectiles[:i], g.EnemyProjectiles[i+1:]...)
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
	s.applyProjectileHit(g, proj, enemy)
	markProjectileTargetHit(proj, enemy)
	g.spawnSkillImpactVisual(proj.Skill, enemyX, enemyY)
	if wasAlive && !enemy.IsAlive() {
		reward := enemy.XPReward
		if reward <= 0 {
			reward = 20
		}
		g.grantPlayerXP(reward)
		if g.Player.Class.Type == gamedata.ClassTypeRanged {
			g.healPlayerWithFeedback(g.Player.Class.KillHealAmount)
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
	s.applyProjectileHit(g, proj, boss)
	markProjectileTargetHit(proj, boss)
	g.spawnSkillImpactVisual(proj.Skill, bossX, bossY)
	if wasAlive && !boss.IsAlive() {
		g.grantPlayerXP(100)
		if g.Player.Class.Type == gamedata.ClassTypeRanged {
			g.healPlayerWithFeedback(g.Player.Class.KillHealAmount * 5)
		}
	}

	if proj.Pierce > 0 {
		proj.Pierce--
		return false
	}
	proj.Alive = false
	return true
}

func (s *projectilesSystem) applyProjectileHit(g *Game, proj *Projectile, target interface{}) {
	if g == nil || proj == nil || target == nil {
		return
	}
	if proj.Skill != nil && proj.Caster != nil {
		g.applySkillWithFeedback(proj.Caster, proj.Skill, []interface{}{target})
		return
	}

	g.applyCombatHitWithFeedback(systems.CombatHitRequest{
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
	g.applySkillWithFeedback(delayed.Caster, delayed.Skill, targets)
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

	s.updateEnemies(g, dt)
	s.updateBoss(g, dt)
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

func (s *movementSystem) updateEnemies(g *Game, dt float32) {
	for _, enemy := range g.Enemies {
		if enemy == nil || !enemy.IsAlive() {
			continue
		}
		if !gamedata.CanAct(&enemy.Effects) {
			continue
		}

		speed := enemy.MoveSpeed * gamedata.MoveSpeedMultiplier(&enemy.Effects)
		if speed <= 0 {
			continue
		}

		moveDeltaX := enemy.IntentMoveX * speed * dt
		moveDeltaY := enemy.IntentMoveY * speed * dt
		if moveDeltaX == 0 && moveDeltaY == 0 {
			continue
		}

		nextX := enemy.PosX + moveDeltaX
		nextY := enemy.PosY + moveDeltaY
		if g.CurrentRoom != nil {
			minX := g.CurrentRoom.X
			minY := g.CurrentRoom.Y
			maxX := g.CurrentRoom.X + g.CurrentRoom.Width - enemy.Hitbox.Width
			maxY := g.CurrentRoom.Y + g.CurrentRoom.Height - enemy.Hitbox.Height
			if nextX < minX {
				nextX = minX
			}
			if nextX > maxX {
				nextX = maxX
			}
			if nextY < minY {
				nextY = minY
			}
			if nextY > maxY {
				nextY = maxY
			}

			candidate := world.AABB{X: nextX, Y: nextY, Width: enemy.Hitbox.Width, Height: enemy.Hitbox.Height}
			if overlapsRoomObstacle(candidate, g.CurrentRoom.Obstacles) {
				nextX = enemy.PosX
				nextY = enemy.PosY
			}
		}

		enemy.PosX = nextX
		enemy.PosY = nextY
	}
}

func (s *movementSystem) updateBoss(g *Game, dt float32) {
	if g == nil || g.Boss == nil || !g.Boss.IsAlive() {
		return
	}
	if !gamedata.CanAct(&g.Boss.Effects) {
		return
	}

	speed := g.Boss.MoveSpeed * gamedata.MoveSpeedMultiplier(&g.Boss.Effects)
	if speed <= 0 {
		return
	}

	moveDeltaX := g.Boss.IntentMoveX * speed * dt
	moveDeltaY := g.Boss.IntentMoveY * speed * dt
	if moveDeltaX == 0 && moveDeltaY == 0 {
		return
	}

	nextX := g.Boss.PosX + moveDeltaX
	nextY := g.Boss.PosY + moveDeltaY
	if g.CurrentRoom != nil {
		minX := g.CurrentRoom.X
		minY := g.CurrentRoom.Y
		maxX := g.CurrentRoom.X + g.CurrentRoom.Width - g.Boss.Hitbox.Width
		maxY := g.CurrentRoom.Y + g.CurrentRoom.Height - g.Boss.Hitbox.Height
		if nextX < minX {
			nextX = minX
		}
		if nextX > maxX {
			nextX = maxX
		}
		if nextY < minY {
			nextY = minY
		}
		if nextY > maxY {
			nextY = maxY
		}

		current := world.AABB{X: g.Boss.PosX, Y: g.Boss.PosY, Width: g.Boss.Hitbox.Width, Height: g.Boss.Hitbox.Height}
		candidate := world.AABB{X: nextX, Y: nextY, Width: g.Boss.Hitbox.Width, Height: g.Boss.Hitbox.Height}
		// If the boss is already intersecting an obstacle, allow movement so it can
		// escape instead of being permanently pinned.
		if overlapsRoomObstacle(candidate, g.CurrentRoom.Obstacles) && !overlapsRoomObstacle(current, g.CurrentRoom.Obstacles) {
			nextX = g.Boss.PosX
			nextY = g.Boss.PosY
		}
	}

	g.Boss.PosX = nextX
	g.Boss.PosY = nextY
}

func overlapsRoomObstacle(candidate world.AABB, obstacles []world.AABB) bool {
	for _, obstacle := range obstacles {
		if systems.AABBOverlap(candidate, obstacle) {
			return true
		}
	}
	return false
}

type combatResolveSystem struct{}

func (s *combatResolveSystem) Name() string { return "Combat Resolve" }

func (s *combatResolveSystem) Update(ctx *RuntimeContext, _ float32) {
	g := ctx.Game
	if ctx.IsMenuOpen || g.Player == nil {
		return
	}

	playerX, playerY := g.Player.Center()
	for _, enemy := range g.Enemies {
		if enemy == nil {
			continue
		}
		hit, payload := enemy.Attack(playerX, playerY)
		if !hit {
			continue
		}

		if payload.AttackMode == gamedata.EnemyAttackProjectile {
			dx := playerX - payload.SourceX
			dy := playerY - payload.SourceY
			distance := systems.GetDistance(0, 0, dx, dy)
			if distance <= 0 {
				distance = 1
			}
			lifetime := payload.ProjectileLifetime
			if lifetime <= 0 {
				lifetime = 2.0
			}
			radius := payload.ProjectileRadius
			if radius <= 0 {
				radius = 6
			}
			speed := payload.ProjectileSpeed
			if speed <= 0 {
				speed = 220
			}
			g.EnemyProjectiles = append(g.EnemyProjectiles, &EnemyProjectile{
				X:          payload.SourceX,
				Y:          payload.SourceY,
				VX:         (dx / distance) * speed,
				VY:         (dy / distance) * speed,
				Speed:      speed,
				Damage:     payload.Damage,
				Radius:     radius,
				Lifetime:   lifetime,
				Alive:      true,
				DamageType: payload.DamageType,
				Effects:    payload.OnHitEffects,
			})
			g.playSound(sfxEnemyCast)
			continue
		}

		g.ApplyPlayerCombatHit(payload.Damage, payload.DamageType, payload.SourceX, payload.SourceY, payload.OnHitEffects)
	}

	if g.Boss != nil && g.Boss.IsAlive() {
		hit, damage, sourceX, sourceY := g.Boss.Attack(playerX, playerY)
		if hit {
			g.ApplyPlayerCombatHit(damage, gamedata.DamagePhysical, sourceX, sourceY, nil)
		}

		playerCenterX, playerCenterY := g.Player.Center()
		for _, event := range g.Boss.ConsumeDamageEvents() {
			if !isPlayerWithinBossEvent(playerCenterX, playerCenterY, g.Player.Hitbox.Width, event) {
				continue
			}
			g.ApplyPlayerCombatHit(event.Damage, event.DamageType, event.X, event.Y, event.Effects)
		}
	}
}

func isPlayerWithinBossEvent(playerX, playerY, playerWidth float32, event gameobjects.BossDamageEvent) bool {
	radius := event.Radius
	if radius <= 0 {
		return false
	}
	distance := systems.GetDistance(playerX, playerY, event.X, event.Y)
	return distance <= radius+playerWidth/2
}

type effectsSystem struct{}

func (s *effectsSystem) Name() string { return "Effects" }

func (s *effectsSystem) Update(ctx *RuntimeContext, dt float32) {
	g := ctx.Game
	if g == nil {
		return
	}

	g.updateSoundCooldowns(dt)
	g.updateSkillVisualEffects(dt)
	g.updateCombatFeedback(dt)
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
		if g.CurrentRoom.Type == world.RoomTypeEvent && !g.CurrentRoom.Completed {
			g.CurrentRoom.EventTimeLeft -= dt
			if g.CurrentRoom.EventTimeLeft < 0 {
				g.CurrentRoom.EventTimeLeft = 0
			}
		}

		roomCleared := g.CheckRoomCompletion()
		if g.CurrentRoom.IsBoss() && roomCleared {
			g.EnterBossReward()
			return
		}
		if roomCleared && !g.CurrentRoom.IsBoss() && g.shouldTriggerMilestoneReward() {
			g.EnterMilestoneReward()
			return
		}

		if !g.CurrentRoom.IsBoss() {
			hadLockedDoor := roomHasLockedDoor(g.CurrentRoom)
			g.CurrentRoom.SetDoorsLocked(!roomCleared)
			if roomCleared && hadLockedDoor && roomHasUnlockedDoor(g.CurrentRoom) {
				g.playSound(sfxDoorOpen)
			}
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

func roomHasLockedDoor(room *world.Room) bool {
	if room == nil {
		return false
	}
	for _, door := range room.Doors {
		if door != nil && door.Locked {
			return true
		}
	}
	return false
}

func roomHasUnlockedDoor(room *world.Room) bool {
	if room == nil {
		return false
	}
	for _, door := range room.Doors {
		if door != nil && !door.Locked {
			return true
		}
	}
	return false
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
