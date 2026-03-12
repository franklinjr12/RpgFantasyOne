package game

import (
	"singlefantasy/app/gamedata"
	"singlefantasy/app/systems"
)

func (g *Game) TryCastSkill(skill *gamedata.Skill, input *systems.Input) {
	if g == nil || g.Player == nil || skill == nil {
		return
	}
	if !systems.CanCast(g.Player, skill) {
		return
	}

	intent := g.buildCastIntent(input)
	if !g.executeSkillDelivery(skill, intent) {
		return
	}
	g.spawnSkillCastVisual(skill, intent)
	g.playSkillCastSFX(skill)
	skill.Use()
}

func (g *Game) buildCastIntent(input *systems.Input) systems.CastIntent {
	if input != nil {
		return systems.BuildCastIntent(g.Player, input.CursorWorldX, input.CursorWorldY)
	}
	cx, cy := g.Player.Center()
	return systems.BuildCastIntent(g.Player, cx, cy)
}

func (g *Game) executeSkillDelivery(skill *gamedata.Skill, intent systems.CastIntent) bool {
	g.Player.UseMana(skill.ManaCost)

	switch skill.Delivery.Type {
	case gamedata.DeliveryInstant:
		g.resolveAndApplySkill(skill, intent)
		return true
	case gamedata.DeliveryProjectile:
		return g.spawnSkillProjectile(skill, intent)
	case gamedata.DeliveryDelayed:
		return g.queueDelayedSkill(skill, intent)
	default:
		return false
	}
}

func (g *Game) resolveAndApplySkill(skill *gamedata.Skill, intent systems.CastIntent) int {
	g.applySkillPreCast(skill, intent)
	targets := systems.ResolveTargets(g.Player, intent, skill.Targeting, g.Enemies, g.Boss)
	g.applySkillWithFeedback(g.Player, skill, targets)
	g.applySkillPostCast(skill, len(targets))
	if len(targets) > 0 {
		impactX, impactY := resolveImpactCenter(intent, targets[0])
		g.spawnSkillImpactVisual(skill, impactX, impactY)
		g.playSkillImpactSFX(skill)
	}
	return len(targets)
}

func (g *Game) applySkillPreCast(skill *gamedata.Skill, intent systems.CastIntent) {
	g.applySelfMovement(skill, intent)
	g.applyManaShield(skill)
}

func (g *Game) applySkillPostCast(skill *gamedata.Skill, targetsHit int) {
	if g == nil || g.Player == nil || skill == nil || targetsHit <= 0 {
		return
	}

	manaRestore := targetsHit * skill.ResourceGain.ManaPerTarget
	if manaRestore > 0 {
		g.Player.GainMana(manaRestore)
	}
}

func (g *Game) applySelfMovement(skill *gamedata.Skill, intent systems.CastIntent) {
	if g == nil || g.Player == nil || skill == nil {
		return
	}
	if skill.SelfMovement.Mode != gamedata.SelfMovementBackwardFromCursor {
		return
	}
	g.applyRetreatRoll(intent, skill.SelfMovement.Distance)
}

func (g *Game) applyManaShield(skill *gamedata.Skill) {
	if g == nil || g.Player == nil || skill == nil {
		return
	}
	ratio := skill.ManaShield.AbsorbFromCurrentManaRatio
	if ratio <= 0 {
		return
	}

	absorbAmount := int(float32(g.Player.Mana) * ratio)
	if absorbAmount <= 0 {
		return
	}
	g.Player.SetManaShield(absorbAmount, skill.ManaShield.Duration)
}

