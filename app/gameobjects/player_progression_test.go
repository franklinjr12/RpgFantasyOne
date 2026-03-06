package gameobjects

import (
	"testing"

	"singlefantasy/app/gamedata"
)

func TestNewPlayerUsesClassBaselineStats(t *testing.T) {
	melee := NewPlayer(0, 0, gamedata.ClassTypeMelee)
	ranged := NewPlayer(0, 0, gamedata.ClassTypeRanged)
	caster := NewPlayer(0, 0, gamedata.ClassTypeCaster)

	meleeBase := gamedata.GetClassBaseStats(gamedata.ClassTypeMelee)
	rangedBase := gamedata.GetClassBaseStats(gamedata.ClassTypeRanged)
	casterBase := gamedata.GetClassBaseStats(gamedata.ClassTypeCaster)

	if melee.Stats.STR != meleeBase.STR || melee.Stats.VIT != meleeBase.VIT || melee.Stats.INT != meleeBase.INT {
		t.Fatalf("melee baseline mismatch: got %+v want %+v", *melee.Stats, *meleeBase)
	}
	if ranged.Stats.DEX != rangedBase.DEX || ranged.Stats.AGI != rangedBase.AGI || ranged.Stats.LUK != rangedBase.LUK {
		t.Fatalf("ranged baseline mismatch: got %+v want %+v", *ranged.Stats, *rangedBase)
	}
	if caster.Stats.INT != casterBase.INT || caster.Stats.VIT != casterBase.VIT || caster.Stats.STR != casterBase.STR {
		t.Fatalf("caster baseline mismatch: got %+v want %+v", *caster.Stats, *casterBase)
	}
}

func TestGainXPIgnoresNonPositiveValues(t *testing.T) {
	player := NewPlayer(0, 0, gamedata.ClassTypeMelee)

	player.GainXP(0)
	player.GainXP(-50)

	if player.Level != 1 || player.XP != 0 || player.StatPoints != 0 {
		t.Fatalf("expected unchanged progression state, got level=%d xp=%d points=%d", player.Level, player.XP, player.StatPoints)
	}
}

func TestGainXPAppliesGrowthBiasAndStatPoints(t *testing.T) {
	player := NewPlayer(0, 0, gamedata.ClassTypeMelee)
	startSTR := player.Stats.STR

	player.GainXP(100)

	if player.Level != 2 {
		t.Fatalf("expected level 2, got %d", player.Level)
	}
	if player.XP != 0 {
		t.Fatalf("expected xp carry 0, got %d", player.XP)
	}
	if player.XPToNext != 200 {
		t.Fatalf("expected xp to next 200, got %d", player.XPToNext)
	}
	if player.StatPoints != gamedata.LevelUpStatPoints {
		t.Fatalf("expected stat points %d, got %d", gamedata.LevelUpStatPoints, player.StatPoints)
	}
	if player.Stats.STR != startSTR+gamedata.LevelUpGrowthStatPoints {
		t.Fatalf("expected growth bias STR to increase by %d, got %d -> %d", gamedata.LevelUpGrowthStatPoints, startSTR, player.Stats.STR)
	}
}

func TestGainXPSupportsMultiLevelCarryOver(t *testing.T) {
	player := NewPlayer(0, 0, gamedata.ClassTypeCaster)
	startINT := player.Stats.INT

	player.GainXP(350)

	if player.Level != 3 {
		t.Fatalf("expected level 3, got %d", player.Level)
	}
	if player.XP != 50 {
		t.Fatalf("expected xp carry-over 50, got %d", player.XP)
	}
	if player.XPToNext != 300 {
		t.Fatalf("expected xp to next 300, got %d", player.XPToNext)
	}
	if player.StatPoints != gamedata.LevelUpStatPoints*2 {
		t.Fatalf("expected stat points %d, got %d", gamedata.LevelUpStatPoints*2, player.StatPoints)
	}
	expectedINT := startINT + gamedata.LevelUpGrowthStatPoints*2
	if player.Stats.INT != expectedINT {
		t.Fatalf("expected INT %d, got %d", expectedINT, player.Stats.INT)
	}
}

func TestApplyStatsRecalculatesAfterAllocationAndEquip(t *testing.T) {
	player := NewPlayer(0, 0, gamedata.ClassTypeMelee)
	baseHP := player.MaxHP
	baseDamage := player.GetAutoAttackDamage()

	player.StatPoints = 1
	player.AddStatPoint(gamedata.StatTypeVIT)

	if player.MaxHP != baseHP+10 {
		t.Fatalf("expected hp increase by 10 after VIT point, got %d -> %d", baseHP, player.MaxHP)
	}

	item := gamedata.NewItem("Training Sword", "", gamedata.ItemSlotWeapon, map[gamedata.StatType]int{
		gamedata.StatTypeSTR: 3,
	}, gamedata.ClassTypeMelee)
	player.EquipItem(item)

	if player.GetAutoAttackDamage() != baseDamage+6 {
		t.Fatalf("expected damage increase by 6 after +3 STR item, got %d -> %d", baseDamage, player.GetAutoAttackDamage())
	}
}
