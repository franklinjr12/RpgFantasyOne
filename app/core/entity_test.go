package core

import "testing"

func TestEntityDamageAndAliveGuards(t *testing.T) {
	entity := &Entity{
		HP:    100,
		MaxHP: 100,
		Alive: true,
	}

	if applied := entity.ApplyDamage(-10); applied != 0 {
		t.Fatalf("expected 0 damage for negative input, got %d", applied)
	}

	if applied := entity.ApplyDamage(25); applied != 25 {
		t.Fatalf("expected applied damage 25, got %d", applied)
	}
	if entity.HP != 75 {
		t.Fatalf("expected HP 75, got %d", entity.HP)
	}
	if !entity.IsAlive() {
		t.Fatalf("entity should still be alive")
	}

	if applied := entity.ApplyDamage(200); applied != 75 {
		t.Fatalf("expected applied damage 75, got %d", applied)
	}
	if entity.HP != 0 {
		t.Fatalf("expected HP 0, got %d", entity.HP)
	}
	if entity.IsAlive() {
		t.Fatalf("entity should be dead")
	}
}

func TestEntityHealGuards(t *testing.T) {
	entity := &Entity{
		HP:    40,
		MaxHP: 100,
		Alive: true,
	}

	if healed := entity.Heal(-5); healed != 0 {
		t.Fatalf("expected 0 heal for negative input, got %d", healed)
	}

	if healed := entity.Heal(30); healed != 30 {
		t.Fatalf("expected healed amount 30, got %d", healed)
	}
	if entity.HP != 70 {
		t.Fatalf("expected HP 70, got %d", entity.HP)
	}

	if healed := entity.Heal(60); healed != 30 {
		t.Fatalf("expected capped heal amount 30, got %d", healed)
	}
	if entity.HP != 100 {
		t.Fatalf("expected HP 100, got %d", entity.HP)
	}
}
