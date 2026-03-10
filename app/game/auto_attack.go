package game

import (
	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/systems"
)

type autoAttackTiming struct {
	Windup  float32
	Recover float32
}

func (g *Game) ApplyPlayerDirectHit(damage int, sourceX, sourceY float32) bool {
	return g.ApplyPlayerCombatHit(damage, gamedata.DamagePhysical, sourceX, sourceY, nil)
}

func (g *Game) ApplyPlayerCombatHit(damage int, damageType gamedata.DamageType, sourceX, sourceY float32, effects []gamedata.EffectSpec) bool {
	if g.Player == nil || !g.Player.IsAlive() || damage <= 0 {
		return false
	}
	if !g.Player.CanTakeDirectHit() {
		return false
	}

	systems.ApplyCombatHit(systems.CombatHitRequest{
		Target:        g.Player,
		BaseDamage:    damage,
		DamageType:    damageType,
		Effects:       effects,
		SuppressFlash: true,
	})
	g.Player.HitFlashTimer = PlayerHitFlashDuration
	g.Player.StartHurtIFrames(PlayerHurtIFrameDuration)
	g.Player.ApplyKnockbackFrom(sourceX, sourceY, PlayerKnockbackImpulse)
	return true
}

func (g *Game) UpdateAutoAttack(dt float32) {
	if g.Player == nil {
		return
	}
	if !gamedata.CanAct(&g.Player.Effects) {
		g.Player.AttackState = gameobjects.PlayerAttackStateIdle
		g.Player.AttackStateTimer = 0
		return
	}

	g.advanceAutoAttackState(dt)
	if g.Player.AttackState != gameobjects.PlayerAttackStateIdle {
		return
	}
	if g.Player.CurrentAttackCooldown > 0 {
		return
	}

	valid, targetX, targetY, targetWidth, targetHeight := g.getPlayerAttackTargetBounds()
	if !valid {
		return
	}

	playerCenterX, playerCenterY := g.Player.Center()
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
	if g.Player.Class.Type == gamedata.ClassTypeCaster && !g.Player.CanUseMana(g.Player.Class.ManaCost) {
		return
	}

	timing := getAutoAttackTiming(g.Player.Class.Type)
	g.Player.AttackState = gameobjects.PlayerAttackStateWindup
	g.Player.AttackStateTimer = timing.Windup
}

func (g *Game) advanceAutoAttackState(dt float32) {
	if g.Player == nil || dt <= 0 {
		return
	}
	if !gamedata.CanAct(&g.Player.Effects) {
		g.Player.AttackState = gameobjects.PlayerAttackStateIdle
		g.Player.AttackStateTimer = 0
		return
	}

	switch g.Player.AttackState {
	case gameobjects.PlayerAttackStateWindup:
		g.Player.AttackStateTimer -= dt
		if g.Player.AttackStateTimer > 0 {
			return
		}

		g.resolveAutoAttackHit()
		g.Player.CurrentAttackCooldown = g.Player.GetAttackCooldown()
		g.Player.AttackState = gameobjects.PlayerAttackStateRecover
		g.Player.AttackStateTimer = getAutoAttackTiming(g.Player.Class.Type).Recover

	case gameobjects.PlayerAttackStateRecover:
		g.Player.AttackStateTimer -= dt
		if g.Player.AttackStateTimer > 0 {
			return
		}
		g.Player.AttackState = gameobjects.PlayerAttackStateIdle
		g.Player.AttackStateTimer = 0
	}
}

func (g *Game) resolveAutoAttackHit() {
	if g.Player == nil {
		return
	}

	damage := g.Player.GetAutoAttackDamage()
	switch g.Player.Class.Type {
	case gamedata.ClassTypeMelee:
		g.resolveMeleeAutoAttack(damage)
	case gamedata.ClassTypeRanged:
		g.resolveRangedAutoAttack(damage)
	case gamedata.ClassTypeCaster:
		g.resolveCasterAutoAttack(damage)
	}
}

func (g *Game) resolveMeleeAutoAttack(damage int) {
	valid, targetX, targetY, targetWidth, targetHeight := g.getPlayerAttackTargetBounds()
	if !valid {
		return
	}

	playerCenterX, playerCenterY := g.Player.Center()
	targetCenterX := targetX + targetWidth/2
	targetCenterY := targetY + targetHeight/2
	distance := systems.GetDistance(playerCenterX, playerCenterY, targetCenterX, targetCenterY)
	if distance > g.Player.AttackRange+MeleeAttackHitRangeBuffer {
		return
	}

	switch t := g.PlayerAttackTarget.(type) {
	case *gameobjects.Enemy:
		wasAlive := t.IsAlive()
		systems.ApplyCombatHit(systems.CombatHitRequest{
			Caster:             g.Player,
			Target:             t,
			BaseDamage:         damage,
			DamageType:         gamedata.DamagePhysical,
			CritMultiplier:     1.5,
			ApplyOnHitHooks:    true,
			UseSourceModifiers: false,
		})
		if wasAlive && !t.IsAlive() {
			g.Player.GainXP(20)
			g.PlayerAttackTarget = nil
		}
	case *gameobjects.Boss:
		wasAlive := t.IsAlive()
		systems.ApplyCombatHit(systems.CombatHitRequest{
			Caster:             g.Player,
			Target:             t,
			BaseDamage:         damage,
			DamageType:         gamedata.DamagePhysical,
			CritMultiplier:     1.5,
			ApplyOnHitHooks:    true,
			UseSourceModifiers: false,
		})
		if wasAlive && !t.IsAlive() {
			g.Player.GainXP(100)
			g.PlayerAttackTarget = nil
		}
	}
}

