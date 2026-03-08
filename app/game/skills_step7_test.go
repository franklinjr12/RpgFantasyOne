//go:build raylib

package game

import (
	"testing"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/settings"
	"singlefantasy/app/systems"
	"singlefantasy/app/world"
)

func TestRetreatRollDisplacementStaysWithinRoomBounds(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(8, 70, gamedata.ClassTypeRanged)
	g.CurrentRoom = &world.Room{X: 0, Y: 0, Width: 180, Height: 180}

	skill := gamedata.NewSkill(gamedata.SkillTypeRetreatRoll)
	centerX, centerY := g.Player.Center()
	input := &systems.Input{
		CursorWorldX: centerX + 150,
		CursorWorldY: centerY,
	}

	startX := g.Player.PosX
	g.TryCastSkill(skill, input)

	if g.Player.PosX >= startX {
		t.Fatalf("expected retreat roll to move player backward on X axis")
	}
	if g.Player.PosX < g.CurrentRoom.X {
		t.Fatalf("expected retreat roll to remain inside room bounds")
	}
}

func TestArcaneDrainManaRestoreClampsToMaxMana(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(60, 60, gamedata.ClassTypeCaster)

	enemyA := gameobjects.NewEnemy(65, 65, false)
	enemyB := gameobjects.NewEnemy(75, 70, false)
	g.Enemies = []*gameobjects.Enemy{enemyA, enemyB}

	skill := gamedata.NewSkill(gamedata.SkillTypeArcaneDrain)
	g.Player.Mana = g.Player.MaxMana - 5
	startMana := g.Player.Mana

	g.TryCastSkill(skill, nil)

	expected := startMana - skill.ManaCost + (2 * skill.ResourceGain.ManaPerTarget)
	if expected > g.Player.MaxMana {
		expected = g.Player.MaxMana
	}
	if g.Player.Mana != expected {
		t.Fatalf("expected clamped mana %d, got %d", expected, g.Player.Mana)
	}
}

func TestFrostFieldZoneReappliesSlowOverTime(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(100, 100, gamedata.ClassTypeCaster)
	enemy := gameobjects.NewEnemy(110, 110, false)
	g.Enemies = []*gameobjects.Enemy{enemy}

	skill := gamedata.NewSkill(gamedata.SkillTypeFrostField)
	centerX, centerY := g.Player.Center()
	input := &systems.Input{
		CursorWorldX: centerX,
		CursorWorldY: centerY,
	}

	g.TryCastSkill(skill, input)
	if len(g.DelayedSkillEffects) != 1 {
		t.Fatalf("expected one delayed zone effect, got %d", len(g.DelayedSkillEffects))
	}

	projectileSystem := &projectilesSystem{}
	projectileSystem.updateDelayedSkillEffects(g, skill.Delivery.Delay)
	if !gamedata.HasEffect(&enemy.Effects, gamedata.EffectSlow) {
		t.Fatalf("expected frost field to apply slow on trigger")
	}

	gamedata.UpdateEffects(&enemy.Effects, 0.8, nil)
	projectileSystem.updateDelayedSkillEffects(g, 1.0)
	if !gamedata.HasEffect(&enemy.Effects, gamedata.EffectSlow) {
		t.Fatalf("expected frost field to keep reapplying slow while zone is active")
	}
	if enemy.Effects[0].TimeLeft < 1.0 {
		t.Fatalf("expected refreshed slow duration from zone tick, got %.2f", enemy.Effects[0].TimeLeft)
	}
}
