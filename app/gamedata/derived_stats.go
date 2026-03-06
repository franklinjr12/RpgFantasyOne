package gamedata

type DerivedStats struct {
	AutoAttackDamage      int
	AttackSpeedMultiplier float32
	MoveSpeed             float32
	CritChance            float32
	PhysicalResist        float32
	MagicalResist         float32
	MaxHP                 int
	MaxMana               int
}

func ComputeEffectiveStats(base *Stats, equipment map[ItemSlot]*Item) Stats {
	if base == nil {
		return Stats{}
	}

	effective := *base
	for _, item := range equipment {
		if item == nil {
			continue
		}
		for statType, bonus := range item.StatBonuses {
			effective.AddStat(statType, bonus)
		}
	}

	return effective
}

func ComputeDerivedStats(classType ClassType, effective Stats) DerivedStats {
	derived := DerivedStats{
		AttackSpeedMultiplier: effective.CalculateAttackSpeed(1.0),
		MoveSpeed:             effective.CalculateMoveSpeed(BasePlayerMoveSpeed),
		CritChance:            clampFloat32((float32(effective.DEX)+float32(effective.LUK))*0.005, 0.0, 0.60),
		PhysicalResist:        clampFloat32(float32(effective.VIT)*0.01, 0.0, 0.60),
		MagicalResist:         clampFloat32(float32(effective.INT)*0.01, 0.0, 0.60),
		MaxHP:                 effective.CalculateMaxHealth(BasePlayerHP),
		MaxMana:               effective.CalculateMaxMana(BasePlayerMana),
	}

	if derived.AttackSpeedMultiplier < 0.2 {
		derived.AttackSpeedMultiplier = 0.2
	}

	switch classType {
	case ClassTypeRanged:
		derived.AutoAttackDamage = effective.CalculateRangedDamage(BaseRangedAutoAttackDamage)
	case ClassTypeCaster:
		derived.AutoAttackDamage = effective.CalculateMagicDamage(BaseCasterAutoAttackDamage)
	default:
		derived.AutoAttackDamage = effective.CalculatePhysicalDamage(BaseMeleeAutoAttackDamage)
	}

	return derived
}

func clampFloat32(value, minValue, maxValue float32) float32 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
