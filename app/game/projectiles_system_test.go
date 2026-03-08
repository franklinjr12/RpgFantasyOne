//go:build raylib

package game

import (
	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/settings"
	"testing"
)

func TestPlayerProjectileExpiresByLifetime(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeRanged)
	g.Projectiles = []*Projectile{
		{
			X:          0,
			Y:          0,
			Alive:      true,
			Lifetime:   0.05,
			HitTargets: map[interface{}]struct{}{},
		},
	}

	system := &projectilesSystem{}
	system.updatePlayerProjectiles(g, 0.1)

	if len(g.Projectiles) != 0 {
		t.Fatalf("expected expired projectile to be removed, got %d", len(g.Projectiles))
	}
}

func TestPlayerProjectileCollisionAppliesDamage(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeRanged)

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
		},
	}

	system := &projectilesSystem{}
	system.updatePlayerProjectiles(g, 0)

	if enemy.IsAlive() {
		t.Fatalf("expected enemy to die from projectile hit")
	}
	if len(g.Projectiles) != 0 {
		t.Fatalf("expected spent projectile to be removed, got %d", len(g.Projectiles))
	}
}

func TestPlayerProjectilePierceHitsMultipleTargets(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeRanged)

	enemyA := gameobjects.NewEnemy(0, 0, false)
	enemyA.HP = 5
	enemyA.MaxHP = 5
	enemyB := gameobjects.NewEnemy(0, 0, false)
	enemyB.HP = 5
	enemyB.MaxHP = 5
	g.Enemies = []*gameobjects.Enemy{enemyA, enemyB}

	enemyX, enemyY := enemyA.Center()
	g.Projectiles = []*Projectile{
		{
			X:          enemyX,
			Y:          enemyY,
			Radius:     10,
			Damage:     10,
			Alive:      true,
			Lifetime:   1.0,
			Pierce:     1,
			HitTargets: map[interface{}]struct{}{},
		},
	}

	system := &projectilesSystem{}
	system.updatePlayerProjectiles(g, 0)

	if enemyA.IsAlive() || enemyB.IsAlive() {
		t.Fatalf("expected both enemies to be hit by piercing projectile")
	}
	if len(g.Projectiles) != 0 {
		t.Fatalf("expected projectile to be removed after exhausting pierce, got %d", len(g.Projectiles))
	}
}

func TestPlayerProjectileDoesNotRehitSameTarget(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeRanged)

	enemy := gameobjects.NewEnemy(0, 0, false)
	enemy.HP = 30
	enemy.MaxHP = 30
	g.Enemies = []*gameobjects.Enemy{enemy}

	enemyX, enemyY := enemy.Center()
	g.Projectiles = []*Projectile{
		{
			X:          enemyX,
			Y:          enemyY,
			Radius:     10,
			Damage:     10,
			Alive:      true,
			Lifetime:   2.0,
			Pierce:     2,
			HitTargets: map[interface{}]struct{}{},
		},
	}

	system := &projectilesSystem{}
	system.updatePlayerProjectiles(g, 0)
	hpAfterFirstHit := enemy.HP
	system.updatePlayerProjectiles(g, 0)

	if enemy.HP != hpAfterFirstHit {
		t.Fatalf("expected no repeat damage to same target from same projectile")
	}
}