func (g *Game) resolveRangedAutoAttack(damage int) {
	valid, targetX, targetY, targetWidth, targetHeight := g.getPlayerAttackTargetBounds()
	if !valid {
		return
	}

	playerCenterX, playerCenterY := g.Player.Center()
	targetCenterX := targetX + targetWidth/2
	targetCenterY := targetY + targetHeight/2

	dx := targetCenterX - playerCenterX
	dy := targetCenterY - playerCenterY
	distance := systems.GetDistance(0, 0, dx, dy)
	if distance <= 0 {
		return
	}

	proj := &Projectile{
		X:          playerCenterX,
		Y:          playerCenterY,
		VX:         (dx / distance) * AutoAttackProjectileSpeed,
		VY:         (dy / distance) * AutoAttackProjectileSpeed,
		Speed:      AutoAttackProjectileSpeed,
		Damage:     damage,
		Radius:     5,
		Lifetime:   2.0,
		Pierce:     0,
		HitTargets: map[interface{}]struct{}{},
		Alive:      true,
		Caster:     g.Player,
		DamageType: gamedata.DamagePhysical,
	}
	g.Projectiles = append(g.Projectiles, proj)
}

func (g *Game) resolveCasterAutoAttack(damage int) {
	if !g.Player.CanUseMana(g.Player.Class.ManaCost) {
		return
	}

	valid, targetX, targetY, targetWidth, targetHeight := g.getPlayerAttackTargetBounds()
	if !valid {
		return
	}

	playerCenterX, playerCenterY := g.Player.Center()
	targetCenterX := targetX + targetWidth/2
	targetCenterY := targetY + targetHeight/2
	distance := systems.GetDistance(playerCenterX, playerCenterY, targetCenterX, targetCenterY)
	if distance > g.Player.AttackRange+MeleeAttackHitRangeBuffer {
		return
	}

	switch t := g.PlayerAttackTarget.(type) {
	case *gameobjects.Enemy:
		wasAlive := t.IsAlive()
		systems.ApplyCombatHit(systems.CombatHitRequest{
			Caster:             g.Player,
			Target:             t,
			BaseDamage:         damage,
			DamageType:         gamedata.DamageMagical,
			CritMultiplier:     1.5,
			ApplyOnHitHooks:    true,
			UseSourceModifiers: false,
		})
		g.Player.UseMana(g.Player.Class.ManaCost)
		if wasAlive && !t.IsAlive() {
			g.Player.GainXP(20)
			g.PlayerAttackTarget = nil
		}
	case *gameobjects.Boss:
		wasAlive := t.IsAlive()
		systems.ApplyCombatHit(systems.CombatHitRequest{
			Caster:             g.Player,
			Target:             t,
			BaseDamage:         damage,
			DamageType:         gamedata.DamageMagical,
			CritMultiplier:     1.5,
			ApplyOnHitHooks:    true,
			UseSourceModifiers: false,
		})
		g.Player.UseMana(g.Player.Class.ManaCost)
		if wasAlive && !t.IsAlive() {
			g.Player.GainXP(100)
			g.PlayerAttackTarget = nil
		}
	}
}

func getAutoAttackTiming(classType gamedata.ClassType) autoAttackTiming {
	switch classType {
	case gamedata.ClassTypeMelee:
		return autoAttackTiming{Windup: MeleeAttackWindup, Recover: MeleeAttackRecover}
	case gamedata.ClassTypeRanged:
		return autoAttackTiming{Windup: RangedAttackWindup, Recover: RangedAttackRecover}
	case gamedata.ClassTypeCaster:
		return autoAttackTiming{Windup: CasterAttackWindup, Recover: CasterAttackRecover}
	default:
		return autoAttackTiming{Windup: 0, Recover: 0}
	}
}

func (g *Game) getPlayerAttackTargetBounds() (bool, float32, float32, float32, float32) {
	if g.Player == nil || g.PlayerAttackTarget == nil {
		return false, 0, 0, 0, 0
	}

	switch t := g.PlayerAttackTarget.(type) {
	case *gameobjects.Enemy:
		if !t.IsAlive() {
			g.PlayerAttackTarget = nil
			return false, 0, 0, 0, 0
		}
		return true, t.PosX, t.PosY, t.Hitbox.Width, t.Hitbox.Height
	case *gameobjects.Boss:
		if t.Enemy == nil || !t.IsAlive() {
			g.PlayerAttackTarget = nil
			return false, 0, 0, 0, 0
		}
		return true, t.PosX, t.PosY, t.Hitbox.Width, t.Hitbox.Height
	default:
		g.PlayerAttackTarget = nil
		return false, 0, 0, 0, 0
	}
}
