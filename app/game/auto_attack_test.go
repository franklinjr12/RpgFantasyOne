//go:build raylib
// +build raylib

package game

import (
	"testing"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/settings"
)

func TestCasterBasicAutoAttackRequiresMeleeRange(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(100, 100, gamedata.ClassTypeCaster)
	enemy := gameobjects.NewEnemy(190, 100, false)
	g.Enemies = []*gameobjects.Enemy{enemy}
	g.PlayerAttackTarget = enemy

	farHPBefore := enemy.HP
	g.UpdateAutoAttack(0)
	g.UpdateAutoAttack(CasterAttackWindup + 0.05)

	if enemy.HP != farHPBefore {
		t.Fatalf("expected no caster basic attack damage from out-of-melee range, hp %d -> %d", farHPBefore, enemy.HP)
	}
	if !g.HasPlayerMoveTarget {
		t.Fatalf("expected caster to path toward target when out of melee basic-attack range")
	}

	enemy.PosX = 130
	enemy.PosY = 100

	closeHPBefore := enemy.HP
	manaBefore := g.Player.Mana
	g.UpdateAutoAttack(0)
	g.UpdateAutoAttack(CasterAttackWindup + 0.05)

	if enemy.HP >= closeHPBefore {
		t.Fatalf("expected caster basic attack to deal damage in melee range")
	}
	if g.Player.Mana != manaBefore-g.Player.Class.ManaCost {
		t.Fatalf("expected caster mana to decrease by %d after basic hit, got %d -> %d", g.Player.Class.ManaCost, manaBefore, g.Player.Mana)
	}
}

