package game

import (
	"fmt"
	"sort"
	"strings"

	"singlefantasy/app/assets"
	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/settings"
	"singlefantasy/app/systems"
	"singlefantasy/app/world"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type AppState int

const (
	StateBoot AppState = iota
	StateMainMenu
	StateClassSelect
	StateRun
	StateReward
	StateResults
)

type RunResults struct {
	Victory            bool
	RunDurationSeconds float32
	RoomsCleared       int
	TotalRooms         int
	SelectedClass      gamedata.ClassType
	RewardPicked       string
}

type Game struct {
	State                    AppState
	Player                   *gameobjects.Player
	Enemies                  []*gameobjects.Enemy
	Boss                     *gameobjects.Boss
	Dungeon                  *world.Dungeon
	Camera                   *systems.Camera
	Projectiles              []*Projectile
	EnemyProjectiles         []*EnemyProjectile
	DelayedSkillEffects      []*DelayedSkillEffect
	SkillVisualEffects       []*SkillVisualEffect
	CombatTextEvents         []*CombatTextEvent
	DirectionalTelegraphs    []*DirectionalTelegraphEvent
	CombatFeedbackSequence   int
	CurrentRoom              *world.Room
	SelectedClass            gamedata.ClassType
	LevelUpMenu              bool
	RewardOptions            []*gamedata.Item
	SelectedReward           int
	RewardContext            gamedata.RewardContext
	RewardHistory            []gamedata.RewardOfferHistoryEntry
	RewardSeed               int64
	MilestoneRewardTriggered bool
	PlayerMoveTargetX        float32
	PlayerMoveTargetY        float32
	HasPlayerMoveTarget      bool
	PlayerAttackTarget       interface{}
	RoomTransitionTimer      float32
	RoomTransitionDuration   float32
	PendingRoomTransition    bool
	BossRewardTriggered      bool
	DebugOverlayEnabled      bool
	BootCompleted            bool
	LastFrameTime            float32
	LastUpdateSteps          int
	RunElapsed               float32
	Results                  RunResults
	RunPipeline              *RuntimePipeline
	Settings                 settings.Settings
	soundPlayer              func(string)
	soundCooldowns           map[string]float32
}

type Projectile struct {
	X          float32
	Y          float32
	VX         float32
	VY         float32
	Speed      float32
	Damage     int
	Radius     float32
	Lifetime   float32
	Pierce     int
	HitTargets map[interface{}]struct{}
	Alive      bool
	Skill      *gamedata.Skill
	Caster     *gameobjects.Player
	DamageType gamedata.DamageType
}

type EnemyProjectile struct {
	X          float32
	Y          float32
	VX         float32
	VY         float32
	Speed      float32
	Damage     int
	Radius     float32
	Lifetime   float32
	Alive      bool
	DamageType gamedata.DamageType
	Effects    []gamedata.EffectSpec
}

type DelayedSkillEffect struct {
	X            float32
	Y            float32
	Radius       float32
	Delay        float32
	ActiveTime   float32
	TickRate     float32
	TickTimer    float32
	Active       bool
	Alive        bool
	Skill        *gamedata.Skill
	Caster       *gameobjects.Player
	Intent       systems.CastIntent
	LastAppliedX float32
	LastAppliedY float32
}

type SkillVisualEffect struct {
	X        float32
	Y        float32
	Radius   float32
	Duration float32
	TimeLeft float32
	Skill    *gamedata.Skill
	Filled   bool
}

func NewGame(cfg settings.Settings) *Game {
	return &Game{
		State:                    StateBoot,
		Player:                   nil,
		Enemies:                  []*gameobjects.Enemy{},
		Boss:                     nil,
		Dungeon:                  nil,
		Camera:                   systems.NewCamera(),
		Projectiles:              []*Projectile{},
		EnemyProjectiles:         []*EnemyProjectile{},
		DelayedSkillEffects:      []*DelayedSkillEffect{},
		SkillVisualEffects:       []*SkillVisualEffect{},
		CombatTextEvents:         []*CombatTextEvent{},
		DirectionalTelegraphs:    []*DirectionalTelegraphEvent{},
		CombatFeedbackSequence:   0,
		CurrentRoom:              nil,
		SelectedClass:            gamedata.ClassTypeMelee,
		LevelUpMenu:              false,
		RewardOptions:            []*gamedata.Item{},
		SelectedReward:           0,
		RewardContext:            gamedata.RewardContextNone,
		RewardHistory:            []gamedata.RewardOfferHistoryEntry{},
		RewardSeed:               world.DefaultDungeonSeed,
		MilestoneRewardTriggered: false,
		PlayerMoveTargetX:        0,
		PlayerMoveTargetY:        0,
		HasPlayerMoveTarget:      false,
		PlayerAttackTarget:       nil,
		RoomTransitionTimer:      0,
		RoomTransitionDuration:   0.25,
		PendingRoomTransition:    false,
		BossRewardTriggered:      false,
		DebugOverlayEnabled:      false,
		BootCompleted:            false,
		LastFrameTime:            0,
		LastUpdateSteps:          0,
		RunElapsed:               0,
		Results:                  RunResults{},
		RunPipeline:              NewRuntimePipeline(),
		Settings:                 cfg,
		soundPlayer:              nil,
		soundCooldowns:           map[string]float32{},
	}
}

func (g *Game) SetFrameDiagnostics(frameTime float32, updateSteps int) {
	g.LastFrameTime = frameTime
	g.LastUpdateSteps = updateSteps
}

func (g *Game) UpdateFrame() {
	if rl.IsKeyPressed(DebugToggleKey) {
		g.DebugOverlayEnabled = !g.DebugOverlayEnabled
	}

	switch g.State {
	case StateBoot:
		g.updateBoot()
	case StateMainMenu:
		g.updateMainMenu()
	case StateClassSelect:
		g.updateClassSelect()
	case StateReward:
		g.updateReward()
	case StateResults:
		g.updateResults()
	}
}

func (g *Game) UpdateFixed(deltaTime float32) {
	if g.State == StateRun {
		g.updateRun(deltaTime)
	}
}

func (g *Game) updateBoot() {
	if !g.BootCompleted {
		manager := assets.Get()
		manager.LoadTexture(
			systems.HumanoidSpriteSheetAssetKey,
			"resources/sprites/Basic Humanoid Sprites 4x.png",
			288,
			288,
			rl.Magenta,
		)
		manager.LoadTexture(
			systems.IconSpriteSheetAssetKey,
			"resources/sprites/raven_fantasy_icons_32x32.png",
			systems.IconSheetWidth,
			systems.IconSheetHeight,
			rl.DarkGray,
		)
		manager.LoadTexture(
			systems.FloorAtlasAssetKey,
			"resources/sprites/atlas_floor-16x16.png",
			256,
			256,
			rl.DarkGray,
		)
		manager.LoadTexture(
			systems.WallsHighAtlasAssetKey,
			"resources/sprites/atlas_walls_high-16x32.png",
			512,
			1024,
			rl.DarkGray,
		)
		manager.LoadFont(assets.FontDefault, "")
		manager.LoadSound("sfx.ui.confirm", "resources/audio/ui_confirm.wav")
		manager.LoadSound(sfxPlayerCast, "resources/sounds/player_cast.wav")
		manager.LoadSound(sfxEnemyCast, "resources/sounds/enemy_cast.wav")
		manager.LoadSound(sfxPlayerHit, "resources/sounds/player_damage_taken.wav")
		manager.LoadSound(sfxEnemyHit, "resources/sounds/enemy_damage_taken.wav")
		manager.LoadSound(sfxPlayerHealing, "resources/sounds/player_healing.mp3")
		manager.LoadSound(sfxPlayerLevelUp, "resources/sounds/level_up.mp3")
		manager.LoadSound(sfxDoorOpen, "resources/sounds/door_open.mp3")
		manager.LoadMusic("music.menu", "resources/audio/menu.ogg")
		g.BootCompleted = true
	}

	g.EnterMainMenu()
}

func (g *Game) EnterMainMenu() {
	g.ResetState()
	g.State = StateMainMenu
}

func (g *Game) EnterClassSelect() {
	g.State = StateClassSelect
}

func (g *Game) DebugLoadRoomTemplate(templateID string) error {
	cfg := world.DefaultDungeonGenerationConfig()
	dungeon, err := world.NewDebugDungeonFromTemplate(templateID, cfg)
	if err != nil {
		return err
	}

	if g.Player == nil {
		g.Player = gameobjects.NewPlayer(0, 0, g.SelectedClass)
	}

	g.Dungeon = dungeon
	g.CurrentRoom = dungeon.GetCurrentRoom()
	if g.CurrentRoom == nil {
		return fmt.Errorf("template %q produced no room", templateID)
	}

	startX := g.CurrentRoom.X + g.CurrentRoom.Width/2
	startY := g.CurrentRoom.Y + g.CurrentRoom.Height/2
	g.Player.PosX = startX - g.Player.Hitbox.Width/2
	g.Player.PosY = startY - g.Player.Hitbox.Height/2
	g.Player.HP = g.Player.MaxHP
	g.Player.Alive = true

	g.Projectiles = []*Projectile{}
	g.EnemyProjectiles = []*EnemyProjectile{}
	g.DelayedSkillEffects = []*DelayedSkillEffect{}
	g.SkillVisualEffects = []*SkillVisualEffect{}
	g.CombatTextEvents = []*CombatTextEvent{}
	g.DirectionalTelegraphs = []*DirectionalTelegraphEvent{}
	g.CombatFeedbackSequence = 0
	g.RoomTransitionTimer = 0
	g.PendingRoomTransition = false
	g.BossRewardTriggered = false
	g.MilestoneRewardTriggered = false
	g.RewardHistory = []gamedata.RewardOfferHistoryEntry{}
	g.RewardContext = gamedata.RewardContextNone
	g.RewardSeed = dungeon.Seed

	g.SpawnRoomEnemies()
	if g.CurrentRoom != nil && !g.CurrentRoom.IsBoss() {
		g.CurrentRoom.SetDoorsLocked(false)
	}

	g.State = StateRun
	return nil
}

func (g *Game) StartRun() {
	g.ResetState()
	g.Dungeon = world.NewDungeon()
	g.CurrentRoom = g.Dungeon.GetCurrentRoom()

	if g.CurrentRoom == nil {
		g.EnterResults(false, "")
		return
	}

	startX := g.CurrentRoom.X + g.CurrentRoom.Width/2
	startY := g.CurrentRoom.Y + g.CurrentRoom.Height/2
	g.Player = gameobjects.NewPlayer(startX, startY, g.SelectedClass)

	g.SpawnRoomEnemies()
	if g.CurrentRoom != nil && !g.CurrentRoom.IsBoss() {
		g.CurrentRoom.SetDoorsLocked(true)
	}
	g.RunElapsed = 0
	g.BossRewardTriggered = false
	g.MilestoneRewardTriggered = false
	g.RewardHistory = []gamedata.RewardOfferHistoryEntry{}
	g.RewardContext = gamedata.RewardContextNone
	if g.Dungeon != nil {
		g.RewardSeed = g.Dungeon.Seed
	} else {
		g.RewardSeed = world.DefaultDungeonSeed
	}
	g.State = StateRun
}

func (g *Game) EnterReward() {
	g.openReward(gamedata.RewardContextBoss)
}

func (g *Game) EnterBossReward() {
	if g.BossRewardTriggered {
		return
	}
	g.BossRewardTriggered = true
	g.openReward(gamedata.RewardContextBoss)
}

func (g *Game) EnterMilestoneReward() {
	if g.MilestoneRewardTriggered {
		return
	}
	g.MilestoneRewardTriggered = true
	g.openReward(gamedata.RewardContextMilestone)
}

func (g *Game) openReward(context gamedata.RewardContext) {
	if g.Player == nil {
		g.EnterResults(false, "")
		return
	}

	offerSize := RewardBossOfferSize
	if context == gamedata.RewardContextMilestone {
		offerSize = RewardMilestoneOfferSize
	}

	request := gamedata.RewardSelectionRequest{
		ClassType: g.Player.Class.Type,
		Biome:     g.rewardBiome(),
		Context:   context,
		OfferSize: offerSize,
		Seed:      g.RewardSeed + int64(len(g.RewardHistory)*31),
		History:   append([]gamedata.RewardOfferHistoryEntry{}, g.RewardHistory...),
	}

	g.RewardOptions = gamedata.SelectRewardOptionsData(request)
	if len(g.RewardOptions) == 0 {
		g.RewardOptions = gamedata.SelectRewardOptionsData(gamedata.RewardSelectionRequest{
			ClassType: g.Player.Class.Type,
			Biome:     "forest",
			Context:   context,
			OfferSize: offerSize,
			Seed:      g.RewardSeed + 777,
		})
	}
	if len(g.RewardOptions) == 0 {
		g.EnterResults(false, "")
		return
	}

	g.RewardContext = context
	g.SelectedReward = 0
	g.State = StateReward
}

func (g *Game) rewardBiome() string {
	if g.CurrentRoom != nil && strings.TrimSpace(g.CurrentRoom.Biome) != "" {
		return g.CurrentRoom.Biome
	}
	if g.Dungeon != nil {
		for _, room := range g.Dungeon.Rooms {
			if room == nil || strings.TrimSpace(room.Biome) == "" {
				continue
			}
			return room.Biome
		}
	}
	return "forest"
}

func (g *Game) shouldTriggerMilestoneReward() bool {
	if g == nil || g.MilestoneRewardTriggered || g.Dungeon == nil {
		return false
	}
	return g.Dungeon.CurrentRoom+1 == RewardMilestoneRoomIndex
}

func (g *Game) EnterResults(victory bool, rewardPicked string) {
	totalRooms := 0
	roomsCleared := 0
	if g.Dungeon != nil {
		totalRooms = len(g.Dungeon.Rooms)
		for _, room := range g.Dungeon.Rooms {
			if room.Completed {
				roomsCleared++
			}
		}
	}

	g.Results = RunResults{
		Victory:            victory,
		RunDurationSeconds: g.RunElapsed,
		RoomsCleared:       roomsCleared,
		TotalRooms:         totalRooms,
		SelectedClass:      g.SelectedClass,
		RewardPicked:       rewardPicked,
	}
	g.RewardOptions = []*gamedata.Item{}
	g.SelectedReward = 0
	g.RewardContext = gamedata.RewardContextNone
	g.State = StateResults
}

func (g *Game) ResetState() {
	g.Player = nil
	g.Enemies = []*gameobjects.Enemy{}
	g.Boss = nil
	g.Dungeon = nil
	g.Projectiles = []*Projectile{}
	g.EnemyProjectiles = []*EnemyProjectile{}
	g.DelayedSkillEffects = []*DelayedSkillEffect{}
	g.SkillVisualEffects = []*SkillVisualEffect{}
	g.CombatTextEvents = []*CombatTextEvent{}
	g.DirectionalTelegraphs = []*DirectionalTelegraphEvent{}
	g.CombatFeedbackSequence = 0
	g.CurrentRoom = nil
	g.Camera = systems.NewCamera()
	g.LevelUpMenu = false
	g.RewardOptions = []*gamedata.Item{}
	g.SelectedReward = 0
	g.RewardContext = gamedata.RewardContextNone
	g.RewardHistory = []gamedata.RewardOfferHistoryEntry{}
	g.RewardSeed = world.DefaultDungeonSeed
	g.MilestoneRewardTriggered = false
	g.PlayerMoveTargetX = 0
	g.PlayerMoveTargetY = 0
	g.HasPlayerMoveTarget = false
	g.PlayerAttackTarget = nil
	g.RoomTransitionTimer = 0
	g.PendingRoomTransition = false
	g.BossRewardTriggered = false
	g.RunElapsed = 0
	g.soundCooldowns = map[string]float32{}
}

func (g *Game) SpawnRoomEnemies() {
	if g.CurrentRoom == nil {
		return
	}

	g.Enemies = []*gameobjects.Enemy{}
	g.Boss = nil

	if g.CurrentRoom.IsBoss() {
		bossX, bossY := g.CurrentRoom.SpawnPoint()
		g.Boss = gameobjects.NewBoss(bossX, bossY, g.CurrentRoom.Biome)
	} else {
		for _, enemyRef := range g.CurrentRoom.Enemies {
			enemy := gameobjects.NewEnemyFromArchetype(enemyRef.X, enemyRef.Y, enemyRef.Type, enemyRef.IsElite, enemyRef.EliteModifier)
			g.Enemies = append(g.Enemies, enemy)
		}
		if g.CurrentRoom.Type == world.RoomTypeEvent && g.CurrentRoom.EventDuration > 0 {
			g.CurrentRoom.EventTimeLeft = g.CurrentRoom.EventDuration
		}
	}
}

func (g *Game) CheckRoomCompletion() bool {
	if g.CurrentRoom == nil {
		return false
	}

	if g.CurrentRoom.IsBoss() {
		if g.Boss != nil && !g.Boss.Alive {
			g.CurrentRoom.Completed = true
			return true
		}
		return false
	}

	if g.CurrentRoom.Type == world.RoomTypeEvent {
		if g.CurrentRoom.EventTimeLeft <= 0 {
			g.CurrentRoom.Completed = true
			return true
		}
		return false
	}

	aliveEnemies := 0
	aliveElites := 0
	for _, enemy := range g.Enemies {
		if enemy.Alive {
			aliveEnemies++
			if enemy.IsElite {
				aliveElites++
			}
		}
	}

	if g.CurrentRoom.Type == world.RoomTypeElite && aliveElites > 0 {
		return false
	}
	if aliveEnemies > 0 {
		return false
	}

	g.CurrentRoom.Completed = true
	return true
}

func (g *Game) AdvanceToNextRoom() {
	if g.Dungeon == nil {
		return
	}

	g.Dungeon.CurrentRoom++
	if g.Dungeon.CurrentRoom >= len(g.Dungeon.Rooms) {
		g.EnterReward()
		return
	}

	g.CurrentRoom = g.Dungeon.GetCurrentRoom()
	if g.CurrentRoom == nil || g.Player == nil {
		return
	}

	entryX, entryY := g.CurrentRoom.EntryPoint()
	startX := entryX - g.Player.Hitbox.Width/2
	startY := entryY - g.Player.Hitbox.Height/2
	g.Player.PosX = startX
	g.Player.PosY = startY
	g.Player.HP = g.Player.MaxHP
	g.Player.Alive = true

	g.SpawnRoomEnemies()
	if !g.CurrentRoom.IsBoss() {
		g.CurrentRoom.SetDoorsLocked(true)
	}
	g.Projectiles = []*Projectile{}
	g.EnemyProjectiles = []*EnemyProjectile{}
	g.DelayedSkillEffects = []*DelayedSkillEffect{}
	g.SkillVisualEffects = []*SkillVisualEffect{}
	g.CombatTextEvents = []*CombatTextEvent{}
	g.DirectionalTelegraphs = []*DirectionalTelegraphEvent{}
	g.CombatFeedbackSequence = 0
	g.RoomTransitionTimer = 0
	g.PendingRoomTransition = false
	g.BossRewardTriggered = false
}

func (g *Game) IsMenuOpen() bool {
	return g.LevelUpMenu
}

func (g *Game) updateMainMenu() {
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		g.EnterClassSelect()
	}
}

