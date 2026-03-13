package gamedata

import "testing"

func TestDataAccessClassSkillEnemy(t *testing.T) {
	class := GetClassData(ClassTypeMelee)
	if class == nil {
		t.Fatalf("expected class data, got nil")
	}
	if class.Name == "" {
		t.Fatalf("expected class name to be populated")
	}
	baseStats := GetClassBaseStats(ClassTypeMelee)
	if baseStats == nil {
		t.Fatalf("expected class base stats, got nil")
	}
	if GetClassGrowthBias(ClassTypeMelee) != StatTypeSTR {
		t.Fatalf("expected melee growth bias STR")
	}

	skill := GetSkillData(SkillTypePowerStrike)
	if skill == nil {
		t.Fatalf("expected skill data, got nil")
	}
	if skill.Name == "" {
		t.Fatalf("expected skill name to be populated")
	}

	enemy := GetEnemyData(EnemyTemplateElite)
	if enemy.MaxHP <= 0 {
		t.Fatalf("expected enemy max HP > 0, got %d", enemy.MaxHP)
	}
	if enemy.Damage <= 0 {
		t.Fatalf("expected enemy damage > 0, got %d", enemy.Damage)
	}

	archetype := GetEnemyArchetypeData(EnemyArchetypeArcher)
	if archetype.Name == "" {
		t.Fatalf("expected enemy archetype name")
	}

	modifier := GetEliteModifierData(EliteModifierScorching)
	if modifier.Name == "" {
		t.Fatalf("expected elite modifier name")
	}

	boss := GetBossEncounterData("forest")
	if boss.ID == "" || boss.MaxHP <= 0 {
		t.Fatalf("expected boss encounter data")
	}

	rewardPool := GetRewardPoolData("forest")
	if len(rewardPool) < 30 {
		t.Fatalf("expected reward pool for forest to contain at least 30 items, got %d", len(rewardPool))
	}

	rewardOptions := SelectRewardOptionsData(RewardSelectionRequest{
		ClassType: ClassTypeMelee,
		Biome:     "forest",
		Context:   RewardContextBoss,
		OfferSize: 3,
		Seed:      99,
	})
	if len(rewardOptions) != 3 {
		t.Fatalf("expected deterministic reward selector to return 3 options, got %d", len(rewardOptions))
	}
}

func TestCasterBasicAttackRangeMatchesMeleeBaseline(t *testing.T) {
	melee := GetClassData(ClassTypeMelee)
	caster := GetClassData(ClassTypeCaster)
	if melee == nil || caster == nil {
		t.Fatalf("expected melee and caster class data")
	}

	if melee.AttackRange != 50 {
		t.Fatalf("expected melee baseline attack range 50, got %.0f", melee.AttackRange)
	}
	if caster.AttackRange != 50 {
		t.Fatalf("expected caster basic attack range 50, got %.0f", caster.AttackRange)
	}
	if caster.AttackRange != melee.AttackRange {
		t.Fatalf("expected caster attack range %.0f to match melee baseline %.0f", caster.AttackRange, melee.AttackRange)
	}
}
