package gamedata

import "math/rand"

type ItemSlot int

const (
	ItemSlotWeapon ItemSlot = iota
	ItemSlotHead
	ItemSlotChest
	ItemSlotLower
)

type Item struct {
	Name             string
	Description      string
	Slot             ItemSlot
	StatBonuses      map[StatType]int
	ClassRestriction ClassType
}

func NewItem(name, desc string, slot ItemSlot, bonuses map[StatType]int, classRestriction ClassType) *Item {
	return &Item{
		Name:             name,
		Description:      desc,
		Slot:             slot,
		StatBonuses:      bonuses,
		ClassRestriction: classRestriction,
	}
}

func GetWeaponPool(classType ClassType) []*Item {
	pool := []*Item{}

	switch classType {
	case ClassTypeMelee:
		pool = append(pool, NewItem("Iron Sword", "A basic sword", ItemSlotWeapon, map[StatType]int{StatTypeSTR: 3}, ClassTypeMelee))
		pool = append(pool, NewItem("Steel Blade", "A stronger blade", ItemSlotWeapon, map[StatType]int{StatTypeSTR: 5, StatTypeAGI: 1}, ClassTypeMelee))
		pool = append(pool, NewItem("War Axe", "Heavy but powerful", ItemSlotWeapon, map[StatType]int{StatTypeSTR: 7}, ClassTypeMelee))
	case ClassTypeRanged:
		pool = append(pool, NewItem("Wooden Bow", "A basic bow", ItemSlotWeapon, map[StatType]int{StatTypeDEX: 3}, ClassTypeRanged))
		pool = append(pool, NewItem("Hunting Bow", "Improved accuracy", ItemSlotWeapon, map[StatType]int{StatTypeDEX: 5, StatTypeAGI: 1}, ClassTypeRanged))
		pool = append(pool, NewItem("Crossbow", "Powerful ranged weapon", ItemSlotWeapon, map[StatType]int{StatTypeDEX: 7}, ClassTypeRanged))
	case ClassTypeCaster:
		pool = append(pool, NewItem("Apprentice Staff", "Basic magic focus", ItemSlotWeapon, map[StatType]int{StatTypeINT: 3}, ClassTypeCaster))
		pool = append(pool, NewItem("Mage Staff", "Enhanced magic", ItemSlotWeapon, map[StatType]int{StatTypeINT: 5, StatTypeVIT: 1}, ClassTypeCaster))
		pool = append(pool, NewItem("Archmage Rod", "Powerful magic focus", ItemSlotWeapon, map[StatType]int{StatTypeINT: 7}, ClassTypeCaster))
	}

	return pool
}

func GetArmorPool(slot ItemSlot) []*Item {
	pool := []*Item{}

	switch slot {
	case ItemSlotHead:
		pool = append(pool, NewItem("Leather Cap", "Basic head protection", ItemSlotHead, map[StatType]int{StatTypeVIT: 2}, ClassTypeMelee))
		pool = append(pool, NewItem("Iron Helmet", "Sturdy helmet", ItemSlotHead, map[StatType]int{StatTypeVIT: 3, StatTypeSTR: 1}, ClassTypeMelee))
		pool = append(pool, NewItem("Mage Hat", "Magic-enhancing hat", ItemSlotHead, map[StatType]int{StatTypeINT: 2, StatTypeVIT: 1}, ClassTypeCaster))
		pool = append(pool, NewItem("Ranger Hood", "Light headgear", ItemSlotHead, map[StatType]int{StatTypeDEX: 2, StatTypeAGI: 1}, ClassTypeRanged))
	case ItemSlotChest:
		pool = append(pool, NewItem("Leather Armor", "Basic protection", ItemSlotChest, map[StatType]int{StatTypeVIT: 3}, ClassTypeMelee))
		pool = append(pool, NewItem("Chain Mail", "Better defense", ItemSlotChest, map[StatType]int{StatTypeVIT: 4, StatTypeSTR: 1}, ClassTypeMelee))
		pool = append(pool, NewItem("Robe", "Magic protection", ItemSlotChest, map[StatType]int{StatTypeINT: 3, StatTypeVIT: 2}, ClassTypeCaster))
		pool = append(pool, NewItem("Ranger Tunic", "Light armor", ItemSlotChest, map[StatType]int{StatTypeDEX: 3, StatTypeAGI: 2}, ClassTypeRanged))
	case ItemSlotLower:
		pool = append(pool, NewItem("Leather Pants", "Basic leg protection", ItemSlotLower, map[StatType]int{StatTypeVIT: 2}, ClassTypeMelee))
		pool = append(pool, NewItem("Iron Greaves", "Heavy leg armor", ItemSlotLower, map[StatType]int{StatTypeVIT: 3}, ClassTypeMelee))
		pool = append(pool, NewItem("Mage Robes", "Magic legwear", ItemSlotLower, map[StatType]int{StatTypeINT: 2, StatTypeAGI: 1}, ClassTypeCaster))
		pool = append(pool, NewItem("Ranger Leggings", "Agile legwear", ItemSlotLower, map[StatType]int{StatTypeDEX: 2, StatTypeAGI: 2}, ClassTypeRanged))
	}

	return pool
}

func GenerateRewardOptions(classType ClassType) []*Item {
	options := []*Item{}

	weaponPool := GetWeaponPool(classType)
	armorPools := []ItemSlot{ItemSlotHead, ItemSlotChest, ItemSlotLower}

	for i := 0; i < 3; i++ {
		if i == 0 {
			options = append(options, weaponPool[rand.Intn(len(weaponPool))])
		} else {
			slot := armorPools[rand.Intn(len(armorPools))]
			armorPool := GetArmorPool(slot)
			classAppropriate := []*Item{}
			for _, item := range armorPool {
				if item.ClassRestriction == classType || item.ClassRestriction == ClassTypeMelee {
					classAppropriate = append(classAppropriate, item)
				}
			}
			if len(classAppropriate) > 0 {
				options = append(options, classAppropriate[rand.Intn(len(classAppropriate))])
			} else {
				options = append(options, armorPool[rand.Intn(len(armorPool))])
			}
		}
	}

	return options
}
