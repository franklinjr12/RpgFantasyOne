//go:build raylib

package systems

import (
	"singlefantasy/app/core"
	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"testing"
)

func TestResolveTargetsSingleTargetPrefersHovered(t *testing.T) {
	player := gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	enemyA := testEnemy(40, 20, 30, 30)
	enemyB := testEnemy(100, 20, 30, 30)

	intent := BuildCastIntent(player, enemyB.PosX+5, enemyB.PosY+5)
	spec := gamedata.TargetingSpec{
		Type:       gamedata.TargetEnemy,
		Range:      200,
		MaxTargets: 1,
	}

	targets := ResolveTargets(player, intent, spec, []*gameobjects.Enemy{enemyA, enemyB}, nil)
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0] != enemyB {
		t.Fatalf("expected hovered enemyB target")
	}
}

func TestResolveTargetsSingleTargetFallsBackToNearestAndStableOrder(t *testing.T) {
	player := gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	enemyA := testEnemy(60, 20, 30, 30)
	enemyB := testEnemy(-20, 20, 30, 30)

	intent := BuildCastIntent(player, 400, 400)
	spec := gamedata.TargetingSpec{
		Type:       gamedata.TargetEnemy,
		Range:      200,
		MaxTargets: 1,
	}

	targets := ResolveTargets(player, intent, spec, []*gameobjects.Enemy{enemyA, enemyB}, nil)
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0] != enemyA {
		t.Fatalf("expected stable first enemy target for distance tie")
	}
}

func TestResolveTargetsArea(t *testing.T) {
	player := gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	enemyNear := testEnemy(120, 20, 30, 30)
	enemyFar := testEnemy(260, 20, 30, 30)

	intent := BuildCastIntent(player, 120, 20)
	spec := gamedata.TargetingSpec{
		Type:       gamedata.TargetArea,
		Range:      300,
		Radius:     80,
		MaxTargets: 10,
	}

	targets := ResolveTargets(player, intent, spec, []*gameobjects.Enemy{enemyNear, enemyFar}, nil)
	if len(targets) != 1 {
		t.Fatalf("expected 1 target in area, got %d", len(targets))
	}
	if targets[0] != enemyNear {
		t.Fatalf("expected near enemy in area result")
	}
}

func TestResolveTargetsDirectionalCone(t *testing.T) {
	player := gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	front := testEnemy(90, 20, 30, 30)
	side := testEnemy(20, 120, 30, 30)

	intent := BuildCastIntent(player, 300, 20)
	spec := gamedata.TargetingSpec{
		Type:                  gamedata.TargetDirection,
		Range:                 150,
		MaxTargets:            10,
		DirectionalArcDegrees: 70,
	}

	targets := ResolveTargets(player, intent, spec, []*gameobjects.Enemy{front, side}, nil)
	if len(targets) != 1 {
		t.Fatalf("expected 1 directional target, got %d", len(targets))
	}
	if targets[0] != front {
		t.Fatalf("expected front enemy in directional cone")
	}
}

func TestResolveTargetsDirectionalLine(t *testing.T) {
	player := gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	inLine := testEnemy(100, 40, 20, 20)
	outsideWidth := testEnemy(100, 90, 20, 20)

	intent := BuildCastIntent(player, 300, 20)
	spec := gamedata.TargetingSpec{
		Type:                 gamedata.TargetDirection,
		Range:                150,
		MaxTargets:           10,
		DirectionalLineWidth: 40,
	}

	targets := ResolveTargets(player, intent, spec, []*gameobjects.Enemy{inLine, outsideWidth}, nil)
	if len(targets) != 1 {
		t.Fatalf("expected 1 directional line target, got %d", len(targets))
	}
	if targets[0] != inLine {
		t.Fatalf("expected line-width target only")
	}
}

func TestCanCastValidation(t *testing.T) {
	player := gameobjects.NewPlayer(0, 0, gamedata.ClassTypeCaster)
	skill := gamedata.NewSkill(gamedata.SkillTypeArcaneBolt)

	if !CanCast(player, skill) {
		t.Fatalf("expected cast to be allowed with default state")
	}

	skill.CurrentCooldown = 1
	if CanCast(player, skill) {
		t.Fatalf("expected cooldown to block cast")
	}
	skill.CurrentCooldown = 0

	player.Mana = 0
	if CanCast(player, skill) {
		t.Fatalf("expected insufficient mana to block cast")
	}
	player.Mana = player.MaxMana

	gamedata.ApplyEffect(&player.Effects, gamedata.Effect{Type: gamedata.EffectSilence, Duration: 1})
	if CanCast(player, skill) {
		t.Fatalf("expected silence to block cast")
	}
	player.Effects = nil

	gamedata.ApplyEffect(&player.Effects, gamedata.Effect{Type: gamedata.EffectStun, Duration: 1})
	if CanCast(player, skill) {
		t.Fatalf("expected stun to block cast")
	}
}

func testEnemy(x, y, w, h float32) *gameobjects.Enemy {
	return &gameobjects.Enemy{
		Entity: core.Entity{
			PosX:   x,
			PosY:   y,
			HP:     100,
			MaxHP:  100,
			Hitbox: core.Hitbox{Width: w, Height: h},
			Alive:  true,
		},
	}
}
