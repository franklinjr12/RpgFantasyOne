package systems

import (
	"math/rand"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
)

type DamageRequest struct {
	Source             *gameobjects.Player
	Target             interface{}
	BaseDamage         float32
	DamageType         gamedata.DamageType
	CritChance         float32
	CritMultiplier     float32
	UseSourceModifiers bool
	SuppressFlash      bool
	CritRoll           *float32
}

type DamageResult struct {
	RequestedDamage int
	AppliedDamage   int
	IsCrit          bool
}

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

func ResolveAndApplyDamage(request DamageRequest) DamageResult {
	if request.Target == nil || request.BaseDamage <= 0 {
		return DamageResult{}
	}

	damage := request.BaseDamage
	if request.UseSourceModifiers && request.Source != nil {
		damage *= sourceDamageMultiplier(request.Source)
	}

	isCrit := false
	if shouldCrit(request) {
		isCrit = true
		damage *= resolveCritMultiplier(request.CritMultiplier)
	}

	finalDamage := int(damage)
	if finalDamage <= 0 {
		return DamageResult{}
	}

	applied := applyDamageToTarget(request.Target, finalDamage, request.DamageType, request.SuppressFlash)
	return DamageResult{
		RequestedDamage: finalDamage,
		AppliedDamage:   applied,
		IsCrit:          isCrit,
	}
}

func sourceDamageMultiplier(source *gameobjects.Player) float32 {
	if source == nil {
		return 1
	}
	multiplier := float32(1)
	if gamedata.HasEffect(&source.Effects, gamedata.EffectDamageBoost) {
		multiplier *= (1 + gamedata.GetEffectMagnitude(&source.Effects, gamedata.EffectDamageBoost))
	}
	return multiplier
}

func shouldCrit(request DamageRequest) bool {
	if request.CritChance <= 0 {
		return false
	}
	roll := rand.Float32()
	if request.CritRoll != nil {
		roll = *request.CritRoll
	}
	return roll < clamp01(request.CritChance)
}

func resolveCritMultiplier(multiplier float32) float32 {
	if multiplier <= 1 {
		return 1.5
	}
	return multiplier
}

func applyDamageToTarget(target interface{}, damage int, damageType gamedata.DamageType, suppressFlash bool) int {
	switch t := target.(type) {
	case *gameobjects.Player:
		return t.ApplyTypedDamage(damage, damageType, !suppressFlash)
	case *gameobjects.Enemy:
		before := t.HP
		t.TakeDamage(damage)
		if before < t.HP {
			return 0
		}
		return before - t.HP
	case *gameobjects.Boss:
		before := t.HP
		t.TakeDamage(damage)
		if before < t.HP {
			return 0
		}
		return before - t.HP
	default:
		return 0
	}
}

func clamp01(value float32) float32 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
