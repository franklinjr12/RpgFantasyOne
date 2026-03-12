package gamedata

import "strings"

type BossEncounterConfig struct {
	ID                  string
	Biome               string
	MaxHP               int
	Damage              int
	MoveSpeed           float32
	AttackCooldown      float32
	AttackRange         float32
	AggroRange          float32
	Width               float32
	Height              float32
	TargetFightDuration float32
	HeavyAttack         BossHeavyAttackConfig
	AreaDenial          BossAreaDenialConfig
	Enrage              BossEnrageConfig
}

type BossHeavyAttackConfig struct {
	TelegraphDuration float32
	Cooldown          float32
	Radius            float32
	Damage            int
	DamageType        DamageType
}

type BossAreaDenialConfig struct {
	Cooldown        float32
	WarningDuration float32
	ActiveDuration  float32
	TickRate        float32
	Radius          float32
	Damage          int
	DamageType      DamageType
	Effects         []EffectSpec
	ZoneCount       int
	SpawnDistance   float32
}

type BossEnrageConfig struct {
	ThresholdHPPercent      float32
	MoveSpeedMultiplier     float32
	DamageMultiplier        float32
	HeavyCooldownMultiplier float32
	AreaCooldownMultiplier  float32
	ZoneCountBonus          int
}

var defaultBossEncounter = BossEncounterConfig{
	ID:                  "forest_warden_alpha",
	Biome:               "forest",
	MaxHP:               700,
	Damage:              18,
	MoveSpeed:           78,
	AttackCooldown:      1.25,
	AttackRange:         92,
	AggroRange:          1000,
	Width:               72,
	Height:              72,
	TargetFightDuration: 75,
	HeavyAttack: BossHeavyAttackConfig{
		TelegraphDuration: 1.2,
		Cooldown:          5.8,
		Radius:            110,
		Damage:            34,
		DamageType:        DamagePhysical,
	},
	AreaDenial: BossAreaDenialConfig{
		Cooldown:        7.2,
		WarningDuration: 1.1,
		ActiveDuration:  4.2,
		TickRate:        0.8,
		Radius:          88,
		Damage:          8,
		DamageType:      DamageMagical,
		Effects: []EffectSpec{
			{
				Type:      EffectSlow,
				Duration:  1.0,
				Magnitude: 0.2,
			},
		},
		ZoneCount:     2,
		SpawnDistance: 130,
	},
	Enrage: BossEnrageConfig{
		ThresholdHPPercent:      0.5,
		MoveSpeedMultiplier:     1.18,
		DamageMultiplier:        1.22,
		HeavyCooldownMultiplier: 0.72,
		AreaCooldownMultiplier:  0.7,
		ZoneCountBonus:          1,
	},
}

var bossEncountersByBiome = map[string]BossEncounterConfig{
	"forest": defaultBossEncounter,
}

func GetBossEncounterConfig(biome string) BossEncounterConfig {
	key := strings.ToLower(strings.TrimSpace(biome))
	if key == "" {
		key = "forest"
	}
	cfg, ok := bossEncountersByBiome[key]
	if !ok {
		cfg = defaultBossEncounter
	}
	return sanitizeBossEncounterConfig(cfg)
}

