package gamedata

import (
	"fmt"
	"sort"
	"strings"
)

type ItemSlot int

const (
	ItemSlotWeapon ItemSlot = iota
	ItemSlotHead
	ItemSlotChest
	ItemSlotLower
)

func (slot ItemSlot) String() string {
	switch slot {
	case ItemSlotWeapon:
		return "Weapon"
	case ItemSlotHead:
		return "Head"
	case ItemSlotChest:
		return "Chest"
	case ItemSlotLower:
		return "Lower"
	default:
		return "Unknown"
	}
}

type ItemEffectType int

const (
	ItemEffectBurnOnHit ItemEffectType = iota
	ItemEffectCritChanceVsSlowed
	ItemEffectLifestealOnHit
	ItemEffectManaOnHit
)

type ItemEffect struct {
	Type      ItemEffectType
	Magnitude float32
	Chance    float32
	Duration  float32
	TickRate  float32
}

type Item struct {
	ID               string
	Name             string
	Description      string
	Slot             ItemSlot
	StatBonuses      map[StatType]int
	ClassRestriction ClassType
	FlavorTags       []ClassType
	Biome            string
	Weight           int
	Effects          []ItemEffect
}

type ItemMetadata struct {
	Biome      string
	Weight     int
	FlavorTags []ClassType
	Effects    []ItemEffect
}

const DefaultRewardSeed int64 = 1337

var biomeItemPools = map[string][]*Item{
	"forest": buildForestItemPool(),
}

func NewItem(name, desc string, slot ItemSlot, bonuses map[StatType]int, classRestriction ClassType) *Item {
	return NewCuratedItem(defaultItemID(name), name, desc, slot, bonuses, classRestriction, ItemMetadata{})
}

func NewCuratedItem(id, name, desc string, slot ItemSlot, bonuses map[StatType]int, classRestriction ClassType, metadata ItemMetadata) *Item {
	biome := normalizeBiome(metadata.Biome)
	if biome == "" {
		biome = "forest"
	}
	weight := metadata.Weight
	if weight <= 0 {
		weight = 10
	}

	flavorTags := copyClassTypes(metadata.FlavorTags)
	if len(flavorTags) == 0 && classRestriction != ClassTypeAny {
		flavorTags = []ClassType{classRestriction}
	}

	itemID := strings.TrimSpace(id)
	if itemID == "" {
		itemID = defaultItemID(name)
	}

	return &Item{
		ID:               itemID,
		Name:             name,
		Description:      desc,
		Slot:             slot,
		StatBonuses:      copyStatBonuses(bonuses),
		ClassRestriction: classRestriction,
		FlavorTags:       flavorTags,
		Biome:            biome,
		Weight:           weight,
		Effects:          copyItemEffects(metadata.Effects),
	}
}

func (item *Item) IsClassAllowed(classType ClassType) bool {
	if item == nil {
		return false
	}
	return item.ClassRestriction == ClassTypeAny || item.ClassRestriction == classType
}

func (item *Item) HasFlavor(classType ClassType) bool {
	if item == nil {
		return false
	}
	for _, tag := range item.FlavorTags {
		if tag == classType {
			return true
		}
	}
	return false
}

func DescribeItemEffect(effect ItemEffect) string {
	switch effect.Type {
	case ItemEffectBurnOnHit:
		chance := effect.Chance
		if chance <= 0 {
			chance = 1
		}
		duration := effect.Duration
		if duration <= 0 {
			duration = 4
		}
		tickRate := effect.TickRate
		if tickRate <= 0 {
			tickRate = 1
		}
		return fmt.Sprintf("%.0f%% burn on hit (%.1f dmg/%.1fs for %.1fs)", chance*100, effect.Magnitude, tickRate, duration)
	case ItemEffectCritChanceVsSlowed:
		return fmt.Sprintf("+%.0f%% crit vs slowed targets", effect.Magnitude*100)
	case ItemEffectLifestealOnHit:
		return fmt.Sprintf("+%.0f%% lifesteal", effect.Magnitude*100)
	case ItemEffectManaOnHit:
		return fmt.Sprintf("+%.0f mana on hit", effect.Magnitude)
	default:
		return ""
	}
}

