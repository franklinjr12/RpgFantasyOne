//go:build raylib
// +build raylib

package game

import (
	"math"
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

func TestSpawnRoomEnemiesBossSpawnsAtRoomCenter(t *testing.T) {
	g := NewGame(settings.Default())
	g.CurrentRoom = &world.Room{
		X:      320,
		Y:      140,
		Width:  800,
		Height: 600,
		Type:   world.RoomTypeBoss,
		Biome:  "forest",
	}

	g.SpawnRoomEnemies()
	if g.Boss == nil {
		t.Fatalf("expected boss to be spawned in boss room")
	}

	bossCenterX, bossCenterY := g.Boss.Center()
	roomCenterX := g.CurrentRoom.X + g.CurrentRoom.Width/2
	roomCenterY := g.CurrentRoom.Y + g.CurrentRoom.Height/2
	if math.Abs(float64(bossCenterX-roomCenterX)) > 0.001 || math.Abs(float64(bossCenterY-roomCenterY)) > 0.001 {
		t.Fatalf("expected boss center (%.2f, %.2f), got (%.2f, %.2f)", roomCenterX, roomCenterY, bossCenterX, bossCenterY)
	}
}

func TestBossMovementChasesInAggroAndWhenProvokedOutOfRange(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(300, 220, gamedata.ClassTypeMelee)
	g.CurrentRoom = &world.Room{
		X:         0,
		Y:         0,
		Width:     1200,
		Height:    800,
		Type:      world.RoomTypeBoss,
		Obstacles: []world.AABB{},
	}
	g.Boss = gameobjects.NewBoss(100, 220, "forest")

	ai := &aiSystem{}
	move := &movementSystem{}
	ctx := NewRuntimeContext(g)

	// Within aggro range: boss should chase and move.
	g.Boss.AggroRange = 500
	g.Boss.PosX = 100
	g.Boss.PosY = 220
	beforeInAggroX := g.Boss.PosX
	ai.Update(ctx, 0.016)
	move.Update(ctx, 0.2)
	if g.Boss.PosX <= beforeInAggroX {
		t.Fatalf("expected boss to move toward player while in aggro range")
	}

	// Outside aggro and not provoked: boss should stay still.
	g.Boss.AggroRange = 30
	g.Boss.Provoked = false
	g.Boss.PosX = 100
	g.Boss.PosY = 220
	g.Player.PosX = 700
	g.Player.PosY = 220
	beforeOutOfAggroX := g.Boss.PosX
	ai.Update(ctx, 0.016)
	move.Update(ctx, 0.2)
	if g.Boss.PosX != beforeOutOfAggroX {
		t.Fatalf("expected boss to stay still while outside aggro range and not provoked")
	}

	// Provoked outside aggro: boss should chase and move.
	g.Boss.TakeDamage(1)
	if !g.Boss.Provoked {
		t.Fatalf("expected boss to be provoked after taking damage")
	}
	beforeProvokedX := g.Boss.PosX
	ai.Update(ctx, 0.016)
	move.Update(ctx, 0.2)
	if g.Boss.PosX <= beforeProvokedX {
		t.Fatalf("expected provoked boss to move toward player outside aggro range")
	}
}

func TestBossMovementCanEscapeInitialObstacleOverlap(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(800, 220, gamedata.ClassTypeMelee)
	g.CurrentRoom = &world.Room{
		X:      0,
		Y:      0,
		Width:  1400,
		Height: 800,
		Type:   world.RoomTypeBoss,
		Obstacles: []world.AABB{
			{
				X:      80,
				Y:      190,
				Width:  120,
				Height: 120,
			},
		},
	}
	g.Boss = gameobjects.NewBoss(100, 220, "forest")
	g.Boss.PosX = 100
	g.Boss.PosY = 220
	g.Boss.AggroRange = 1000

	ai := &aiSystem{}
	move := &movementSystem{}
	ctx := NewRuntimeContext(g)

	beforeX := g.Boss.PosX
	ai.Update(ctx, 0.016)
	move.Update(ctx, 0.2)

	if g.Boss.PosX <= beforeX {
		t.Fatalf("expected boss to move even when starting inside an obstacle overlap")
	}
}

func TestAdvanceToNextRoomKeepsPlayerHPAndProgressesRoomState(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(0, 0, gamedata.ClassTypeMelee)
	g.Player.HP = g.Player.MaxHP - 35

	startRoom := &world.Room{
		Type: world.RoomTypeCombat,
	}
	nextRoomDoor := &world.Door{
		Locked: false,
	}
	nextRoom := &world.Room{
		X:      300,
		Y:      120,
		Width:  500,
		Height: 300,
		Type:   world.RoomTypeCombat,
		Doors:  []*world.Door{nextRoomDoor},
		Enemies: []*world.EnemyRef{
			{
				X:    360,
				Y:    200,
				Type: gamedata.EnemyArchetypeRaider,
			},
		},
	}
	g.Dungeon = &world.Dungeon{
		CurrentRoom: 0,
		Rooms:       []*world.Room{startRoom, nextRoom},
	}
	g.CurrentRoom = startRoom
	g.Projectiles = []*Projectile{{Alive: true}}
	g.EnemyProjectiles = []*EnemyProjectile{{Alive: true}}
	g.DelayedSkillEffects = []*DelayedSkillEffect{{Alive: true}}
	g.SkillVisualEffects = []*SkillVisualEffect{{TimeLeft: 1}}
	g.CombatTextEvents = []*CombatTextEvent{{Text: "hit"}}
	g.DirectionalTelegraphs = []*DirectionalTelegraphEvent{{Duration: 1}}
	g.RoomTransitionTimer = 0.2
	g.PendingRoomTransition = true
	g.BossRewardTriggered = true

	beforeHP := g.Player.HP
	g.AdvanceToNextRoom()

	if g.Dungeon.CurrentRoom != 1 {
		t.Fatalf("expected dungeon current room to advance to 1, got %d", g.Dungeon.CurrentRoom)
	}
	if g.CurrentRoom != nextRoom {
		t.Fatalf("expected current room to advance to next room")
	}
	if g.Player.HP != beforeHP {
		t.Fatalf("expected player hp to persist across room transition (%d), got %d", beforeHP, g.Player.HP)
	}
	if !g.Player.Alive {
		t.Fatalf("expected player to be alive after room transition")
	}

	entryX, entryY := nextRoom.EntryPoint()
	expectedX := entryX - g.Player.Hitbox.Width/2
	expectedY := entryY - g.Player.Hitbox.Height/2
	if g.Player.PosX != expectedX || g.Player.PosY != expectedY {
		t.Fatalf("expected player reposition to entry point (%.2f, %.2f), got (%.2f, %.2f)", expectedX, expectedY, g.Player.PosX, g.Player.PosY)
	}

	if len(g.Enemies) != 1 {
		t.Fatalf("expected one spawned enemy in next room, got %d", len(g.Enemies))
	}
	if len(g.CurrentRoom.Doors) == 0 || !g.CurrentRoom.Doors[0].Locked {
		t.Fatalf("expected next room doors to be locked on entry")
	}
	if len(g.Projectiles) != 0 || len(g.EnemyProjectiles) != 0 || len(g.DelayedSkillEffects) != 0 || len(g.SkillVisualEffects) != 0 || len(g.CombatTextEvents) != 0 || len(g.DirectionalTelegraphs) != 0 {
		t.Fatalf("expected transient combat state to be cleared on room transition")
	}
	if g.RoomTransitionTimer != 0 || g.PendingRoomTransition {
		t.Fatalf("expected transition timer/flag to be reset")
	}
	if g.BossRewardTriggered {
		t.Fatalf("expected boss reward trigger flag reset")
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