func (g *Game) updateClassSelect() {
	if rl.IsKeyPressed(rl.KeyOne) {
		g.SelectedClass = gamedata.ClassTypeMelee
	}
	if rl.IsKeyPressed(rl.KeyTwo) {
		g.SelectedClass = gamedata.ClassTypeRanged
	}
	if rl.IsKeyPressed(rl.KeyThree) {
		g.SelectedClass = gamedata.ClassTypeCaster
	}
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		g.StartRun()
	}
}

func (g *Game) updateReward() {
	if rl.IsKeyPressed(rl.KeyOne) {
		g.SelectedReward = 0
	}
	if len(g.RewardOptions) > 1 && rl.IsKeyPressed(rl.KeyTwo) {
		g.SelectedReward = 1
	}
	if len(g.RewardOptions) > 2 && rl.IsKeyPressed(rl.KeyThree) {
		g.SelectedReward = 2
	}

	if !rl.IsKeyPressed(rl.KeyEnter) {
		return
	}

	g.confirmRewardSelection()
}

func (g *Game) confirmRewardSelection() {
	if g == nil {
		return
	}

	rewardPicked := "None"
	if g.Player != nil && g.SelectedReward >= 0 && g.SelectedReward < len(g.RewardOptions) {
		item := g.RewardOptions[g.SelectedReward]
		if item != nil {
			g.Player.EquipItem(item)
			rewardPicked = item.Name
		}
	}

	g.RewardHistory = append(g.RewardHistory, gamedata.BuildRewardHistoryEntry(g.RewardContext, g.RewardOptions))

	if g.RewardContext == gamedata.RewardContextMilestone {
		g.RewardOptions = []*gamedata.Item{}
		g.SelectedReward = 0
		g.RewardContext = gamedata.RewardContextNone
		g.State = StateRun
		return
	}

	g.EnterResults(true, rewardPicked)
}

