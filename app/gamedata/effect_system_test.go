package gamedata

import (
	"math"
	"testing"
)

func TestApplyEffectRefreshesDurationAndKeepsStrongestMagnitude(t *testing.T) {
	effects := []EffectInstance{}

	ApplyEffect(&effects, Effect{Type: EffectSlow, Duration: 2, Magnitude: 0.2})
	if len(effects) != 1 {
		t.Fatalf("expected one effect instance, got %d", len(effects))
	}

	ApplyEffect(&effects, Effect{Type: EffectSlow, Duration: 5, Magnitude: 0.1})
	if len(effects) != 1 {
		t.Fatalf("expected non-stacking refresh, got %d instances", len(effects))
	}
	if effects[0].TimeLeft != 5 {
		t.Fatalf("expected duration refresh to 5, got %.2f", effects[0].TimeLeft)
	}
	if effects[0].Magnitude != 0.2 {
		t.Fatalf("expected strongest magnitude to persist, got %.2f", effects[0].Magnitude)
	}

	ApplyEffect(&effects, Effect{Type: EffectSlow, Duration: 3, Magnitude: 0.6})
	if effects[0].Magnitude != 0.6 {
		t.Fatalf("expected strongest magnitude update to 0.6, got %.2f", effects[0].Magnitude)
	}
}

func TestUpdateEffectsSupportsLargeDeltaAndDotOnlyTicks(t *testing.T) {
	effects := []EffectInstance{
		{
			Effect: Effect{
				Type:      EffectBurn,
				Duration:  5,
				Magnitude: 2,
				TickRate:  1,
			},
			TimeLeft: 5,
		},
		{
			Effect: Effect{
				Type:      EffectSlow,
				Duration:  5,
				Magnitude: 10,
				TickRate:  1,
			},
			TimeLeft: 5,
		},
	}

	totalDamage := 0
	takeDamage := func(amount int) {
		totalDamage += amount
	}

	UpdateEffects(&effects, 2.5, takeDamage)
	if totalDamage != 4 {
		t.Fatalf("expected two burn ticks for 4 damage, got %d", totalDamage)
	}
	if math.Abs(float64(effects[0].TickTimer-0.5)) > 0.0001 {
		t.Fatalf("expected burn tick remainder 0.5, got %.2f", effects[0].TickTimer)
	}

	UpdateEffects(&effects, 0.6, takeDamage)
	if totalDamage != 6 {
		t.Fatalf("expected third burn tick after remainder, got %d", totalDamage)
	}
	if math.Abs(float64(effects[0].TickTimer-0.1)) > 0.0001 {
		t.Fatalf("expected burn tick remainder 0.1, got %.2f", effects[0].TickTimer)
	}
}

func TestEffectQueryHelpers(t *testing.T) {
	effects := []EffectInstance{}
	if !CanAct(&effects) {
		t.Fatalf("expected empty effects to allow acting")
	}
	if !CanCast(&effects) {
		t.Fatalf("expected empty effects to allow casting")
	}
	if HasCrowdControl(&effects) {
		t.Fatalf("expected empty effects to have no crowd control")
	}
	if MoveSpeedMultiplier(&effects) != 1 {
		t.Fatalf("expected neutral move speed multiplier 1")
	}

	ApplyEffect(&effects, Effect{Type: EffectSilence, Duration: 2})
	if !CanAct(&effects) {
		t.Fatalf("expected silence to keep actions available")
	}
	if CanCast(&effects) {
		t.Fatalf("expected silence to block casts")
	}

	ApplyEffect(&effects, Effect{Type: EffectStun, Duration: 1})
	if CanAct(&effects) {
		t.Fatalf("expected stun to block actions")
	}
	if CanCast(&effects) {
		t.Fatalf("expected stun to block casts")
	}

	effects = []EffectInstance{}
	ApplyEffect(&effects, Effect{Type: EffectSlow, Duration: 4, Magnitude: 0.2})
	ApplyEffect(&effects, Effect{Type: EffectMoveSpeedReduction, Duration: 4, Magnitude: 0.25})
	ApplyEffect(&effects, Effect{Type: EffectMoveSpeedBoost, Duration: 4, Magnitude: 0.1})
	multiplier := MoveSpeedMultiplier(&effects)
	if math.Abs(float64(multiplier-0.66)) > 0.0001 {
		t.Fatalf("expected combined move multiplier 0.66, got %.4f", multiplier)
	}

	ApplyEffect(&effects, Effect{Type: EffectFreeze, Duration: 1})
	if MoveSpeedMultiplier(&effects) != 0 {
		t.Fatalf("expected freeze to force zero move speed multiplier")
	}
}