func sanitizeBossEncounterConfig(cfg BossEncounterConfig) BossEncounterConfig {
	if strings.TrimSpace(cfg.ID) == "" {
		cfg.ID = defaultBossEncounter.ID
	}
	cfg.Biome = strings.ToLower(strings.TrimSpace(cfg.Biome))
	if cfg.Biome == "" {
		cfg.Biome = defaultBossEncounter.Biome
	}
	if cfg.MaxHP <= 1 {
		cfg.MaxHP = defaultBossEncounter.MaxHP
	}
	if cfg.Damage <= 0 {
		cfg.Damage = defaultBossEncounter.Damage
	}
	if cfg.MoveSpeed <= 0 {
		cfg.MoveSpeed = defaultBossEncounter.MoveSpeed
	}
	if cfg.AttackCooldown <= 0 {
		cfg.AttackCooldown = defaultBossEncounter.AttackCooldown
	}
	if cfg.AttackRange <= 0 {
		cfg.AttackRange = defaultBossEncounter.AttackRange
	}
	if cfg.AggroRange <= 0 {
		cfg.AggroRange = defaultBossEncounter.AggroRange
	}
	if cfg.Width <= 0 {
		cfg.Width = defaultBossEncounter.Width
	}
	if cfg.Height <= 0 {
		cfg.Height = defaultBossEncounter.Height
	}
	if cfg.TargetFightDuration <= 0 {
		cfg.TargetFightDuration = defaultBossEncounter.TargetFightDuration
	}
	if cfg.HeavyAttack.TelegraphDuration <= 0 {
		cfg.HeavyAttack.TelegraphDuration = defaultBossEncounter.HeavyAttack.TelegraphDuration
	}
	if cfg.HeavyAttack.Cooldown <= 0 {
		cfg.HeavyAttack.Cooldown = defaultBossEncounter.HeavyAttack.Cooldown
	}
	if cfg.HeavyAttack.Radius <= 0 {
		cfg.HeavyAttack.Radius = defaultBossEncounter.HeavyAttack.Radius
	}
	if cfg.HeavyAttack.Damage <= 0 {
		cfg.HeavyAttack.Damage = defaultBossEncounter.HeavyAttack.Damage
	}
	if cfg.HeavyAttack.DamageType != DamagePhysical && cfg.HeavyAttack.DamageType != DamageMagical {
		cfg.HeavyAttack.DamageType = defaultBossEncounter.HeavyAttack.DamageType
	}
	if cfg.AreaDenial.Cooldown <= 0 {
		cfg.AreaDenial.Cooldown = defaultBossEncounter.AreaDenial.Cooldown
	}
	if cfg.AreaDenial.WarningDuration <= 0 {
		cfg.AreaDenial.WarningDuration = defaultBossEncounter.AreaDenial.WarningDuration
	}
	if cfg.AreaDenial.ActiveDuration <= 0 {
		cfg.AreaDenial.ActiveDuration = defaultBossEncounter.AreaDenial.ActiveDuration
	}
	if cfg.AreaDenial.TickRate <= 0 {
		cfg.AreaDenial.TickRate = defaultBossEncounter.AreaDenial.TickRate
	}
	if cfg.AreaDenial.Radius <= 0 {
		cfg.AreaDenial.Radius = defaultBossEncounter.AreaDenial.Radius
	}
	if cfg.AreaDenial.Damage <= 0 {
		cfg.AreaDenial.Damage = defaultBossEncounter.AreaDenial.Damage
	}
	if cfg.AreaDenial.DamageType != DamagePhysical && cfg.AreaDenial.DamageType != DamageMagical {
		cfg.AreaDenial.DamageType = defaultBossEncounter.AreaDenial.DamageType
	}
	if cfg.AreaDenial.ZoneCount <= 0 {
		cfg.AreaDenial.ZoneCount = defaultBossEncounter.AreaDenial.ZoneCount
	}
	if cfg.AreaDenial.SpawnDistance < 0 {
		cfg.AreaDenial.SpawnDistance = defaultBossEncounter.AreaDenial.SpawnDistance
	}
	if cfg.Enrage.ThresholdHPPercent <= 0 || cfg.Enrage.ThresholdHPPercent >= 1 {
		cfg.Enrage.ThresholdHPPercent = defaultBossEncounter.Enrage.ThresholdHPPercent
	}
	if cfg.Enrage.MoveSpeedMultiplier <= 0 {
		cfg.Enrage.MoveSpeedMultiplier = defaultBossEncounter.Enrage.MoveSpeedMultiplier
	}
	if cfg.Enrage.DamageMultiplier <= 0 {
		cfg.Enrage.DamageMultiplier = defaultBossEncounter.Enrage.DamageMultiplier
	}
	if cfg.Enrage.HeavyCooldownMultiplier <= 0 {
		cfg.Enrage.HeavyCooldownMultiplier = defaultBossEncounter.Enrage.HeavyCooldownMultiplier
	}
	if cfg.Enrage.AreaCooldownMultiplier <= 0 {
		cfg.Enrage.AreaCooldownMultiplier = defaultBossEncounter.Enrage.AreaCooldownMultiplier
	}
	if cfg.Enrage.ZoneCountBonus < 0 {
		cfg.Enrage.ZoneCountBonus = defaultBossEncounter.Enrage.ZoneCountBonus
	}
	cfg.AreaDenial.Effects = append([]EffectSpec(nil), cfg.AreaDenial.Effects...)
	return cfg
}
