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
	Type           ClassType
	Name           string
	PrimaryStat    StatType
	AttackRange    float32
	LifestealPercent float32
	KillHealAmount  int
	ManaCost        int
	ManaToHealthRate float32
}

func GetClass(classType ClassType) *Class {
	switch classType {
	case ClassTypeMelee:
		return &Class{
			Type:            ClassTypeMelee,
			Name:            "Warrior",
			PrimaryStat:     StatTypeSTR,
			AttackRange:     50,
			LifestealPercent: 0.2,
		}
	case ClassTypeRanged:
		return &Class{
			Type:           ClassTypeRanged,
			Name:           "Ranger",
			PrimaryStat:    StatTypeDEX,
			AttackRange:    200,
			KillHealAmount: 20,
		}
	case ClassTypeCaster:
		return &Class{
			Type:            ClassTypeCaster,
			Name:            "Mage",
			PrimaryStat:     StatTypeINT,
			AttackRange:     150,
			ManaCost:        10,
			ManaToHealthRate: 2.0,
		}
	default:
		return GetClass(ClassTypeMelee)
	}
}
