package systems

import (
	"math"
	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
)

type TargetType int

const (
	TargetSelf TargetType = iota
	TargetEnemy
	TargetArea
)

type TargetingSpec struct {
	Type       TargetType
	Range      float32
	Radius     float32
	MaxTargets int
}

type DeliveryType int

const (
	DeliveryInstant DeliveryType = iota
	DeliveryProjectile
	DeliveryDelayed
)

type DeliverySpec struct {
	Type     DeliveryType
	Speed    float32
	Delay    float32
	Lifetime float32
}

type DamageType int

const (
	DamagePhysical DamageType = iota
	DamageMagical
	DamageTrue
)

type DamageSpec struct {
	Base       float32
	Scaling    map[gamedata.StatType]float32
	DamageType DamageType
	CritChance float32
	CritMult   float32
}

type EffectSpec struct {
	Type      EffectType
	Duration  float32
	Magnitude float32
	TickRate  float32
}

func ComputeDamage(spec *gamedata.DamageSpec, stats *gamedata.Stats) float32 {
	if spec == nil {
		return 0
	}
	dmg := spec.Base

	if spec.Scaling != nil {
		for stat, factor := range spec.Scaling {
			statValue := float32(stats.GetStat(stat))
			dmg += statValue * factor
		}
	}

	return dmg
}

func ResolveTargets(caster *gameobjects.Player, intentX, intentY float32, spec gamedata.TargetingSpec, enemies []*gameobjects.Enemy, boss *gameobjects.Boss) []interface{} {
	var targets []interface{}

	casterX := caster.X + caster.Width/2
	casterY := caster.Y + caster.Height/2

	switch spec.Type {
	case gamedata.TargetSelf:
		targets = append(targets, caster)

	case gamedata.TargetEnemy:
		if spec.Range > 0 {
			for _, enemy := range enemies {
				if !enemy.Alive {
					continue
				}
				enemyX := enemy.X + enemy.Width/2
				enemyY := enemy.Y + enemy.Height/2
				dx := enemyX - casterX
				dy := enemyY - casterY
				distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
				if distance <= spec.Range {
					targets = append(targets, enemy)
					if len(targets) >= spec.MaxTargets {
						break
					}
				}
			}
			if boss != nil && boss.Alive {
				bossX := boss.X + boss.Width/2
				bossY := boss.Y + boss.Height/2
				dx := bossX - casterX
				dy := bossY - casterY
				distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
				if distance <= spec.Range {
					targets = append(targets, boss)
				}
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
			if !enemy.Alive {
				continue
			}
			enemyX := enemy.X + enemy.Width/2
			enemyY := enemy.Y + enemy.Height/2
			dx := enemyX - centerX
			dy := enemyY - centerY
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if distance <= spec.Radius {
				targets = append(targets, enemy)
				if len(targets) >= spec.MaxTargets {
					break
				}
			}
		}
		if boss != nil && boss.Alive {
			bossX := boss.X + boss.Width/2
			bossY := boss.Y + boss.Height/2
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

			var finalDamage int
			switch t := target.(type) {
			case *gameobjects.Player:
				finalDamage = int(rawDamage)
				t.TakeDamage(finalDamage)
			case *gameobjects.Enemy:
				finalDamage = int(rawDamage)
				t.TakeDamage(finalDamage)
			case *gameobjects.Boss:
				finalDamage = int(rawDamage)
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

		if skill.Effects != nil {
			for _, effectSpec := range skill.Effects {
				switch t := target.(type) {
				case *gameobjects.Player:
					gamedata.ApplyEffect(&t.Effects, gamedata.Effect{
						Type:      gamedata.EffectType(effectSpec.Type),
						Duration:  effectSpec.Duration,
						Magnitude: effectSpec.Magnitude,
						TickRate:  effectSpec.TickRate,
					})
				case *gameobjects.Enemy:
					gamedata.ApplyEffect(&t.Effects, gamedata.Effect{
						Type:      gamedata.EffectType(effectSpec.Type),
						Duration:  effectSpec.Duration,
						Magnitude: effectSpec.Magnitude,
						TickRate:  effectSpec.TickRate,
					})
				case *gameobjects.Boss:
					if t.Enemy != nil {
						gamedata.ApplyEffect(&t.Enemy.Effects, gamedata.Effect{
							Type:      gamedata.EffectType(effectSpec.Type),
							Duration:  effectSpec.Duration,
							Magnitude: effectSpec.Magnitude,
							TickRate:  effectSpec.TickRate,
						})
					}
				}
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

