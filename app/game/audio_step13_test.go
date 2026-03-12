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

func TestDamageSFXRoutesByTargetType(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)

	played := make([]string, 0, 4)
	g.soundPlayer = func(key string) {
		played = append(played, key)
	}

	enemy := gameobjects.NewEnemy(0, 0, false)
	g.applyCombatHitWithFeedback(systems.CombatHitRequest{
		Caster:             g.Player,
		Target:             enemy,
		BaseDamage:         10,
		DamageType:         gamedata.DamagePhysical,
		CritMultiplier:     1.5,
		ApplyOnHitHooks:    true,
		UseSourceModifiers: false,
	})

	g.updateSoundCooldowns(hitSoundCooldownSeconds)
	hit := g.ApplyPlayerCombatHit(8, gamedata.DamagePhysical, -20, 0, nil)
	if !hit {
		t.Fatalf("expected player hit")
	}

	if !containsSoundKey(played, sfxEnemyHit) {
		t.Fatalf("expected enemy-hit sound")
	}
	if !containsSoundKey(played, sfxPlayerHit) {
		t.Fatalf("expected player-hit sound")
	}
}

func TestPassiveLifestealDoesNotPlayHealingSFX(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	g.Player.HP = g.Player.MaxHP - 40

	played := make([]string, 0, 4)
	g.soundPlayer = func(key string) {
		played = append(played, key)
	}

	enemy := gameobjects.NewEnemy(0, 0, false)
	enemy.HP = 500
	enemy.MaxHP = 500

	g.applyCombatHitWithFeedback(systems.CombatHitRequest{
		Caster:             g.Player,
		Target:             enemy,
		BaseDamage:         20,
		DamageType:         gamedata.DamagePhysical,
		CritMultiplier:     1.5,
		ApplyOnHitHooks:    true,
		UseSourceModifiers: false,
	})

	foundHealPopup := false
	for _, event := range g.CombatTextEvents {
		if event != nil && event.Kind == CombatTextHeal {
			foundHealPopup = true
			break
		}
	}
	if !foundHealPopup {
		t.Fatalf("expected lifesteal heal popup")
	}
	if containsSoundKey(played, sfxPlayerHealing) {
		t.Fatalf("expected passive lifesteal to suppress healing sound")
	}
}

func TestGrantPlayerXPPlaysLevelUpSFXOnThresholdCross(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeCaster)

	played := make([]string, 0, 2)
	g.soundPlayer = func(key string) {
		played = append(played, key)
	}

	g.grantPlayerXP(g.Player.XPToNext)

	if g.Player.Level != 2 {
		t.Fatalf("expected level 2, got %d", g.Player.Level)
	}
	if countSoundKey(played, sfxPlayerLevelUp) != 1 {
		t.Fatalf("expected one level-up sound, got %d", countSoundKey(played, sfxPlayerLevelUp))
	}
}

func TestDoorOpenSFXTriggersOnceOnUnlockEdge(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeRanged)
	g.CurrentRoom = &world.Room{
		Type: world.RoomTypeCombat,
		Doors: []*world.Door{
			{
				Bounds: world.AABB{X: 900, Y: 900, Width: 40, Height: 80},
				Locked: true,
			},
		},
	}
	g.Enemies = []*gameobjects.Enemy{}

	played := make([]string, 0, 2)
	g.soundPlayer = func(key string) {
		played = append(played, key)
	}

	system := &dungeonRunSystem{}
	system.Update(NewRuntimeContext(g), 0.016)
	system.Update(NewRuntimeContext(g), 0.016)

	if countSoundKey(played, sfxDoorOpen) != 1 {
		t.Fatalf("expected one door-open sound, got %d", countSoundKey(played, sfxDoorOpen))
	}
	if roomHasLockedDoor(g.CurrentRoom) {
		t.Fatalf("expected room doors to be unlocked")
	}
}

func containsSoundKey(played []string, key string) bool {
	return countSoundKey(played, key) > 0
}

func countSoundKey(played []string, key string) int {
	count := 0
	for _, value := range played {
		if value == key {
			count++
		}
	}
	return count
}
