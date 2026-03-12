//go:build raylib
// +build raylib

package game

import (
	"testing"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/settings"
	"singlefantasy/app/world"
)

func TestEnterBossRewardBuildsCuratedUniqueOptionsOnce(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)

	g.EnterBossReward()
	if g.State != StateReward {
		t.Fatalf("expected reward state after boss reward entry")
	}
	if !g.BossRewardTriggered {
		t.Fatalf("expected boss reward trigger flag")
	}
	if len(g.RewardOptions) != 3 {
		t.Fatalf("expected 3 reward options, got %d", len(g.RewardOptions))
	}
	if g.RewardContext != gamedata.RewardContextBoss {
		t.Fatalf("expected boss reward context")
	}

	seen := map[string]struct{}{}
	for i, item := range g.RewardOptions {
		if item == nil {
			t.Fatalf("expected reward option %d to be non-nil", i)
		}
		key := item.Name
		if _, exists := seen[key]; exists {
			t.Fatalf("expected unique reward options, found duplicate %q", key)
		}
		seen[key] = struct{}{}
	}

	first := g.RewardOptions
	g.EnterBossReward()
	if len(g.RewardOptions) != len(first) {
		t.Fatalf("expected second entry to be ignored")
	}
}

func TestDungeonRunSystemBossClearEntersBossReward(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeRanged)
	g.State = StateRun
	g.CurrentRoom = &world.Room{
		Type: world.RoomTypeBoss,
	}
	g.Boss = gameobjects.NewBoss(100, 100, "forest")
	g.Boss.Alive = false

	system := &dungeonRunSystem{}
	system.Update(NewRuntimeContext(g), 0.016)

	if g.State != StateReward {
		t.Fatalf("expected reward state after boss clear, got %s", g.GetStateName())
	}
	if !g.BossRewardTriggered {
		t.Fatalf("expected boss reward trigger flag")
	}
	if len(g.RewardOptions) != 3 {
		t.Fatalf("expected 3 reward options, got %d", len(g.RewardOptions))
	}
	if g.RewardContext != gamedata.RewardContextBoss {
		t.Fatalf("expected boss reward context")
	}
}

func TestBossMechanicsDoNotPersistAcrossRunReset(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	g.Boss = gameobjects.NewBoss(100, 100, "forest")
	g.Boss.AreaCooldownRemaining = 0
	g.Boss.HeavyCooldownRemaining = 999

	g.Boss.Update(0.016, 120, 120)
	if g.Boss.ActiveZoneCount() == 0 {
		t.Fatalf("expected active zones before reset")
	}

	g.ResetState()
	if g.Boss != nil {
		t.Fatalf("expected boss to be cleared by reset")
	}
	if g.BossRewardTriggered {
		t.Fatalf("expected boss reward trigger flag reset")
	}
	if g.MilestoneRewardTriggered {
		t.Fatalf("expected milestone reward trigger flag reset")
	}
	if len(g.RewardHistory) != 0 {
		t.Fatalf("expected reward history to reset")
	}
}

func TestBossHeavyAndAreaDamageRespectPlayerIFrames(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(420, 0, gamedata.ClassTypeMelee)
	g.Boss = gameobjects.NewBoss(0, 0, "forest")
	g.Boss.HeavyCooldownRemaining = 999
	g.Boss.AreaCooldownRemaining = 999

	combat := &combatResolveSystem{}
	ctx := NewRuntimeContext(g)
	playerX, playerY := g.Player.Center()

	g.Boss.HeavyState = gameobjects.BossHeavyAttackTelegraph
	g.Boss.HeavyTelegraph = gameobjects.BossTelegraph{
		X:        playerX,
		Y:        playerY,
		Radius:   g.Boss.Config.HeavyAttack.Radius,
		Duration: 0.05,
		TimeLeft: 0.01,
	}

	startHP := g.Player.HP
	g.Boss.Update(0.02, playerX, playerY)
	combat.Update(ctx, 0)
	if g.Player.HP >= startHP {
		t.Fatalf("expected heavy event to damage player")
	}
	hpAfterFirstHeavy := g.Player.HP

	g.Boss.HeavyState = gameobjects.BossHeavyAttackTelegraph
	g.Boss.HeavyTelegraph = gameobjects.BossTelegraph{
		X:        playerX,
		Y:        playerY,
		Radius:   g.Boss.Config.HeavyAttack.Radius,
		Duration: 0.05,
		TimeLeft: 0.01,
	}
	g.Boss.Update(0.02, playerX, playerY)
	combat.Update(ctx, 0)
	if g.Player.HP != hpAfterFirstHeavy {
		t.Fatalf("expected heavy damage to be ignored during i-frames")
	}

	g.Player.HurtIFrameTimer = 0
	g.Boss.AreaZones = []*gameobjects.BossAreaDenialZone{
		{
			X:               playerX,
			Y:               playerY,
			Radius:          g.Boss.Config.AreaDenial.Radius,
			WarningDuration: 0,
			WarningTimeLeft: 0,
			ActiveDuration:  1,
			ActiveTimeLeft:  1,
			TickRate:        0.1,
			TickTimer:       0.11,
			Damage:          g.Boss.Config.AreaDenial.Damage,
			DamageType:      g.Boss.Config.AreaDenial.DamageType,
			Active:          true,
		},
	}
	g.Boss.Update(0.02, playerX, playerY)
	combat.Update(ctx, 0)
	hpAfterFirstArea := g.Player.HP
	if hpAfterFirstArea >= hpAfterFirstHeavy {
		t.Fatalf("expected area denial to damage player after i-frames reset")
	}

	g.Boss.AreaZones = []*gameobjects.BossAreaDenialZone{
		{
			X:               playerX,
			Y:               playerY,
			Radius:          g.Boss.Config.AreaDenial.Radius,
			WarningDuration: 0,
			WarningTimeLeft: 0,
			ActiveDuration:  1,
			ActiveTimeLeft:  1,
			TickRate:        0.1,
			TickTimer:       0.11,
			Damage:          g.Boss.Config.AreaDenial.Damage,
			DamageType:      g.Boss.Config.AreaDenial.DamageType,
			Active:          true,
		},
	}
	g.Boss.Update(0.02, playerX, playerY)
	combat.Update(ctx, 0)
	if g.Player.HP != hpAfterFirstArea {
		t.Fatalf("expected area denial damage to be ignored during i-frames")
	}
}