func (g *Game) updateResults() {
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeySpace) {
		g.EnterMainMenu()
	}
}

func (g *Game) updateRun(deltaTime float32) {
	if g.Player == nil {
		return
	}

	if g.RunPipeline == nil {
		g.RunPipeline = NewRuntimePipeline()
	}
	g.RunPipeline.Update(NewRuntimeContext(g), deltaTime)
}

func (g *Game) Draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.RayWhite)

	switch g.State {
	case StateBoot:
		g.drawBootScreen()
	case StateMainMenu:
		g.drawMainMenu()
	case StateClassSelect:
		g.drawClassSelect()
	case StateRun:
		g.drawRun()
	case StateReward:
		g.drawRewardSelection()
	case StateResults:
		g.drawResults()
	}

	if g.DebugOverlayEnabled {
		systems.DrawDebugOverlay(g.GetDebugLines())
	}

	rl.EndDrawing()
}

func (g *Game) drawBootScreen() {
	rl.DrawText("Booting...", WindowWidth/2-80, WindowHeight/2-20, 40, rl.Black)
}

func (g *Game) drawMainMenu() {
	rl.DrawText("Single Fantasy", WindowWidth/2-140, WindowHeight/2-120, 48, rl.Black)
	rl.DrawText("Press ENTER or SPACE to Start", WindowWidth/2-170, WindowHeight/2-20, 24, rl.DarkGray)
	rl.DrawText("F3 toggles debug overlay", WindowWidth/2-130, WindowHeight/2+20, 20, rl.Gray)
}

