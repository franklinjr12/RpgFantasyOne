package gamedata

import "testing"

func TestEnemyArchetypeRegistryHasMinimumRoster(t *testing.T) {
	pool := EnemyArchetypeTypes()
	if len(pool) < 6 {
		t.Fatalf("expected at least 6 archetypes, got %d", len(pool))
	}

	for _, archetypeType := range pool {
		archetype := GetEnemyArchetype(archetypeType)
		if archetype.Name == "" {
			t.Fatalf("expected archetype %d to have name", archetypeType)
		}
		if archetype.MaxHP <= 0 {
			t.Fatalf("expected archetype %s max hp > 0", archetype.Name)
		}
		if archetype.Damage <= 0 {
			t.Fatalf("expected archetype %s damage > 0", archetype.Name)
		}
		if archetype.MoveSpeed <= 0 {
			t.Fatalf("expected archetype %s move speed > 0", archetype.Name)
		}
		if archetype.AttackCooldown <= 0 {
			t.Fatalf("expected archetype %s attack cooldown > 0", archetype.Name)
		}
		if archetype.AttackRange <= 0 {
			t.Fatalf("expected archetype %s attack range > 0", archetype.Name)
		}
		if archetype.AggroRange <= 0 {
			t.Fatalf("expected archetype %s aggro range > 0", archetype.Name)
		}
		if archetype.Width <= 0 || archetype.Height <= 0 {
			t.Fatalf("expected archetype %s dimensions > 0", archetype.Name)
		}
		if archetype.XPReward <= 0 {
			t.Fatalf("expected archetype %s xp reward > 0", archetype.Name)
		}
		if archetype.ThreatValue <= 0 {
			t.Fatalf("expected archetype %s threat value > 0", archetype.Name)
		}
	}
}

func TestEliteModifierRegistryIsDeterministic(t *testing.T) {
	pool := EliteModifierTypes()
	if len(pool) < 2 {
		t.Fatalf("expected at least 2 elite modifiers, got %d", len(pool))
	}

	firstCall := EliteModifierTypes()
	secondCall := EliteModifierTypes()
	if len(firstCall) != len(secondCall) {
		t.Fatalf("elite modifier pool size changed between calls")
	}
	for i := range firstCall {
		if firstCall[i] != secondCall[i] {
			t.Fatalf("elite modifier pool order not deterministic at index %d", i)
		}
	}

	for _, modifierType := range pool {
		mod := GetEliteModifier(modifierType)
		if mod.Name == "" {
			t.Fatalf("expected modifier %d to have name", modifierType)
		}
		if mod.HPMultiplier <= 1 {
			t.Fatalf("expected modifier %s hp multiplier > 1", mod.Name)
		}
		if mod.DmgMultiplier <= 1 {
			t.Fatalf("expected modifier %s damage multiplier > 1", mod.Name)
		}
		if len(mod.OnHitEffects) == 0 {
			t.Fatalf("expected modifier %s to have on-hit effects", mod.Name)
		}
	}
}
