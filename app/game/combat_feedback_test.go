//go:build raylib

package game

import (
	"fmt"
	"testing"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/settings"
	"singlefantasy/app/systems"
)

func TestApplyCombatHitWithFeedbackSpawnsCritAndStatusText(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	enemy := gameobjects.NewEnemy(0, 0, false)
	roll := float32(0)

	g.applyCombatHitWithFeedback(systems.CombatHitRequest{
		Caster:             g.Player,
		Target:             enemy,
		BaseDamage:         10,
		DamageType:         gamedata.DamagePhysical,
		CritChance:         1,
		CritMultiplier:     2,
		Effects:            []gamedata.EffectSpec{{Type: gamedata.EffectStun, Duration: 1}, {Type: gamedata.EffectSilence, Duration: 1}, {Type: gamedata.EffectStun, Duration: 1}},
		ApplyOnHitHooks:    true,
		UseSourceModifiers: false,
		CritRoll:           &roll,
	})

	damageEvents := 0
	damageWasCrit := false
	statusCounts := map[string]int{}
	for _, event := range g.CombatTextEvents {
		if event == nil {
			continue
		}
		if event.Kind == CombatTextDamage {
			damageEvents++
			damageWasCrit = event.IsCrit
		}
		if event.Kind == CombatTextStatus {
			statusCounts[event.Text]++
		}
	}

	if damageEvents != 1 {
		t.Fatalf("expected one damage popup, got %d", damageEvents)
	}
	if !damageWasCrit {
		t.Fatalf("expected crit popup styling")
	}
	if statusCounts["STUNNED"] != 1 {
		t.Fatalf("expected one STUNNED popup, got %d", statusCounts["STUNNED"])
	}
	if statusCounts["SILENCED"] != 1 {
		t.Fatalf("expected one SILENCED popup, got %d", statusCounts["SILENCED"])
	}
}

func TestApplyCombatHitWithFeedbackSpawnsHealPopupFromLifesteal(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	g.Player.HP = g.Player.MaxHP - 40
	enemy := gameobjects.NewEnemy(0, 0, false)
	enemy.HP = 500
	enemy.MaxHP = 500
	roll := float32(1)
	beforeHP := g.Player.HP

	g.applyCombatHitWithFeedback(systems.CombatHitRequest{
		Caster:             g.Player,
		Target:             enemy,
		BaseDamage:         20,
		DamageType:         gamedata.DamagePhysical,
		CritChance:         0,
		CritMultiplier:     1.5,
		ApplyOnHitHooks:    true,
		UseSourceModifiers: false,
		CritRoll:           &roll,
	})

	healed := g.Player.HP - beforeHP
	if healed <= 0 {
		t.Fatalf("expected lifesteal healing")
	}

	expectedText := fmt.Sprintf("+%d", healed)
	foundHealPopup := false
	for _, event := range g.CombatTextEvents {
		if event == nil {
			continue
		}
		if event.Kind == CombatTextHeal && event.Text == expectedText {
			foundHealPopup = true
			break
		}
	}
	if !foundHealPopup {
		t.Fatalf("expected heal popup %q", expectedText)
	}
}

func TestRangedKillHealSpawnsFloatingHealPopup(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeRanged)
	g.Player.HP = g.Player.MaxHP - 50

	enemy := gameobjects.NewEnemy(0, 0, false)
	enemy.HP = 5
	enemy.MaxHP = 5
	g.Enemies = []*gameobjects.Enemy{enemy}

	enemyX, enemyY := enemy.Center()
	g.Projectiles = []*Projectile{
		{
			X:          enemyX,
			Y:          enemyY,
			Radius:     10,
			Damage:     10,
			Alive:      true,
			Lifetime:   1.0,
			HitTargets: map[interface{}]struct{}{},
			Caster:     g.Player,
			DamageType: gamedata.DamagePhysical,
		},
	}

	beforeHP := g.Player.HP
	system := &projectilesSystem{}
	system.updatePlayerProjectiles(g, 0)

	healed := g.Player.HP - beforeHP
	if healed != g.Player.Class.KillHealAmount {
		t.Fatalf("expected kill-heal amount %d, got %d", g.Player.Class.KillHealAmount, healed)
	}

	expectedText := fmt.Sprintf("+%d", healed)
	foundHealPopup := false
	for _, event := range g.CombatTextEvents {
		if event == nil {
			continue
		}
		if event.Kind == CombatTextHeal && event.Text == expectedText {
			foundHealPopup = true
			break
		}
	}
	if !foundHealPopup {
		t.Fatalf("expected heal popup %q", expectedText)
	}
}

func TestTryCastDirectionalSkillSpawnsDirectionalTelegraph(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	skill := gamedata.NewSkill(gamedata.SkillTypeShockwaveSlam)

	g.TryCastSkill(skill, &systems.Input{
		CursorWorldX: 160,
		CursorWorldY: 20,
	})

	if len(g.DirectionalTelegraphs) == 0 {
		t.Fatalf("expected directional telegraph")
	}
	first := g.DirectionalTelegraphs[0]
	if first.StartX == first.EndX && first.StartY == first.EndY {
		t.Fatalf("expected directional telegraph with non-zero length")
	}
}

func TestCombatHitFlashTimersUseConfiguredDurations(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	enemy := gameobjects.NewEnemy(0, 0, false)
	roll := float32(1)

	g.applyCombatHitWithFeedback(systems.CombatHitRequest{
		Caster:             g.Player,
		Target:             enemy,
		BaseDamage:         10,
		DamageType:         gamedata.DamagePhysical,
		CritMultiplier:     1.5,
		ApplyOnHitHooks:    true,
		UseSourceModifiers: false,
		CritRoll:           &roll,
	})
	if enemy.HitFlashTimer != gameobjects.EntityHitFlashDuration {
		t.Fatalf("expected enemy hit flash %.2f, got %.2f", gameobjects.EntityHitFlashDuration, enemy.HitFlashTimer)
	}

	hit := g.ApplyPlayerCombatHit(8, gamedata.DamagePhysical, -20, 0, nil)
	if !hit {
		t.Fatalf("expected player to receive hit")
	}
	if g.Player.HitFlashTimer != PlayerHitFlashDuration {
		t.Fatalf("expected player hit flash %.2f, got %.2f", PlayerHitFlashDuration, g.Player.HitFlashTimer)
	}
}
