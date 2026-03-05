package systems

import (
	"math"
	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
)

func ComputeDamage(spec *gamedata.DamageSpec, stats *gamedata.Stats) float32 {
	if spec == nil || stats == nil {
		return 0
	}

	damage := spec.Base
	for stat, factor := range spec.Scaling {
		damage += float32(stats.GetStat(stat)) * factor
	}

	return damage
}

func ResolveTargets(caster *gameobjects.Player, intentX, intentY float32, spec gamedata.TargetingSpec, enemies []*gameobjects.Enemy, boss *gameobjects.Boss) []interface{} {
	var targets []interface{}

	casterX, casterY := caster.Center()

	switch spec.Type {
	case gamedata.TargetSelf:
		targets = append(targets, caster)
	case gamedata.TargetEnemy:
		for _, enemy := range enemies {
			if !enemy.IsAlive() {
				continue
			}
			enemyX, enemyY := enemy.Center()
			dx := enemyX - casterX
			dy := enemyY - casterY
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if distance <= spec.Range {
				targets = append(targets, enemy)
				if spec.MaxTargets > 0 && len(targets) >= spec.MaxTargets {
					break
				}
			}
		}

		if boss != nil && boss.IsAlive() {
			bossX, bossY := boss.Center()
			dx := bossX - casterX
			dy := bossY - casterY
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if distance <= spec.Range {
				targets = append(targets, boss)
			}
		}
	case gamedata.TargetArea:
		centerX := intentX
		centerY := intentY
		if spec.Range == 0 {
			centerX = casterX
			centerY = casterY
		}

		for _, enemy := range enemies {
			if !enemy.IsAlive() {
				continue
			}
			enemyX, enemyY := enemy.Center()
			dx := enemyX - centerX
			dy := enemyY - centerY
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if distance <= spec.Radius {
				targets = append(targets, enemy)
				if spec.MaxTargets > 0 && len(targets) >= spec.MaxTargets {
					break
				}
			}
		}

		if boss != nil && boss.IsAlive() {
			bossX, bossY := boss.Center()
			dx := bossX - centerX
			dy := bossY - centerY
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if distance <= spec.Radius {
				targets = append(targets, boss)
			}
		}
	}

	return targets
}

func ApplySkill(caster *gameobjects.Player, skill *gamedata.Skill, targets []interface{}) {
	for _, target := range targets {
		if skill.DamageSpec != nil {
			rawDamage := ComputeDamage(skill.DamageSpec, caster.Stats)

			if gamedata.HasEffect(&caster.Effects, gamedata.EffectDamageBoost) {
				magnitude := gamedata.GetEffectMagnitude(&caster.Effects, gamedata.EffectDamageBoost)
				rawDamage *= (1.0 + magnitude)
			}

			finalDamage := int(rawDamage)

			switch t := target.(type) {
			case *gameobjects.Player:
				t.TakeDamage(finalDamage)
			case *gameobjects.Enemy:
				t.TakeDamage(finalDamage)
			case *gameobjects.Boss:
				t.TakeDamage(finalDamage)
			}

			if skill.DamageSpec.DamageType == gamedata.DamagePhysical && caster.Class.LifestealPercent > 0 {
				lifesteal := int(float32(finalDamage) * caster.Class.LifestealPercent)
				caster.Heal(lifesteal)
			}

			if gamedata.HasEffect(&caster.Effects, gamedata.EffectLifesteal) {
				magnitude := gamedata.GetEffectMagnitude(&caster.Effects, gamedata.EffectLifesteal)
				lifesteal := int(float32(finalDamage) * magnitude)
				caster.Heal(lifesteal)
			}
		}

		for _, effectSpec := range skill.Effects {
			effect := gamedata.Effect{
				Type:      effectSpec.Type,
				Duration:  effectSpec.Duration,
				Magnitude: effectSpec.Magnitude,
				TickRate:  effectSpec.TickRate,
			}

			switch t := target.(type) {
			case *gameobjects.Player:
				gamedata.ApplyEffect(&t.Effects, effect)
			case *gameobjects.Enemy:
				gamedata.ApplyEffect(&t.Effects, effect)
			case *gameobjects.Boss:
				gamedata.ApplyEffect(&t.Effects, effect)
			}
		}
	}
}

func CanCast(caster *gameobjects.Player, skill *gamedata.Skill) bool {
	if !skill.CanUse() {
		return false
	}

	if skill.ManaCost > 0 && !caster.CanUseMana(skill.ManaCost) {
		return false
	}

	if gamedata.HasEffect(&caster.Effects, gamedata.EffectSilence) || gamedata.HasEffect(&caster.Effects, gamedata.EffectStun) {
		return false
	}

	return true
}