func GetWeaponPool(classType ClassType) []*Item {
	pool := GetBiomeItemPool("forest")
	out := make([]*Item, 0)
	for _, item := range pool {
		if item == nil || item.Slot != ItemSlotWeapon {
			continue
		}
		if !item.IsClassAllowed(classType) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func GetArmorPool(slot ItemSlot) []*Item {
	pool := GetBiomeItemPool("forest")
	out := make([]*Item, 0)
	for _, item := range pool {
		if item == nil || item.Slot != slot {
			continue
		}
		if slot == ItemSlotWeapon {
			continue
		}
		out = append(out, item)
	}
	return out
}

func GetBiomeItemPool(biome string) []*Item {
	normalized := normalizeBiome(biome)
	pool, ok := biomeItemPools[normalized]
	if !ok {
		pool = biomeItemPools["forest"]
	}
	return cloneItems(pool)
}

func CountBiomeItems(biome string) int {
	return len(GetBiomeItemPool(biome))
}

func CountBiomeFlavorItems(biome string, classType ClassType) int {
	count := 0
	for _, item := range GetBiomeItemPool(biome) {
		if item != nil && item.HasFlavor(classType) {
			count++
		}
	}
	return count
}

func GenerateRewardOptions(classType ClassType) []*Item {
	return SelectRewardOptions(RewardSelectionRequest{
		ClassType: classType,
		Biome:     "forest",
		Context:   RewardContextBoss,
		OfferSize: 3,
		Seed:      DefaultRewardSeed,
	})
}

func buildForestItemPool() []*Item {
	allFlavors := []ClassType{ClassTypeMelee, ClassTypeRanged, ClassTypeCaster}
	return []*Item{
		NewCuratedItem("melee_vanguard_sword", "Vanguard Sword", "Reliable steel edge.", ItemSlotWeapon, map[StatType]int{StatTypeSTR: 4}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 14, FlavorTags: []ClassType{ClassTypeMelee}}),
		NewCuratedItem("melee_bloodletter_axe", "Bloodletter Axe", "Feeds on close combat.", ItemSlotWeapon, map[StatType]int{StatTypeSTR: 5, StatTypeVIT: 1}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 9, FlavorTags: []ClassType{ClassTypeMelee}, Effects: []ItemEffect{{Type: ItemEffectLifestealOnHit, Magnitude: 0.06}}}),
		NewCuratedItem("melee_ember_cleaver", "Ember Cleaver", "Leaves enemies scorched.", ItemSlotWeapon, map[StatType]int{StatTypeSTR: 3, StatTypeAGI: 1}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeMelee}, Effects: []ItemEffect{{Type: ItemEffectBurnOnHit, Magnitude: 3.0, Chance: 0.35, Duration: 4, TickRate: 1}}}),
		NewCuratedItem("melee_bruiser_helm", "Bruiser Helm", "Built to trade blows.", ItemSlotHead, map[StatType]int{StatTypeVIT: 3, StatTypeSTR: 1}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 12, FlavorTags: []ClassType{ClassTypeMelee}}),
		NewCuratedItem("melee_warhorn_helm", "Warhorn Helm", "Sharper finishers on hindered foes.", ItemSlotHead, map[StatType]int{StatTypeSTR: 2, StatTypeLUK: 1}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeMelee}, Effects: []ItemEffect{{Type: ItemEffectCritChanceVsSlowed, Magnitude: 0.08}}}),
		NewCuratedItem("melee_ashguard_cap", "Ashguard Cap", "Heat-worn but stubborn.", ItemSlotHead, map[StatType]int{StatTypeVIT: 2, StatTypeAGI: 1}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 9, FlavorTags: []ClassType{ClassTypeMelee}, Effects: []ItemEffect{{Type: ItemEffectBurnOnHit, Magnitude: 2.5, Chance: 0.25, Duration: 4, TickRate: 1}}}),
		NewCuratedItem("melee_legion_plate", "Legion Plate", "Heavy frontline shell.", ItemSlotChest, map[StatType]int{StatTypeVIT: 4, StatTypeSTR: 2}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 13, FlavorTags: []ClassType{ClassTypeMelee}}),
		NewCuratedItem("melee_oathbound_mail", "Oathbound Mail", "Rewards relentless pressure.", ItemSlotChest, map[StatType]int{StatTypeVIT: 3, StatTypeSTR: 1}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeMelee}, Effects: []ItemEffect{{Type: ItemEffectLifestealOnHit, Magnitude: 0.05}}}),
		NewCuratedItem("melee_crushing_armor", "Crushing Armor", "Punishes controlled targets.", ItemSlotChest, map[StatType]int{StatTypeSTR: 3, StatTypeVIT: 2}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeMelee}, Effects: []ItemEffect{{Type: ItemEffectCritChanceVsSlowed, Magnitude: 0.06}}}),
		NewCuratedItem("melee_ironmarch_greaves", "Ironmarch Greaves", "Stable footing for brawls.", ItemSlotLower, map[StatType]int{StatTypeVIT: 3, StatTypeSTR: 1}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 12, FlavorTags: []ClassType{ClassTypeMelee}}),
		NewCuratedItem("melee_charger_pants", "Charger Pants", "Momentum through contact.", ItemSlotLower, map[StatType]int{StatTypeAGI: 2, StatTypeSTR: 2}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 10, FlavorTags: []ClassType{ClassTypeMelee}}),
		NewCuratedItem("melee_cinder_greaves", "Cinder Greaves", "Kicks leave an ember trail.", ItemSlotLower, map[StatType]int{StatTypeVIT: 2, StatTypeLUK: 2}, ClassTypeMelee, ItemMetadata{Biome: "forest", Weight: 7, FlavorTags: []ClassType{ClassTypeMelee}, Effects: []ItemEffect{{Type: ItemEffectBurnOnHit, Magnitude: 2.2, Chance: 0.2, Duration: 4, TickRate: 1}}}),

		NewCuratedItem("ranged_hunter_bow", "Hunter Bow", "Light and steady draw.", ItemSlotWeapon, map[StatType]int{StatTypeDEX: 4}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 14, FlavorTags: []ClassType{ClassTypeRanged}}),
		NewCuratedItem("ranged_falcon_crossbow", "Falcon Crossbow", "Deadly against slowed prey.", ItemSlotWeapon, map[StatType]int{StatTypeDEX: 5, StatTypeAGI: 1}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 9, FlavorTags: []ClassType{ClassTypeRanged}, Effects: []ItemEffect{{Type: ItemEffectCritChanceVsSlowed, Magnitude: 0.1}}}),
		NewCuratedItem("ranged_venom_bow", "Venom Bow", "Barbs ignite weak spots.", ItemSlotWeapon, map[StatType]int{StatTypeDEX: 3, StatTypeLUK: 2}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeRanged}, Effects: []ItemEffect{{Type: ItemEffectBurnOnHit, Magnitude: 2.8, Chance: 0.3, Duration: 4, TickRate: 1}}}),
		NewCuratedItem("ranged_scout_hood", "Scout Hood", "Clear sight through clutter.", ItemSlotHead, map[StatType]int{StatTypeDEX: 2, StatTypeAGI: 2}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 12, FlavorTags: []ClassType{ClassTypeRanged}}),
		NewCuratedItem("ranged_marksman_mask", "Marksman Mask", "Precision when targets are hindered.", ItemSlotHead, map[StatType]int{StatTypeDEX: 3, StatTypeLUK: 2}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeRanged}, Effects: []ItemEffect{{Type: ItemEffectCritChanceVsSlowed, Magnitude: 0.12}}}),
		NewCuratedItem("ranged_windveil_cap", "Windveil Cap", "Quick resets between shots.", ItemSlotHead, map[StatType]int{StatTypeAGI: 3, StatTypeDEX: 1}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 10, FlavorTags: []ClassType{ClassTypeRanged}}),
		NewCuratedItem("ranged_pathfinder_tunic", "Pathfinder Tunic", "Balanced skirmish kit.", ItemSlotChest, map[StatType]int{StatTypeDEX: 3, StatTypeAGI: 2}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 12, FlavorTags: []ClassType{ClassTypeRanged}}),
		NewCuratedItem("ranged_ambush_vest", "Ambush Vest", "Converts burst into sustain.", ItemSlotChest, map[StatType]int{StatTypeDEX: 2, StatTypeLUK: 2}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeRanged}, Effects: []ItemEffect{{Type: ItemEffectLifestealOnHit, Magnitude: 0.04}}}),
		NewCuratedItem("ranged_briar_coat", "Briar Coat", "Needle traps on impact.", ItemSlotChest, map[StatType]int{StatTypeVIT: 2, StatTypeDEX: 2}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 7, FlavorTags: []ClassType{ClassTypeRanged}, Effects: []ItemEffect{{Type: ItemEffectBurnOnHit, Magnitude: 2.4, Chance: 0.22, Duration: 4, TickRate: 1}}}),
		NewCuratedItem("ranged_trail_leggings", "Trail Leggings", "Mobility under pressure.", ItemSlotLower, map[StatType]int{StatTypeAGI: 3, StatTypeDEX: 2}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 12, FlavorTags: []ClassType{ClassTypeRanged}}),
		NewCuratedItem("ranged_sharpshot_boots", "Sharpshot Boots", "Crit windows on controlled targets.", ItemSlotLower, map[StatType]int{StatTypeDEX: 3, StatTypeLUK: 1}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeRanged}, Effects: []ItemEffect{{Type: ItemEffectCritChanceVsSlowed, Magnitude: 0.08}}}),
		NewCuratedItem("ranged_skirmisher_pants", "Skirmisher Pants", "Restores momentum while firing.", ItemSlotLower, map[StatType]int{StatTypeAGI: 2, StatTypeVIT: 2}, ClassTypeRanged, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeRanged}, Effects: []ItemEffect{{Type: ItemEffectManaOnHit, Magnitude: 2}}}),

		NewCuratedItem("caster_novice_staff_plus", "Novice Staff+", "Focused arcane channel.", ItemSlotWeapon, map[StatType]int{StatTypeINT: 4}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 14, FlavorTags: []ClassType{ClassTypeCaster}}),
		NewCuratedItem("caster_frostfocus_rod", "Frostfocus Rod", "Punishes slowed enemies.", ItemSlotWeapon, map[StatType]int{StatTypeINT: 5, StatTypeDEX: 1}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 9, FlavorTags: []ClassType{ClassTypeCaster}, Effects: []ItemEffect{{Type: ItemEffectCritChanceVsSlowed, Magnitude: 0.1}}}),
		NewCuratedItem("caster_cinder_staff", "Cinder Staff", "Arcane flames linger on hit.", ItemSlotWeapon, map[StatType]int{StatTypeINT: 3, StatTypeVIT: 1}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeCaster}, Effects: []ItemEffect{{Type: ItemEffectBurnOnHit, Magnitude: 3.3, Chance: 0.3, Duration: 4, TickRate: 1}}}),
		NewCuratedItem("caster_arcanist_hat", "Arcanist Hat", "Reliable spell throughput.", ItemSlotHead, map[StatType]int{StatTypeINT: 3, StatTypeVIT: 1}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 12, FlavorTags: []ClassType{ClassTypeCaster}}),
		NewCuratedItem("caster_seer_circlet", "Seer Circlet", "Reads openings in slowed foes.", ItemSlotHead, map[StatType]int{StatTypeINT: 2, StatTypeLUK: 2}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeCaster}, Effects: []ItemEffect{{Type: ItemEffectCritChanceVsSlowed, Magnitude: 0.11}}}),
		NewCuratedItem("caster_ember_veil", "Ember Veil", "Arcane sparks ignite targets.", ItemSlotHead, map[StatType]int{StatTypeINT: 2, StatTypeAGI: 1}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 7, FlavorTags: []ClassType{ClassTypeCaster}, Effects: []ItemEffect{{Type: ItemEffectBurnOnHit, Magnitude: 2.6, Chance: 0.2, Duration: 4, TickRate: 1}}}),
		NewCuratedItem("caster_scholar_robe", "Scholar Robe", "Steady defensive weave.", ItemSlotChest, map[StatType]int{StatTypeINT: 4, StatTypeVIT: 2}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 13, FlavorTags: []ClassType{ClassTypeCaster}}),
		NewCuratedItem("caster_manaweave_robe", "Manaweave Robe", "Returns mana through combat.", ItemSlotChest, map[StatType]int{StatTypeINT: 3, StatTypeVIT: 2}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeCaster}, Effects: []ItemEffect{{Type: ItemEffectManaOnHit, Magnitude: 3}}}),
		NewCuratedItem("caster_occult_cassock", "Occult Cassock", "Leeches power from each hit.", ItemSlotChest, map[StatType]int{StatTypeINT: 3, StatTypeLUK: 2}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeCaster}, Effects: []ItemEffect{{Type: ItemEffectLifestealOnHit, Magnitude: 0.05}}}),
		NewCuratedItem("caster_mystic_slacks", "Mystic Slacks", "Low drag spell movement.", ItemSlotLower, map[StatType]int{StatTypeINT: 3, StatTypeAGI: 2}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 12, FlavorTags: []ClassType{ClassTypeCaster}}),
		NewCuratedItem("caster_ritual_pants", "Ritual Pants", "Sustained casting rhythm.", ItemSlotLower, map[StatType]int{StatTypeINT: 2, StatTypeVIT: 2}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeCaster}, Effects: []ItemEffect{{Type: ItemEffectManaOnHit, Magnitude: 2}}}),
		NewCuratedItem("caster_glacial_legwraps", "Glacial Legwraps", "Critical windows on slowed enemies.", ItemSlotLower, map[StatType]int{StatTypeINT: 2, StatTypeDEX: 2}, ClassTypeCaster, ItemMetadata{Biome: "forest", Weight: 8, FlavorTags: []ClassType{ClassTypeCaster}, Effects: []ItemEffect{{Type: ItemEffectCritChanceVsSlowed, Magnitude: 0.09}}}),

		NewCuratedItem("shared_tempered_bandana", "Tempered Bandana", "Simple utility cloth.", ItemSlotHead, map[StatType]int{StatTypeAGI: 2, StatTypeVIT: 1}, ClassTypeAny, ItemMetadata{Biome: "forest", Weight: 11, FlavorTags: allFlavors}),
		NewCuratedItem("shared_travelers_mail", "Traveler's Mail", "Adaptable plated layer.", ItemSlotChest, map[StatType]int{StatTypeVIT: 2, StatTypeDEX: 1}, ClassTypeAny, ItemMetadata{Biome: "forest", Weight: 11, FlavorTags: allFlavors}),
		NewCuratedItem("shared_reinforced_treads", "Reinforced Treads", "Balanced lower armor.", ItemSlotLower, map[StatType]int{StatTypeAGI: 1, StatTypeDEX: 1, StatTypeVIT: 1}, ClassTypeAny, ItemMetadata{Biome: "forest", Weight: 11, FlavorTags: allFlavors}),
	}
}