func (g *Game) applyRetreatRoll(intent systems.CastIntent, rollDistance float32) {
	playerCenterX, playerCenterY := g.Player.Center()
	dx := playerCenterX - intent.CursorX
	dy := playerCenterY - intent.CursorY
	distance := systems.GetDistance(0, 0, dx, dy)
	if distance <= 0 {
		return
	}
	if rollDistance <= 0 {
		return
	}

	moveX := (dx / distance) * rollDistance
	moveY := (dy / distance) * rollDistance
	nextX, nextY := systems.ResolvePlayerMovement(
		g.Player.PosX,
		g.Player.PosY,
		g.Player.Hitbox.Width,
		g.Player.Hitbox.Height,
		moveX,
		moveY,
		g.CurrentRoom,
	)
	g.Player.PosX = nextX
	g.Player.PosY = nextY
}

func (g *Game) spawnSkillProjectile(skill *gamedata.Skill, intent systems.CastIntent) bool {
	playerCenterX, playerCenterY := g.Player.Center()
	speed := skill.Delivery.Speed
	if speed <= 0 {
		return false
	}

	lifetime := skill.Delivery.Lifetime
	if lifetime <= 0 {
		lifetime = 2.0
	}
	radius := skill.Delivery.ProjectileRadius
	if radius <= 0 {
		radius = 5
	}

	proj := &Projectile{
		X:          playerCenterX,
		Y:          playerCenterY,
		VX:         intent.DirectionX * speed,
		VY:         intent.DirectionY * speed,
		Speed:      speed,
		Damage:     int(systems.ComputeDamage(skill.DamageSpec, g.Player.GetEffectiveStats())),
		Radius:     radius,
		Lifetime:   lifetime,
		Pierce:     skill.Delivery.Pierce,
		HitTargets: map[interface{}]struct{}{},
		Alive:      true,
		Skill:      skill,
		Caster:     g.Player,
	}
	if skill.DamageSpec != nil {
		proj.DamageType = skill.DamageSpec.DamageType
	}
	g.Projectiles = append(g.Projectiles, proj)
	return true
}

func (g *Game) queueDelayedSkill(skill *gamedata.Skill, intent systems.CastIntent) bool {
	delay := skill.Delivery.Delay
	if delay <= 0 {
		delay = 0.1
	}

	centerX, centerY := g.resolveDelayedCenter(skill, intent)
	updatedIntent := intent
	updatedIntent.CursorX = centerX
	updatedIntent.CursorY = centerY

	radius := skill.Targeting.Radius
	if radius <= 0 {
		radius = 40
	}

	delayed := &DelayedSkillEffect{
		X:            centerX,
		Y:            centerY,
		Radius:       radius,
		Delay:        delay,
		ActiveTime:   skill.Delivery.ZoneDuration,
		TickRate:     skill.Delivery.ZoneTickRate,
		Active:       false,
		Alive:        true,
		Skill:        skill,
		Caster:       g.Player,
		Intent:       updatedIntent,
		LastAppliedX: centerX,
		LastAppliedY: centerY,
	}
	g.DelayedSkillEffects = append(g.DelayedSkillEffects, delayed)
	return true
}

func (g *Game) resolveDelayedCenter(skill *gamedata.Skill, intent systems.CastIntent) (float32, float32) {
	casterX, casterY := g.Player.Center()
	centerX := intent.CursorX
	centerY := intent.CursorY

	if skill.Targeting.Type == gamedata.TargetArea && skill.Targeting.Range == 0 {
		centerX = casterX
		centerY = casterY
	}
	if skill.Targeting.Range <= 0 {
		return centerX, centerY
	}

	dx := centerX - casterX
	dy := centerY - casterY
	distance := systems.GetDistance(0, 0, dx, dy)
	if distance <= skill.Targeting.Range || distance <= 0 {
		return centerX, centerY
	}

	ratio := skill.Targeting.Range / distance
	return casterX + dx*ratio, casterY + dy*ratio
}

func resolveImpactCenter(intent systems.CastIntent, target interface{}) (float32, float32) {
	switch t := target.(type) {
	case interface{ Center() (float32, float32) }:
		return t.Center()
	default:
		return intent.CursorX, intent.CursorY
	}
}
