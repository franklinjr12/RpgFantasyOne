package gamedata

import "testing"

func TestGetBossEncounterConfigReturnsSanitizedDefaults(t *testing.T) {
	cfg := GetBossEncounterConfig("forest")
	if cfg.Biome != "forest" {
		t.Fatalf("expected biome forest, got %q", cfg.Biome)
	}
	if cfg.MaxHP <= 1 || cfg.Damage <= 0 || cfg.MoveSpeed <= 0 {
		t.Fatalf("expected positive boss base stats, got hp=%d damage=%d move=%.2f", cfg.MaxHP, cfg.Damage, cfg.MoveSpeed)
	}
	if cfg.HeavyAttack.TelegraphDuration <= 0 || cfg.HeavyAttack.Cooldown <= 0 || cfg.HeavyAttack.Radius <= 0 || cfg.HeavyAttack.Damage <= 0 {
		t.Fatalf("expected valid heavy attack config: %+v", cfg.HeavyAttack)
	}
	if cfg.AreaDenial.WarningDuration <= 0 || cfg.AreaDenial.ActiveDuration <= 0 || cfg.AreaDenial.TickRate <= 0 || cfg.AreaDenial.Radius <= 0 || cfg.AreaDenial.Damage <= 0 {
		t.Fatalf("expected valid area denial config: %+v", cfg.AreaDenial)
	}
	if cfg.Enrage.ThresholdHPPercent <= 0 || cfg.Enrage.ThresholdHPPercent >= 1 {
		t.Fatalf("invalid enrage threshold: %.2f", cfg.Enrage.ThresholdHPPercent)
	}
	if cfg.Enrage.MoveSpeedMultiplier <= 0 || cfg.Enrage.DamageMultiplier <= 0 {
		t.Fatalf("invalid enrage multipliers: %+v", cfg.Enrage)
	}
}

func TestGetBossEncounterConfigUnknownBiomeFallsBack(t *testing.T) {
	forest := GetBossEncounterConfig("forest")
	unknown := GetBossEncounterConfig("unknown")
	if unknown.ID != forest.ID {
		t.Fatalf("expected unknown biome to fallback to default id %q, got %q", forest.ID, unknown.ID)
	}
}

func TestSanitizeBossEncounterConfigFixesInvalidValues(t *testing.T) {
	cfg := sanitizeBossEncounterConfig(BossEncounterConfig{
		ID:     "",
		Biome:  "   ",
		MaxHP:  0,
		Damage: -1,
		HeavyAttack: BossHeavyAttackConfig{
			TelegraphDuration: 0,
			Cooldown:          0,
			Radius:            -3,
			Damage:            0,
		},
		AreaDenial: BossAreaDenialConfig{
			Cooldown:        0,
			WarningDuration: 0,
			ActiveDuration:  0,
			TickRate:        0,
			Radius:          0,
			Damage:          0,
			ZoneCount:       0,
			SpawnDistance:   -1,
		},
		Enrage: BossEnrageConfig{
			ThresholdHPPercent:      2,
			MoveSpeedMultiplier:     0,
			DamageMultiplier:        0,
			HeavyCooldownMultiplier: 0,
			AreaCooldownMultiplier:  0,
			ZoneCountBonus:          -2,
		},
	})

	if cfg.MaxHP <= 1 || cfg.Damage <= 0 {
		t.Fatalf("expected sanitized base values, got hp=%d damage=%d", cfg.MaxHP, cfg.Damage)
	}
	if cfg.HeavyAttack.TelegraphDuration <= 0 || cfg.HeavyAttack.Cooldown <= 0 || cfg.HeavyAttack.Radius <= 0 || cfg.HeavyAttack.Damage <= 0 {
		t.Fatalf("expected sanitized heavy config: %+v", cfg.HeavyAttack)
	}
	if cfg.AreaDenial.Cooldown <= 0 || cfg.AreaDenial.WarningDuration <= 0 || cfg.AreaDenial.ActiveDuration <= 0 || cfg.AreaDenial.TickRate <= 0 || cfg.AreaDenial.Radius <= 0 || cfg.AreaDenial.Damage <= 0 || cfg.AreaDenial.ZoneCount <= 0 {
		t.Fatalf("expected sanitized area config: %+v", cfg.AreaDenial)
	}
	if cfg.Enrage.ThresholdHPPercent <= 0 || cfg.Enrage.ThresholdHPPercent >= 1 {
		t.Fatalf("expected sanitized enrage threshold, got %.2f", cfg.Enrage.ThresholdHPPercent)
	}
	if cfg.Enrage.ZoneCountBonus < 0 {
		t.Fatalf("expected non-negative zone bonus, got %d", cfg.Enrage.ZoneCountBonus)
	}
}