func (g *Game) drawClassSelect() {
	rl.DrawText("Select Class", WindowWidth/2-110, WindowHeight/2-140, 40, rl.Black)
	classNames := []string{"1) Warrior", "2) Ranger", "3) Mage"}
	for i, name := range classNames {
		color := rl.Black
		if int(g.SelectedClass) == i {
			color = rl.Blue
		}
		rl.DrawText(name, WindowWidth/2-80, WindowHeight/2-60+int32(i*35), 26, color)
	}

	selectedClass := gamedata.GetClassData(g.SelectedClass)
	if selectedClass != nil {
		base := selectedClass.BaselineStats
		baselineText := fmt.Sprintf("Base Stats: STR %d  AGI %d  VIT %d  INT %d  DEX %d  LUK %d", base.STR, base.AGI, base.VIT, base.INT, base.DEX, base.LUK)
		growthText := fmt.Sprintf("Growth Bias: +%d %s per level", gamedata.LevelUpGrowthStatPoints, selectedClass.GrowthBias.String())
		rl.DrawText(baselineText, WindowWidth/2-360, WindowHeight/2+20, 20, rl.DarkGray)
		rl.DrawText(growthText, WindowWidth/2-360, WindowHeight/2+50, 20, rl.DarkGray)
	}

	g.drawClassSkillPreview(g.SelectedClass)
	rl.DrawText("Press ENTER or SPACE to Confirm", WindowWidth/2-180, WindowHeight/2+275, 22, rl.DarkGray)
}

