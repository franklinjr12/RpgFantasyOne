//go:build raylib

package game

import (
	"testing"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/settings"
)

func TestApplyPlayerCombatHitAppliesEnemyEffects(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	startHP := g.Player.HP

	hit := g.ApplyPlayerCombatHit(8, gamedata.DamagePhysical, -20, 0, []gamedata.EffectSpec{
		{
			Type:      gamedata.EffectBurn,
			Duration:  2.0,
			Magnitude: 2.0,
			TickRate:  1.0,
		},
	})
	if !hit {
		t.Fatalf("expected combat hit to apply")
	}
	if g.Player.HP >= startHP {
		t.Fatalf("expected player hp to be reduced")
	}
	if !gamedata.HasEffect(&g.Player.Effects, gamedata.EffectBurn) {
		t.Fatalf("expected burn effect to be applied from enemy hit")
	}
}