func TestDungeonRunSystemRoomFourClearEntersMilestoneReward(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeCaster)
	g.State = StateRun
	g.Dungeon = &world.Dungeon{
		Seed:        55,
		CurrentRoom: 3,
		Rooms: []*world.Room{
			{Type: world.RoomTypeStart},
			{Type: world.RoomTypeCombat},
			{Type: world.RoomTypeCombat},
			{Type: world.RoomTypeCombat, Biome: "forest"},
		},
	}
	g.CurrentRoom = g.Dungeon.Rooms[g.Dungeon.CurrentRoom]
	g.Enemies = []*gameobjects.Enemy{}

	system := &dungeonRunSystem{}
	system.Update(NewRuntimeContext(g), 0.016)

	if g.State != StateReward {
		t.Fatalf("expected reward state after room 4 clear, got %s", g.GetStateName())
	}
	if !g.MilestoneRewardTriggered {
		t.Fatalf("expected milestone reward trigger flag")
	}
	if g.RewardContext != gamedata.RewardContextMilestone {
		t.Fatalf("expected milestone reward context")
	}
	if len(g.RewardOptions) != RewardMilestoneOfferSize {
		t.Fatalf("expected %d milestone reward options, got %d", RewardMilestoneOfferSize, len(g.RewardOptions))
	}
}

func TestMilestoneRewardNotTriggeredTwice(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeRanged)
	g.State = StateRun
	g.MilestoneRewardTriggered = true
	g.Dungeon = &world.Dungeon{
		Seed:        42,
		CurrentRoom: 3,
		Rooms: []*world.Room{
			{Type: world.RoomTypeStart},
			{Type: world.RoomTypeCombat},
			{Type: world.RoomTypeCombat},
			{Type: world.RoomTypeCombat, Biome: "forest"},
		},
	}
	g.CurrentRoom = g.Dungeon.Rooms[g.Dungeon.CurrentRoom]
	g.Enemies = []*gameobjects.Enemy{}

	system := &dungeonRunSystem{}
	system.Update(NewRuntimeContext(g), 0.016)

	if g.State != StateRun {
		t.Fatalf("expected run state when milestone already triggered, got %s", g.GetStateName())
	}
}

func TestConfirmMilestoneRewardReturnsToRunAndEquipsItem(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	g.State = StateReward
	g.RewardContext = gamedata.RewardContextMilestone
	g.SelectedReward = 0
	g.RewardOptions = []*gamedata.Item{
		gamedata.NewCuratedItem(
			"test_milestone_head",
			"Milestone Helm",
			"",
			gamedata.ItemSlotHead,
			map[gamedata.StatType]int{gamedata.StatTypeVIT: 2},
			gamedata.ClassTypeMelee,
			gamedata.ItemMetadata{Biome: "forest", Weight: 10},
		),
	}

	g.confirmRewardSelection()

	if g.State != StateRun {
		t.Fatalf("expected run state after milestone confirm, got %s", g.GetStateName())
	}
	if g.RewardContext != gamedata.RewardContextNone {
		t.Fatalf("expected reward context reset after milestone confirm")
	}
	if len(g.RewardOptions) != 0 {
		t.Fatalf("expected reward options cleared after milestone confirm")
	}
	if len(g.RewardHistory) != 1 {
		t.Fatalf("expected reward history entry recorded, got %d", len(g.RewardHistory))
	}
	if g.RewardHistory[0].Context != gamedata.RewardContextMilestone {
		t.Fatalf("expected milestone context in reward history")
	}
	equipped := g.Player.Equipment[gamedata.ItemSlotHead]
	if equipped == nil || equipped.ID != "test_milestone_head" {
		t.Fatalf("expected selected milestone item equipped in head slot")
	}
}

func TestConfirmBossRewardEndsRunInResults(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeCaster)
	g.State = StateReward
	g.RewardContext = gamedata.RewardContextBoss
	g.SelectedReward = 0
	g.RewardOptions = []*gamedata.Item{
		gamedata.NewCuratedItem(
			"test_boss_weapon",
			"Boss Staff",
			"",
			gamedata.ItemSlotWeapon,
			map[gamedata.StatType]int{gamedata.StatTypeINT: 4},
			gamedata.ClassTypeCaster,
			gamedata.ItemMetadata{Biome: "forest", Weight: 10},
		),
	}

	g.confirmRewardSelection()

	if g.State != StateResults {
		t.Fatalf("expected results state after boss reward confirm, got %s", g.GetStateName())
	}
	if !g.Results.Victory {
		t.Fatalf("expected victory results after boss reward confirm")
	}
	if g.Results.RewardPicked != "Boss Staff" {
		t.Fatalf("expected picked reward name to be recorded, got %q", g.Results.RewardPicked)
	}
	if len(g.RewardHistory) != 1 {
		t.Fatalf("expected reward history entry recorded, got %d", len(g.RewardHistory))
	}
	if g.RewardHistory[0].Context != gamedata.RewardContextBoss {
		t.Fatalf("expected boss context in reward history")
	}
}
