//go:build raylib
// +build raylib

package systems

import (
	"testing"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
)

func TestResolveAndApplyDamageSupportsCrit(t *testing.T) {
	enemy := gameobjects.NewEnemy(0, 0, false)
	roll := float32(0)
	result := ResolveAndApplyDamage(DamageRequest{
		Target:         enemy,
		BaseDamage:     10,
		DamageType:     gamedata.DamagePhysical,
		CritChance:     1,
		CritMultiplier: 2,
		CritRoll:       &roll,
	})

	if !result.IsCrit {
		t.Fatalf("expected critical hit")
	}
	if result.RequestedDamage != 20 {
		t.Fatalf("expected crit requested damage 20, got %d", result.RequestedDamage)
	}
	if result.AppliedDamage != 20 {
		t.Fatalf("expected crit applied damage 20, got %d", result.AppliedDamage)
	}
}

func TestResolveAndApplyDamageUsesTypedMitigation(t *testing.T) {
	player := gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	roll := float32(1)
	startHP := player.HP
	result := ResolveAndApplyDamage(DamageRequest{
		Target:         player,
		BaseDamage:     100,
		DamageType:     gamedata.DamagePhysical,
		CritChance:     0,
		CritMultiplier: 1.5,
		CritRoll:       &roll,
		SuppressFlash:  true,
	})

	if result.AppliedDamage != 93 {
		t.Fatalf("expected physical mitigation applied damage 93, got %d", result.AppliedDamage)
	}
	if player.HP != startHP-93 {
		t.Fatalf("expected hp %d, got %d", startHP-93, player.HP)
	}
}

func TestApplyCombatHitAppliesSkillEffectsAndOnHitHooks(t *testing.T) {
	caster := gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	caster.HP = 40
	enemy := gameobjects.NewEnemy(0, 0, false)
	skill := gamedata.NewSkill(gamedata.SkillTypeShockwaveSlam)
	gamedata.ApplyEffect(&caster.Effects, gamedata.Effect{
		Type:      gamedata.EffectLifesteal,
		Duration:  5,
		Magnitude: 0.5,
	})

	roll := float32(1)
	result := ApplyCombatHit(CombatHitRequest{
		Caster:             caster,
		Target:             enemy,
		Skill:              skill,
		ApplyOnHitHooks:    true,
		UseSourceModifiers: true,
		CritRoll:           &roll,
	})

	if result.Damage.AppliedDamage <= 0 {
		t.Fatalf("expected skill damage to be applied")
	}
	if !gamedata.HasEffect(&enemy.Effects, gamedata.EffectSlow) {
		t.Fatalf("expected skill slow effect to be applied")
	}
	if caster.HP <= 40 {
		t.Fatalf("expected on-hit lifesteal healing, got hp %d", caster.HP)
	}
}

func TestApplyCombatHitSupportsAutoAttackPath(t *testing.T) {
	caster := gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	caster.HP = 20
	enemy := gameobjects.NewEnemy(0, 0, false)
	roll := float32(1)

	result := ApplyCombatHit(CombatHitRequest{
		Caster:             caster,
		Target:             enemy,
		BaseDamage:         20,
		DamageType:         gamedata.DamagePhysical,
		CritChance:         0,
		CritMultiplier:     1.5,
		ApplyOnHitHooks:    true,
		UseSourceModifiers: false,
		CritRoll:           &roll,
	})

	if result.Damage.AppliedDamage != 20 {
		t.Fatalf("expected base damage 20, got %d", result.Damage.AppliedDamage)
	}
	if caster.HP != 24 {
		t.Fatalf("expected class lifesteal heal to hp 24, got %d", caster.HP)
	}
}