func (g *Game) drawRun() {
	if g.Dungeon != nil {
		for _, room := range g.Dungeon.Rooms {
			systems.DrawRoom(room, g.Camera)
		}
	}

	for _, delayed := range g.DelayedSkillEffects {
		if delayed == nil || !delayed.Alive {
			continue
		}
		if delayed.Active {
			systems.DrawActiveSkillZone(delayed.X, delayed.Y, delayed.Radius, delayed.Skill, g.Camera)
		} else {
			systems.DrawDelayedTelegraph(delayed.X, delayed.Y, delayed.Radius, delayed.Skill, g.Camera)
		}
	}

	for _, visual := range g.SkillVisualEffects {
		if visual == nil || visual.TimeLeft <= 0 || visual.Duration <= 0 {
			continue
		}
		systems.DrawSkillCastPulse(visual.X, visual.Y, visual.Radius, visual.TimeLeft/visual.Duration, visual.Skill, visual.Filled, g.Camera)
	}

	queue := make([]systems.RenderQueueItem, 0, len(g.Enemies)+len(g.Projectiles)+len(g.EnemyProjectiles)+4)
	stableID := 0

	for _, proj := range g.Projectiles {
		if !proj.Alive {
			continue
		}
		depthY, depthX := systems.DepthSortKey(proj.X, proj.Y)
		projectile := proj
		queue = append(queue, systems.RenderQueueItem{
			DepthY:   depthY,
			DepthX:   depthX,
			StableID: stableID,
			Draw: func() {
				systems.DrawSkillProjectile(projectile.X, projectile.Y, projectile.Radius, projectile.Skill, g.Camera)
			},
		})
		stableID++
	}

	for _, enemy := range g.Enemies {
		if !enemy.IsAlive() {
			continue
		}
		depthY, depthX := systems.DepthSortKey(enemy.PosX+enemy.Hitbox.Width/2, enemy.PosY+enemy.Hitbox.Height)
		targetEnemy := enemy
		queue = append(queue, systems.RenderQueueItem{
			DepthY:   depthY,
			DepthX:   depthX,
			StableID: stableID,
			Draw: func() {
				systems.DrawEnemy(targetEnemy, g.Camera)
			},
		})
		stableID++
	}

	for _, proj := range g.EnemyProjectiles {
		if !proj.Alive {
			continue
		}
		depthY, depthX := systems.DepthSortKey(proj.X, proj.Y)
		enemyProjectile := proj
		queue = append(queue, systems.RenderQueueItem{
			DepthY:   depthY,
			DepthX:   depthX,
			StableID: stableID,
			Draw: func() {
				systems.DrawEnemyProjectile(enemyProjectile.X, enemyProjectile.Y, enemyProjectile.Radius, g.Camera)
			},
		})
		stableID++
	}

	if g.Boss != nil && g.Boss.IsAlive() {
		depthY, depthX := systems.DepthSortKey(g.Boss.PosX+g.Boss.Hitbox.Width/2, g.Boss.PosY+g.Boss.Hitbox.Height)
		boss := g.Boss
		queue = append(queue, systems.RenderQueueItem{
			DepthY:   depthY,
			DepthX:   depthX,
			StableID: stableID,
			Draw: func() {
				systems.DrawBoss(boss, g.Camera)
			},
		})
		stableID++

		for _, proj := range g.Boss.Projectiles {
			if !proj.Alive {
				continue
			}
			depthY, depthX := systems.DepthSortKey(proj.X, proj.Y)
			bossProj := proj
			queue = append(queue, systems.RenderQueueItem{
				DepthY:   depthY,
				DepthX:   depthX,
				StableID: stableID,
				Draw: func() {
					systems.DrawBossProjectile(bossProj.X, bossProj.Y, bossProj.Radius, g.Camera)
				},
			})
			stableID++
		}
	}

	if g.Player != nil {
		depthY, depthX := systems.DepthSortKey(g.Player.PosX+g.Player.Hitbox.Width/2, g.Player.PosY+g.Player.Hitbox.Height)
		queue = append(queue, systems.RenderQueueItem{
			DepthY:   depthY,
			DepthX:   depthX,
			StableID: stableID,
			Draw: func() {
				systems.DrawPlayer(g.Player, g.Camera)
			},
		})
		stableID++
	}

	systems.SortRenderQueue(queue)
	for _, item := range queue {
		if item.Draw != nil {
			item.Draw()
		}
	}

	g.drawCombatFeedback()

	if g.Player != nil {
		systems.DrawSkillBar(g.Player, g.Settings.SkillLabels())
	}

	g.drawRunHUD()

	if g.LevelUpMenu {
		rl.DrawRectangle(WindowWidth/2-200, WindowHeight/2-150, 400, 300, rl.NewColor(0, 0, 0, 200))
		rl.DrawText("Level Up! Allocate Stat Points", WindowWidth/2-180, WindowHeight/2-120, 24, rl.White)
		rl.DrawText(fmt.Sprintf("Points: %d", g.Player.StatPoints), WindowWidth/2-180, WindowHeight/2-90, 20, rl.White)
		rl.DrawText(
			fmt.Sprintf("Growth Bias: +%d %s/level", gamedata.LevelUpGrowthStatPoints, g.Player.Class.GrowthBias.String()),
			WindowWidth/2-180,
			WindowHeight/2-75,
			16,
			rl.LightGray,
		)
		stats := []string{"1: STR", "2: AGI", "3: VIT", "4: INT", "5: DEX", "6: LUK"}
		statValues := []int{g.Player.Stats.STR, g.Player.Stats.AGI, g.Player.Stats.VIT, g.Player.Stats.INT, g.Player.Stats.DEX, g.Player.Stats.LUK}
		for i, stat := range stats {
			text := fmt.Sprintf("%s: %d", stat, statValues[i])
			rl.DrawText(text, WindowWidth/2-180, WindowHeight/2-45+int32(i*25), 18, rl.White)
		}
	}

	if g.RoomTransitionTimer > 0 && g.RoomTransitionDuration > 0 {
		alphaRatio := g.RoomTransitionTimer / g.RoomTransitionDuration
		if alphaRatio < 0 {
			alphaRatio = 0
		}
		if alphaRatio > 1 {
			alphaRatio = 1
		}
		alpha := uint8(alphaRatio * 210)
		rl.DrawRectangle(0, 0, WindowWidth, WindowHeight, rl.NewColor(0, 0, 0, alpha))
	}
}

