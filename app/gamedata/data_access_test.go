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
}
