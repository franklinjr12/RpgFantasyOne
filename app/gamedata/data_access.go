package gamedata

func GetClassData(classType ClassType) *Class {
	return GetClass(classType)
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

func GetWeaponData(classType ClassType) []*Item {
	return GetWeaponPool(classType)
}

func GetArmorData(slot ItemSlot) []*Item {
	return GetArmorPool(slot)
}

func GetRewardData(classType ClassType) []*Item {
	return GenerateRewardOptions(classType)
}