func (g *Game) drawRewardSelection() {
	rl.DrawRectangle(WindowWidth/2-360, WindowHeight/2-240, 720, 500, rl.NewColor(0, 0, 0, 220))
	title := "Select Boss Reward"
	if g.RewardContext == gamedata.RewardContextMilestone {
		title = "Select Milestone Reward"
	}
	rl.DrawText(title, WindowWidth/2-170, WindowHeight/2-220, 30, rl.White)

	for i, item := range g.RewardOptions {
		if item == nil {
			continue
		}

		y := WindowHeight/2 - 165 + int32(i*130)
		color := rl.White
		if g.SelectedReward == i {
			color = rl.Yellow
			rl.DrawRectangle(WindowWidth/2-350, y-8, 700, 120, rl.NewColor(255, 255, 0, 45))
		}

		g.drawRewardItemIcon(item, float32(WindowWidth/2-338), float32(y+4), g.SelectedReward == i)
		rl.DrawText(fmt.Sprintf("%d: [%s] %s", i+1, item.Slot.String(), item.Name), WindowWidth/2-284, y, 22, color)
		rl.DrawText(item.Description, WindowWidth/2-284, y+24, 18, rl.Gray)

		statLine := formatItemBonusLine(item.StatBonuses)
		rl.DrawText(statLine, WindowWidth/2-284, y+48, 17, rl.LightGray)

		if len(item.Effects) > 0 {
			effectText := gamedata.DescribeItemEffect(item.Effects[0])
			if effectText != "" {
				rl.DrawText("Effect: "+effectText, WindowWidth/2-284, y+70, 16, rl.NewColor(180, 210, 255, 255))
			}
		}

		var equipped *gamedata.Item
		if g.Player != nil {
			equipped = g.Player.Equipment[item.Slot]
		}
		equippedName := "None"
		if equipped != nil {
			equippedName = equipped.Name
		}
		rl.DrawText("Equipped: "+equippedName, WindowWidth/2+80, y+4, 16, rl.NewColor(200, 200, 200, 255))

		deltas := rewardDeltaLines(item, equipped)
		for deltaIndex, deltaLine := range deltas {
			if deltaIndex >= 3 {
				break
			}
			deltaColor := rl.NewColor(140, 220, 140, 255)
			if strings.Contains(deltaLine, "-") {
				deltaColor = rl.NewColor(235, 125, 125, 255)
			}
			rl.DrawText(deltaLine, WindowWidth/2+80, y+26+int32(deltaIndex*18), 16, deltaColor)
		}
	}

	maxOption := len(g.RewardOptions)
	rl.DrawText(fmt.Sprintf("Press ENTER to confirm, 1-%d to choose", maxOption), WindowWidth/2-190, WindowHeight/2+226, 18, rl.White)
}

