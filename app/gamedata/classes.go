package gamedata

type ClassType int

const (
	ClassTypeMelee ClassType = iota
	ClassTypeRanged
	ClassTypeCaster
)

type StatType int

const (
	StatTypeSTR StatType = iota
	StatTypeAGI
	StatTypeVIT
	StatTypeINT
	StatTypeDEX
	StatTypeLUK
)

type Class struct {
	Type             ClassType
	Name             string
	PrimaryStat      StatType
	BaselineStats    Stats
	GrowthBias       StatType
	AttackRange      float32
	LifestealPercent float32
	KillHealAmount   int
	ManaCost         int
	ManaToHealthRate float32
}

var classTable = map[ClassType]Class{
	ClassTypeMelee: {
		Type:             ClassTypeMelee,
		Name:             "Warrior",
		PrimaryStat:      StatTypeSTR,
		BaselineStats:    Stats{STR: 8, AGI: 5, VIT: 7, INT: 2, DEX: 4, LUK: 4},
		GrowthBias:       StatTypeSTR,
		AttackRange:      50,
		LifestealPercent: 0.2,
	},
	ClassTypeRanged: {
		Type:           ClassTypeRanged,
		Name:           "Ranger",
		PrimaryStat:    StatTypeDEX,
		BaselineStats:  Stats{STR: 3, AGI: 8, VIT: 5, INT: 3, DEX: 8, LUK: 6},
		GrowthBias:     StatTypeDEX,
		AttackRange:    200,
		KillHealAmount: 20,
	},
	ClassTypeCaster: {
		Type:             ClassTypeCaster,
		Name:             "Mage",
		PrimaryStat:      StatTypeINT,
		BaselineStats:    Stats{STR: 2, AGI: 4, VIT: 6, INT: 9, DEX: 6, LUK: 4},
		GrowthBias:       StatTypeINT,
		AttackRange:      150,
		ManaCost:         10,
		ManaToHealthRate: 2.0,
	},
}

func GetClass(classType ClassType) *Class {
	class, ok := classTable[classType]
	if !ok {
		defaultClass := classTable[ClassTypeMelee]
		return &defaultClass
	}
	return &class
}

func GetClassTable() map[ClassType]Class {
	table := make(map[ClassType]Class, len(classTable))
	for classType, class := range classTable {
		table[classType] = class
	}
	return table
}

func (s StatType) String() string {
	switch s {
	case StatTypeSTR:
		return "STR"
	case StatTypeAGI:
		return "AGI"
	case StatTypeVIT:
		return "VIT"
	case StatTypeINT:
		return "INT"
	case StatTypeDEX:
		return "DEX"
	case StatTypeLUK:
		return "LUK"
	default:
		return "UNKNOWN"
	}
}

func (c Class) BaseStats() *Stats {
	base := c.BaselineStats
	return &base
}
