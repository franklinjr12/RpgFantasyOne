package gameobjects

import (
	"testing"

	"singlefantasy/app/gamedata"
)

func TestResolveEnemyIntentMeleeChaseAndAttack(t *testing.T) {
	enemy := NewEnemyFromArchetype(0, 0, gamedata.EnemyArchetypeRaider, false, gamedata.EliteModifierScorching)
	enemy.Update(0.016)

	ResolveEnemyIntent(enemy, 300, 0)
	if enemy.State != EnemyStateChasing {
		t.Fatalf("expected chase state, got %d", enemy.State)
	}
	if enemy.IntentMoveX <= 0 {
		t.Fatalf("expected positive chase movement intent")
	}
	if enemy.WantsAttack {
		t.Fatalf("expected no attack intent while out of range")
	}

	enemy.Update(0.016)
	ResolveEnemyIntent(enemy, 20, 20)
	if enemy.State != EnemyStateAttacking {
		t.Fatalf("expected attacking state, got %d", enemy.State)
	}
	if !enemy.WantsAttack {
		t.Fatalf("expected attack intent when in range")
	}
}

func TestResolveEnemyIntentRangedRetreat(t *testing.T) {
	enemy := NewEnemyFromArchetype(100, 100, gamedata.EnemyArchetypeArcher, false, gamedata.EliteModifierScorching)
	enemy.Update(0.016)

	ResolveEnemyIntent(enemy, 110, 110)
	if enemy.State != EnemyStateRetreating {
		t.Fatalf("expected retreating state, got %d", enemy.State)
	}
	if enemy.IntentMoveX == 0 && enemy.IntentMoveY == 0 {
		t.Fatalf("expected non-zero retreat intent")
	}
}

func TestEliteAttackPayloadIncludesModifierEffects(t *testing.T) {
	enemy := NewEnemyFromArchetype(0, 0, gamedata.EnemyArchetypeRaider, true, gamedata.EliteModifierScorching)
	enemy.State = EnemyStateAttacking
	enemy.WantsAttack = true
	enemy.CurrentCooldown = 0

	hit, payload := enemy.Attack(60, 0)
	if !hit {
		t.Fatalf("expected elite attack hit payload")
	}
	if payload.Damage <= 0 {
		t.Fatalf("expected elite attack damage > 0")
	}
	if len(payload.OnHitEffects) == 0 {
		t.Fatalf("expected elite attack payload to include modifier effects")
	}

	hasBurn := false
	for _, effect := range payload.OnHitEffects {
		if effect.Type == gamedata.EffectBurn {
			hasBurn = true
			break
		}
	}
	if !hasBurn {
		t.Fatalf("expected scorching modifier to add burn effect on hit")
	}
}