func formatItemBonusLine(bonuses map[gamedata.StatType]int) string {
	if len(bonuses) == 0 {
		return "Bonuses: none"
	}
	orderedStats := []gamedata.StatType{
		gamedata.StatTypeSTR,
		gamedata.StatTypeAGI,
		gamedata.StatTypeVIT,
		gamedata.StatTypeINT,
		gamedata.StatTypeDEX,
		gamedata.StatTypeLUK,
	}
	parts := make([]string, 0, len(orderedStats))
	for _, statType := range orderedStats {
		value, exists := bonuses[statType]
		if !exists || value == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s %+d", statType.String(), value))
	}
	if len(parts) == 0 {
		return "Bonuses: none"
	}
	return "Bonuses: " + strings.Join(parts, ", ")
}

func rewardDeltaLines(offered, equipped *gamedata.Item) []string {
	orderedStats := []gamedata.StatType{
		gamedata.StatTypeSTR,
		gamedata.StatTypeAGI,
		gamedata.StatTypeVIT,
		gamedata.StatTypeINT,
		gamedata.StatTypeDEX,
		gamedata.StatTypeLUK,
	}
	lines := make([]string, 0, len(orderedStats))
	for _, statType := range orderedStats {
		offeredValue := 0
		equippedValue := 0
		if offered != nil {
			offeredValue = offered.StatBonuses[statType]
		}
		if equipped != nil {
			equippedValue = equipped.StatBonuses[statType]
		}
		delta := offeredValue - equippedValue
		if delta == 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s %+d", statType.String(), delta))
	}
	if len(lines) == 0 {
		return []string{"No stat change"}
	}
	return lines
}

func (g *Game) drawResults() {
	title := "Run Complete"
	titleColor := rl.DarkGray
	if g.Results.Victory {
		title = "Victory!"
		titleColor = rl.Green
	} else {
		title = "Defeat!"
		titleColor = rl.Red
	}

	rl.DrawText(title, WindowWidth/2-90, WindowHeight/2-180, 48, titleColor)

	className := "Warrior"
	switch g.Results.SelectedClass {
	case gamedata.ClassTypeRanged:
		className = "Ranger"
	case gamedata.ClassTypeCaster:
		className = "Mage"
	}

	rl.DrawText(fmt.Sprintf("Class: %s", className), WindowWidth/2-150, WindowHeight/2-90, 26, rl.Black)
	rl.DrawText(fmt.Sprintf("Rooms Cleared: %d/%d", g.Results.RoomsCleared, g.Results.TotalRooms), WindowWidth/2-150, WindowHeight/2-55, 26, rl.Black)
	rl.DrawText(fmt.Sprintf("Run Time: %.1fs", g.Results.RunDurationSeconds), WindowWidth/2-150, WindowHeight/2-20, 26, rl.Black)
	rl.DrawText(fmt.Sprintf("Reward Picked: %s", g.Results.RewardPicked), WindowWidth/2-150, WindowHeight/2+15, 26, rl.Black)
	rl.DrawText("Press ENTER or SPACE to return to Main Menu", WindowWidth/2-230, WindowHeight/2+90, 24, rl.DarkGray)
}

