package gamedata

import "testing"

func TestStep7SkillSpecsExposeRuntimeBehaviorData(t *testing.T) {
	retreatRoll := NewSkill(SkillTypeRetreatRoll)
	if retreatRoll.SelfMovement.Mode != SelfMovementBackwardFromCursor {
		t.Fatalf("expected retreat roll movement mode backward-from-cursor")
	}
	if retreatRoll.SelfMovement.Distance <= 0 {
		t.Fatalf("expected retreat roll movement distance > 0")
	}

	manaShield := NewSkill(SkillTypeManaShield)
	if manaShield.ManaShield.AbsorbFromCurrentManaRatio <= 0 {
		t.Fatalf("expected mana shield absorb ratio > 0")
	}
	if manaShield.ManaShield.Duration <= 0 {
		t.Fatalf("expected mana shield duration > 0")
	}

	frostField := NewSkill(SkillTypeFrostField)
	if frostField.Delivery.ZoneDuration <= 0 {
		t.Fatalf("expected frost field zone duration > 0")
	}
	if frostField.Delivery.ZoneTickRate <= 0 {
		t.Fatalf("expected frost field zone tick rate > 0")
	}

	arcaneDrain := NewSkill(SkillTypeArcaneDrain)
	if arcaneDrain.ResourceGain.ManaPerTarget <= 0 {
		t.Fatalf("expected arcane drain mana gain per target > 0")
	}
}

func TestStep7PoisonTipPercentHPTickHasCaps(t *testing.T) {
	poisonTip := NewSkill(SkillTypePoisonTip)
	if len(poisonTip.Effects) == 0 {
		t.Fatalf("expected poison tip effects")
	}

	poison := poisonTip.Effects[0]
	if poison.Type != EffectPoison {
		t.Fatalf("expected poison tip primary effect to be poison")
	}
	if poison.PercentMaxHPPerTick <= 0 {
		t.Fatalf("expected poison tip to include percent max hp per tick")
	}
	if poison.MaxTickDamage <= 0 {
		t.Fatalf("expected poison tip max tick damage cap > 0")
	}
	if poison.MinTickDamage <= 0 {
		t.Fatalf("expected poison tip min tick damage > 0")
	}
}
