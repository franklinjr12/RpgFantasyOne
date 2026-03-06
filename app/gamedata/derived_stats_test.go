package gamedata

import "testing"

func TestComputeEffectiveStatsAppliesEquipmentBonuses(t *testing.T) {
	base := &Stats{STR: 8, AGI: 4, VIT: 6, INT: 2, DEX: 3, LUK: 1}
	equipment := map[ItemSlot]*Item{
		ItemSlotWeapon: NewItem("Test Sword", "", ItemSlotWeapon, map[StatType]int{
			StatTypeSTR: 3,
			StatTypeDEX: 2,
		}, ClassTypeMelee),
		ItemSlotHead: NewItem("Test Cap", "", ItemSlotHead, map[StatType]int{
			StatTypeVIT: 1,
			StatTypeAGI: 2,
		}, ClassTypeMelee),
	}

	effective := ComputeEffectiveStats(base, equipment)

	if effective.STR != 11 || effective.AGI != 6 || effective.VIT != 7 || effective.DEX != 5 {
		t.Fatalf("unexpected effective stats: %+v", effective)
	}
	if base.STR != 8 || base.AGI != 4 || base.VIT != 6 || base.DEX != 3 {
		t.Fatalf("base stats were mutated: %+v", *base)
	}
}

func TestComputeDerivedStatsUsesExistingCoreFormulas(t *testing.T) {
	effective := Stats{STR: 8, AGI: 5, VIT: 7, INT: 2, DEX: 4, LUK: 4}
	derived := ComputeDerivedStats(ClassTypeMelee, effective)

	if derived.MaxHP != 170 {
		t.Fatalf("expected max hp 170, got %d", derived.MaxHP)
	}
	if derived.MaxMana != 60 {
		t.Fatalf("expected max mana 60, got %d", derived.MaxMana)
	}
	if derived.MoveSpeed != 225 {
		t.Fatalf("expected move speed 225, got %.2f", derived.MoveSpeed)
	}
	if derived.AttackSpeedMultiplier != 1.10 {
		t.Fatalf("expected attack speed 1.10, got %.2f", derived.AttackSpeedMultiplier)
	}
	if derived.AutoAttackDamage != 26 {
		t.Fatalf("expected melee auto-attack damage 26, got %d", derived.AutoAttackDamage)
	}
}

func TestComputeDerivedStatsAppliesClassAttackIdentity(t *testing.T) {
	effective := Stats{STR: 12, AGI: 3, VIT: 4, INT: 9, DEX: 11, LUK: 2}

	melee := ComputeDerivedStats(ClassTypeMelee, effective)
	ranged := ComputeDerivedStats(ClassTypeRanged, effective)
	caster := ComputeDerivedStats(ClassTypeCaster, effective)

	if melee.AutoAttackDamage != 34 {
		t.Fatalf("expected melee auto-attack 34, got %d", melee.AutoAttackDamage)
	}
	if ranged.AutoAttackDamage != 32 {
		t.Fatalf("expected ranged auto-attack 32, got %d", ranged.AutoAttackDamage)
	}
	if caster.AutoAttackDamage != 33 {
		t.Fatalf("expected caster auto-attack 33, got %d", caster.AutoAttackDamage)
	}
}

func TestComputeDerivedStatsClampsCritAndResists(t *testing.T) {
	effective := Stats{STR: 1, AGI: 1, VIT: 200, INT: 200, DEX: 200, LUK: 200}
	derived := ComputeDerivedStats(ClassTypeMelee, effective)

	if derived.CritChance != 0.60 {
		t.Fatalf("expected crit chance cap 0.60, got %.2f", derived.CritChance)
	}
	if derived.PhysicalResist != 0.60 {
		t.Fatalf("expected physical resist cap 0.60, got %.2f", derived.PhysicalResist)
	}
	if derived.MagicalResist != 0.60 {
		t.Fatalf("expected magical resist cap 0.60, got %.2f", derived.MagicalResist)
	}
}