func (g *Game) GetStateName() string {
	switch g.State {
	case StateBoot:
		return "Boot"
	case StateMainMenu:
		return "MainMenu"
	case StateClassSelect:
		return "ClassSelect"
	case StateRun:
		return "Run"
	case StateReward:
		return "Reward"
	case StateResults:
		return "Results"
	default:
		return "Unknown"
	}
}

func (g *Game) GetDebugLines() []string {
	lines := []string{
		fmt.Sprintf("FPS: %d", rl.GetFPS()),
		fmt.Sprintf("State: %s", g.GetStateName()),
		fmt.Sprintf("Fixed updates/frame: %d", g.LastUpdateSteps),
		fmt.Sprintf("Frame time (clamped): %.3f s", g.LastFrameTime),
	}

	if g.State == StateRun {
		roomIndex := 0
		totalRooms := 0
		if g.Dungeon != nil {
			roomIndex = g.Dungeon.CurrentRoom + 1
			totalRooms = len(g.Dungeon.Rooms)
		}

		aliveEnemies := 0
		compositionCounts := map[string]int{}
		eliteCounts := map[string]int{}
		for _, enemy := range g.Enemies {
			if enemy.Alive {
				aliveEnemies++
				compositionCounts[enemy.Name]++
				if enemy.IsElite {
					modifier := enemy.EliteModifierName
					if modifier == "" {
						modifier = "Elite"
					}
					eliteCounts[modifier]++
				}
			}
		}

		bossProjectiles := 0
		if g.Boss != nil {
			for _, proj := range g.Boss.Projectiles {
				if proj.Alive {
					bossProjectiles++
				}
			}
		}

		activeProjectiles := 0
		for _, proj := range g.Projectiles {
			if proj.Alive {
				activeProjectiles++
			}
		}

		lines = append(lines, fmt.Sprintf("Room: %d/%d", roomIndex, totalRooms))
		if g.CurrentRoom != nil {
			lines = append(lines, fmt.Sprintf("Template: %s", g.CurrentRoom.TemplateID))
			lines = append(lines, fmt.Sprintf("Room type: %s (rot %d)", g.CurrentRoom.Type.String(), g.CurrentRoom.Rotation))
			if g.CurrentRoom.Type == world.RoomTypeEvent {
				lines = append(lines, fmt.Sprintf("Event timer: %.1fs", g.CurrentRoom.EventTimeLeft))
			}
		}
		lines = append(lines, fmt.Sprintf("Enemies (alive/total): %d/%d", aliveEnemies, len(g.Enemies)))
		if len(compositionCounts) > 0 {
			compositionParts := make([]string, 0, len(compositionCounts))
			for name, count := range compositionCounts {
				compositionParts = append(compositionParts, fmt.Sprintf("%s:%d", name, count))
			}
			sort.Strings(compositionParts)
			lines = append(lines, fmt.Sprintf("Enemy types: %s", strings.Join(compositionParts, ", ")))
		}
		if len(eliteCounts) > 0 {
			eliteParts := make([]string, 0, len(eliteCounts))
			for modifier, count := range eliteCounts {
				eliteParts = append(eliteParts, fmt.Sprintf("%s:%d", modifier, count))
			}
			sort.Strings(eliteParts)
			lines = append(lines, fmt.Sprintf("Elites: %s", strings.Join(eliteParts, ", ")))
		}
		enemyProjectiles := 0
		for _, proj := range g.EnemyProjectiles {
			if proj.Alive {
				enemyProjectiles++
			}
		}

		lines = append(lines, fmt.Sprintf("Projectiles (player/enemy/boss): %d/%d/%d", activeProjectiles, enemyProjectiles, bossProjectiles))
		if g.Boss != nil && g.Boss.IsAlive() {
			lines = append(lines, fmt.Sprintf("Boss phase: %s", g.Boss.Phase.String()))
			lines = append(lines, fmt.Sprintf("Boss heavy: %s (%.2fs)", g.Boss.HeavyState.String(), g.Boss.HeavyTimeRemaining()))
			lines = append(lines, fmt.Sprintf("Boss zones active: %d", g.Boss.ActiveZoneCount()))
		}
		activeDelayed := 0
		for _, delayed := range g.DelayedSkillEffects {
			if delayed != nil && delayed.Alive {
				activeDelayed++
			}
		}
		lines = append(lines, fmt.Sprintf("Delayed skill effects: %d", activeDelayed))
		if g.RunPipeline != nil {
			lines = append(lines, fmt.Sprintf("Pipeline: %s", g.RunPipeline.OrderString()))
		}
	}

	return lines
}

func (g *Game) GetPlayerMoveSpeed() float32 {
	if g.Player == nil {
		return 0
	}

	return g.Player.MoveSpeed * gamedata.MoveSpeedMultiplier(&g.Player.Effects)
}
