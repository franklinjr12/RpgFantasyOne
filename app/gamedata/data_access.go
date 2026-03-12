package gamedata

func GetClassData(classType ClassType) *Class {
	return GetClass(classType)
}

func GetClassBaseStats(classType ClassType) *Stats {
	class := GetClass(classType)
	if class == nil {
		return NewStats()
	}
	return class.BaseStats()
}

func GetClassGrowthBias(classType ClassType) StatType {
	class := GetClass(classType)
	if class == nil {
		return StatTypeSTR
	}
	return class.GrowthBias
}

func GetSkillData(skillType SkillType) *Skill {
	return NewSkill(skillType)
}

func GetClassSkillData(classType ClassType) []*Skill {
	return GetClassSkills(classType)
}

func GetEnemyData(templateType EnemyTemplateType) EnemyTemplate {
	return GetEnemyTemplate(templateType)
}

func GetEnemyArchetypeData(archetype EnemyArchetypeType) EnemyArchetype {
	return GetEnemyArchetype(archetype)
}

func GetEnemyArchetypePool() []EnemyArchetypeType {
	return EnemyArchetypeTypes()
}

func GetEliteModifierData(modifierType EliteModifierType) EliteModifier {
	return GetEliteModifier(modifierType)
}

func GetEliteModifierPool() []EliteModifierType {
	return EliteModifierTypes()
}

func GetWeaponData(classType ClassType) []*Item {
	return GetWeaponPool(classType)
}

func GetArmorData(slot ItemSlot) []*Item {
	return GetArmorPool(slot)
}

func GetRewardData(classType ClassType) []*Item {
	return rewardPoolForRequest(classType, "forest")
}

func GetRewardPoolData(biome string) []*Item {
	return GetBiomeItemPool(biome)
}

func SelectRewardOptionsData(request RewardSelectionRequest) []*Item {
	return SelectRewardOptions(request)
}

func GetBossEncounterData(biome string) BossEncounterConfig {
	return GetBossEncounterConfig(biome)
}