func normalizeBiome(biome string) string {
	trimmed := strings.TrimSpace(strings.ToLower(biome))
	if trimmed == "" {
		return "forest"
	}
	return trimmed
}

func defaultItemID(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	if lower == "" {
		return "item"
	}

	builder := strings.Builder{}
	builder.Grow(len(lower))
	lastUnderscore := false
	for _, char := range lower {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			builder.WriteRune(char)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteRune('_')
			lastUnderscore = true
		}
	}
	id := strings.Trim(builder.String(), "_")
	if id == "" {
		return "item"
	}
	return id
}

func cloneItems(items []*Item) []*Item {
	out := make([]*Item, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		copyItem := *item
		copyItem.StatBonuses = copyStatBonuses(item.StatBonuses)
		copyItem.FlavorTags = copyClassTypes(item.FlavorTags)
		copyItem.Effects = copyItemEffects(item.Effects)
		out = append(out, &copyItem)
	}
	return out
}

func copyStatBonuses(bonuses map[StatType]int) map[StatType]int {
	if len(bonuses) == 0 {
		return map[StatType]int{}
	}
	out := make(map[StatType]int, len(bonuses))
	for statType, value := range bonuses {
		out[statType] = value
	}
	return out
}

func copyClassTypes(values []ClassType) []ClassType {
	if len(values) == 0 {
		return nil
	}
	out := make([]ClassType, len(values))
	copy(out, values)
	return out
}

func copyItemEffects(effects []ItemEffect) []ItemEffect {
	if len(effects) == 0 {
		return nil
	}
	out := make([]ItemEffect, len(effects))
	copy(out, effects)
	return out
}

func sortedItemIDs(items []*Item) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		ids = append(ids, item.ID)
	}
	sort.Strings(ids)
	return ids
}
