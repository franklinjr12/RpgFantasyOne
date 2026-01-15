package gamedata

type Stats struct {
	STR int
	AGI int
	VIT int
	INT int
	DEX int
	LUK int
}

func NewStats() *Stats {
	return &Stats{
		STR: 5,
		AGI: 5,
		VIT: 5,
		INT: 5,
		DEX: 5,
		LUK: 5,
	}
}

func (s *Stats) GetStat(statType StatType) int {
	switch statType {
	case StatTypeSTR:
		return s.STR
	case StatTypeAGI:
		return s.AGI
	case StatTypeVIT:
		return s.VIT
	case StatTypeINT:
		return s.INT
	case StatTypeDEX:
		return s.DEX
	case StatTypeLUK:
		return s.LUK
	default:
		return 0
	}
}

func (s *Stats) AddStat(statType StatType, amount int) {
	switch statType {
	case StatTypeSTR:
		s.STR += amount
	case StatTypeAGI:
		s.AGI += amount
	case StatTypeVIT:
		s.VIT += amount
	case StatTypeINT:
		s.INT += amount
	case StatTypeDEX:
		s.DEX += amount
	case StatTypeLUK:
		s.LUK += amount
	}
}

func (s *Stats) CalculatePhysicalDamage(baseDamage int) int {
	return baseDamage + s.STR*2
}

func (s *Stats) CalculateMagicDamage(baseDamage int) int {
	return baseDamage + s.INT*2
}

func (s *Stats) CalculateRangedDamage(baseDamage int) int {
	return baseDamage + s.DEX*2
}

func (s *Stats) CalculateMaxHealth(baseHealth int) int {
	return baseHealth + s.VIT*10
}

func (s *Stats) CalculateMoveSpeed(baseSpeed float32) float32 {
	return baseSpeed + float32(s.AGI)*5
}

func (s *Stats) CalculateAttackSpeed(baseSpeed float32) float32 {
	return baseSpeed * (1.0 + float32(s.AGI)*0.02)
}

func (s *Stats) CalculateMaxMana(baseMana int) int {
	return baseMana + s.INT*5
}

func (s *Stats) CalculateCritChance() float32 {
	return float32(s.DEX+s.LUK) * 0.5
}

func (s *Stats) CalculateAutoAttackDamage(baseDamage int) int {
	return baseDamage + s.STR*2
}
