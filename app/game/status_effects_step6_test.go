//go:build raylib
// +build raylib

package game

import (
	"testing"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/settings"
	"singlefantasy/app/systems"
)

func TestGetPlayerMoveSpeedFreezeAndSlow(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(100, 100, gamedata.ClassTypeMelee)

	baseSpeed := g.Player.MoveSpeed
	gamedata.ApplyEffect(&g.Player.Effects, gamedata.Effect{Type: gamedata.EffectSlow, Duration: 2, Magnitude: 0.5})
	if g.GetPlayerMoveSpeed() >= baseSpeed {
		t.Fatalf("expected slow to reduce move speed")
	}

	gamedata.ApplyEffect(&g.Player.Effects, gamedata.Effect{Type: gamedata.EffectFreeze, Duration: 1})
	if g.GetPlayerMoveSpeed() != 0 {
		t.Fatalf("expected freeze to set move speed to zero")
	}
}

func TestSilenceBlocksCastingButNotMovement(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(100, 100, gamedata.ClassTypeCaster)
	skill := g.Player.Skills[0]

	gamedata.ApplyEffect(&g.Player.Effects, gamedata.Effect{Type: gamedata.EffectSilence, Duration: 2})
	if systems.CanCast(g.Player, skill) {
		t.Fatalf("expected silence to block casting")
	}
	if g.GetPlayerMoveSpeed() <= 0 {
		t.Fatalf("expected silence to preserve movement")
	}
}

func TestStunBlocksAutoAttackWindupAndResolve(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(100, 100, gamedata.ClassTypeMelee)
	enemy := gameobjects.NewEnemy(110, 100, false)
	g.Enemies = []*gameobjects.Enemy{enemy}
	g.PlayerAttackTarget = enemy

	gamedata.ApplyEffect(&g.Player.Effects, gamedata.Effect{Type: gamedata.EffectStun, Duration: 1})
	startEnemyHP := enemy.HP
	g.UpdateAutoAttack(0.2)

	if g.Player.AttackState != gameobjects.PlayerAttackStateIdle {
		t.Fatalf("expected stunned player to stay idle, got state %d", g.Player.AttackState)
	}
	if enemy.HP != startEnemyHP {
		t.Fatalf("expected stunned player to deal no auto attack damage")
	}
}
