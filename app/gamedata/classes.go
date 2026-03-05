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
		AttackRange:      50,
		LifestealPercent: 0.2,
	},
	ClassTypeRanged: {
		Type:           ClassTypeRanged,
		Name:           "Ranger",
		PrimaryStat:    StatTypeDEX,
		AttackRange:    200,
		KillHealAmount: 20,
	},
	ClassTypeCaster: {
		Type:             ClassTypeCaster,
		Name:             "Mage",
		PrimaryStat:      StatTypeINT,
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
