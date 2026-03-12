package world

import (
	"fmt"
	"math/rand"

	"singlefantasy/app/gamedata"
)

const (
	DefaultDungeonSeed      int64   = 1337
	DefaultRunMinRooms      int     = 8
	DefaultRunMaxRooms      int     = 12
	DefaultRunMinEventRooms int     = 1
	DefaultRunMaxEventRooms int     = 2
	DungeonRoomSpacingX     float32 = 220
	DungeonEventDuration    float32 = 14
)

type Dungeon struct {
	Seed        int64
	Rooms       []*Room
	CurrentRoom int
}

type DungeonGenerationConfig struct {
	Seed            int64
	Biome           string
	RoomsRoot       string
	MinRooms        int
	MaxRooms        int
	MinEventRooms   int
	MaxEventRooms   int
	AllowRotations  bool
	RequiredTags    []string
	PreferredRoomID []string
}

type selectedTemplate struct {
	template  *RoomTemplate
	rotation  int
	entryDoor DoorMarker
	exitDoor  DoorMarker
}

func DefaultDungeonGenerationConfig() DungeonGenerationConfig {
	return DungeonGenerationConfig{
		Seed:           DefaultDungeonSeed,
		Biome:          "forest",
		RoomsRoot:      DefaultRoomsRoot,
		MinRooms:       DefaultRunMinRooms,
		MaxRooms:       DefaultRunMaxRooms,
		MinEventRooms:  DefaultRunMinEventRooms,
		MaxEventRooms:  DefaultRunMaxEventRooms,
		AllowRotations: true,
	}
}

func NewDungeon() *Dungeon {
	dungeon, err := NewDungeonWithConfig(DefaultDungeonGenerationConfig())
	if err != nil {
		panic(err)
	}
	return dungeon
}

func NewDungeonWithConfig(cfg DungeonGenerationConfig) (*Dungeon, error) {
	if cfg.MinRooms < 4 {
		cfg.MinRooms = 4
	}
	if cfg.MaxRooms < cfg.MinRooms {
		cfg.MaxRooms = cfg.MinRooms
	}
	if cfg.MinEventRooms < 0 {
		cfg.MinEventRooms = 0
	}
	if cfg.MaxEventRooms < cfg.MinEventRooms {
		cfg.MaxEventRooms = cfg.MinEventRooms
	}
	if cfg.Biome == "" {
		cfg.Biome = "forest"
	}

	registry, err := LoadRoomTemplateRegistry(cfg.RoomsRoot)
	if err != nil {
		return nil, err
	}

	rng := rand.New(rand.NewSource(cfg.Seed))
	totalRooms := cfg.MinRooms
	if cfg.MaxRooms > cfg.MinRooms {
		totalRooms += rng.Intn(cfg.MaxRooms - cfg.MinRooms + 1)
	}

	plan := buildRoomTypePlan(rng, totalRooms, cfg.MinEventRooms, cfg.MaxEventRooms)
	rooms := make([]*Room, 0, len(plan))
	selected := make([]selectedTemplate, 0, len(plan))

	var prevTemplateID string
	prevExitY := -1
	prevExitDoorWorldY := float32(0)

	for index, roomType := range plan {
		needEntry := index > 0
		needExit := index < len(plan)-1

		minDifficulty, maxDifficulty := difficultyBandForIndex(index, len(plan), roomType)
		picked, err := pickTemplateForSlot(rng, registry, cfg, roomType, minDifficulty, maxDifficulty, needEntry, needExit, prevExitY, prevTemplateID)
		if err != nil {
			return nil, err
		}

		roomX := float32(0)
		roomY := float32(0)
		if len(rooms) > 0 {
			prev := rooms[len(rooms)-1]
			roomX = prev.X + prev.Width + DungeonRoomSpacingX
			offsetY := float32(picked.entryDoor.Y)*RoomTemplateTileSize + RoomTemplateTileSize/2
			roomY = prevExitDoorWorldY - offsetY
		}

		room := buildRoomFromTemplate(picked.template, picked.rotation, roomX, roomY, index, rng)
		if room == nil {
			return nil, fmt.Errorf("failed to instantiate room from template %q", picked.template.ID)
		}

		if len(rooms) > 0 {
			prev := rooms[len(rooms)-1]
			prevSelection := selected[len(selected)-1]
			prevDoor := findDoorByMarker(prev, prevSelection.exitDoor)
			if prevDoor == nil {
				return nil, fmt.Errorf("missing outgoing door on template %q", prev.TemplateID)
			}
			prevDoor.TargetRoomIndex = index
			prevDoor.Locked = true
		}

		rooms = append(rooms, room)
		selected = append(selected, picked)
		prevTemplateID = picked.template.ID
		prevExitY = picked.exitDoor.Y

		if needExit {
			exitDoor := findDoorByMarker(room, picked.exitDoor)
			if exitDoor == nil {
				return nil, fmt.Errorf("missing exit door in room template %q", picked.template.ID)
			}
			prevExitDoorWorldY = exitDoor.Bounds.Y + exitDoor.Bounds.Height/2
		}
	}

	return &Dungeon{
		Seed:        cfg.Seed,
		Rooms:       rooms,
		CurrentRoom: 0,
	}, nil
}

func NewDebugDungeonFromTemplate(templateID string, cfg DungeonGenerationConfig) (*Dungeon, error) {
	if cfg.Biome == "" {
		cfg.Biome = "forest"
	}
	if cfg.RoomsRoot == "" {
		cfg.RoomsRoot = DefaultRoomsRoot
	}

	registry, err := LoadRoomTemplateRegistry(cfg.RoomsRoot)
	if err != nil {
		return nil, err
	}
	template := registry.GetByID(templateID)
	if template == nil {
		return nil, fmt.Errorf("room template %q not found", templateID)
	}

	rng := rand.New(rand.NewSource(cfg.Seed))
	room := buildRoomFromTemplate(template, 0, 0, 0, 0, rng)
	if room == nil {
		return nil, fmt.Errorf("failed to instantiate template %q", templateID)
	}
	return &Dungeon{
		Seed:        cfg.Seed,
		Rooms:       []*Room{room},
		CurrentRoom: 0,
	}, nil
}

func buildRoomTypePlan(rng *rand.Rand, totalRooms, minEventRooms, maxEventRooms int) []RoomType {
	if totalRooms < 4 {
		totalRooms = 4
	}

	plan := make([]RoomType, totalRooms)
	plan[0] = RoomTypeStart
	plan[totalRooms-1] = RoomTypeBoss
	for i := 1; i < totalRooms-1; i++ {
		plan[i] = RoomTypeCombat
	}

	maxEventSlots := totalRooms - 3
	if maxEventSlots < 0 {
		maxEventSlots = 0
	}
	if maxEventRooms > maxEventSlots {
		maxEventRooms = maxEventSlots
	}
	if minEventRooms > maxEventRooms {
		minEventRooms = maxEventRooms
	}
	eventCount := minEventRooms
	if maxEventRooms > minEventRooms {
		eventCount += rng.Intn(maxEventRooms - minEventRooms + 1)
	}

	eventCandidates := make([]int, 0, maxEventSlots)
	for i := 2; i <= totalRooms-2; i++ {
		eventCandidates = append(eventCandidates, i)
	}
	rng.Shuffle(len(eventCandidates), func(i, j int) {
		eventCandidates[i], eventCandidates[j] = eventCandidates[j], eventCandidates[i]
	})
	for i := 0; i < eventCount && i < len(eventCandidates); i++ {
		plan[eventCandidates[i]] = RoomTypeEvent
	}

	eliteCandidates := make([]int, 0, totalRooms-3)
	for i := 2; i <= totalRooms-2; i++ {
		eliteCandidates = append(eliteCandidates, i)
	}
	rng.Shuffle(len(eliteCandidates), func(i, j int) {
		eliteCandidates[i], eliteCandidates[j] = eliteCandidates[j], eliteCandidates[i]
	})
	if len(eliteCandidates) > 0 {
		plan[eliteCandidates[0]] = RoomTypeElite
	}

	return plan
}

func difficultyBandForIndex(index, total int, roomType RoomType) (int, int) {
	if roomType == RoomTypeBoss {
		return 3, 3
	}
	if roomType == RoomTypeStart {
		return 1, 1
	}
	if total <= 1 {
		return 1, 3
	}
	progress := float32(index) / float32(total-1)
	if progress < 0.34 {
		return 1, 1
	}
	if progress < 0.68 {
		return 1, 2
	}
	return 2, 3
}

func pickTemplateForSlot(rng *rand.Rand, registry *RoomTemplateRegistry, cfg DungeonGenerationConfig, roomType RoomType, minDifficulty, maxDifficulty int, needEntry, needExit bool, prevExitY int, disallowID string) (selectedTemplate, error) {
	query := TemplateQuery{
		Biome:         cfg.Biome,
		RoomType:      roomType,
		MinDifficulty: minDifficulty,
		MaxDifficulty: maxDifficulty,
		RequiredTags:  append([]string(nil), cfg.RequiredTags...),
	}
	if needEntry {
		query.RequiredDoors = append(query.RequiredDoors, DoorDirectionWest)
	}
	if needExit {
		query.RequiredDoors = append(query.RequiredDoors, DoorDirectionEast)
	}

	baseCandidates := registry.Query(query)
	if len(baseCandidates) == 0 {
		query.MinDifficulty = 1
		query.MaxDifficulty = 3
		baseCandidates = registry.Query(query)
	}
	result, ok := pickTemplateCandidate(rng, baseCandidates, roomType, needEntry, needExit, prevExitY, disallowID, cfg.AllowRotations)
	if ok {
		return result, nil
	}

	result, ok = pickTemplateCandidate(rng, baseCandidates, roomType, needEntry, needExit, prevExitY, "", cfg.AllowRotations)
	if ok {
		return result, nil
	}

	return selectedTemplate{}, fmt.Errorf("no room template matched slot constraints type=%s difficulty=%d..%d", roomType.String(), minDifficulty, maxDifficulty)
}

type weightedTemplateChoice struct {
	value  selectedTemplate
	weight int
}

func pickTemplateCandidate(rng *rand.Rand, templates []*RoomTemplate, roomType RoomType, needEntry, needExit bool, prevExitY int, disallowID string, allowRotations bool) (selectedTemplate, bool) {
	choices := make([]weightedTemplateChoice, 0, len(templates))
	for _, template := range templates {
		if template == nil {
			continue
		}
		if disallowID != "" && template.ID == disallowID {
			continue
		}

		rotations := []int{0}
		if allowRotations && template.AllowRotation {
			rotations = []int{0, 90, 180, 270}
		}

		for _, rotation := range rotations {
			rotated := template
			if rotation != 0 {
				var err error
				rotated, err = RotateRoomTemplate(template, rotation)
				if err != nil {
					continue
				}
			}

			entry := DoorMarker{}
			exit := DoorMarker{}

			if needEntry {
				marker, ok := pickDoorMarker(rotated.Doors, DoorDirectionWest, prevExitY)
				if !ok {
					continue
				}
				entry = marker
			}

			if needExit {
				preferred := -1
				if needEntry {
					preferred = entry.Y
				}
				marker, ok := pickDoorMarker(rotated.Doors, DoorDirectionEast, preferred)
				if !ok {
					continue
				}
				exit = marker
			}

			if roomType == RoomTypeBoss {
				exit = DoorMarker{}
			}

			choices = append(choices, weightedTemplateChoice{
				value: selectedTemplate{
					template:  rotated,
					rotation:  rotation,
					entryDoor: entry,
					exitDoor:  exit,
				},
				weight: max(1, template.Weight),
			})
		}
	}

	if len(choices) == 0 {
		return selectedTemplate{}, false
	}

	totalWeight := 0
	for _, choice := range choices {
		totalWeight += choice.weight
	}
	if totalWeight <= 0 {
		return choices[0].value, true
	}

	roll := rng.Intn(totalWeight)
	acc := 0
	for _, choice := range choices {
		acc += choice.weight
		if roll < acc {
			return choice.value, true
		}
	}
	return choices[len(choices)-1].value, true
}

func pickDoorMarker(markers []DoorMarker, direction DoorDirection, preferredY int) (DoorMarker, bool) {
	candidates := make([]DoorMarker, 0, len(markers))
	for _, marker := range markers {
		if marker.Direction == direction {
			candidates = append(candidates, marker)
		}
	}
	if len(candidates) == 0 {
		return DoorMarker{}, false
	}
	if preferredY >= 0 {
		for _, marker := range candidates {
			if marker.Y == preferredY {
				return marker, true
			}
		}
	}
	return candidates[0], true
}

func buildRoomFromTemplate(template *RoomTemplate, rotation int, x, y float32, progressionIndex int, rng *rand.Rand) *Room {
	if template == nil {
		return nil
	}

	width := float32(template.Width * RoomTemplateTileSize)
	height := float32(template.Height * RoomTemplateTileSize)
	room := &Room{
		X:                x,
		Y:                y,
		Width:            width,
		Height:           height,
		Type:             template.Type,
		ProgressionIndex: progressionIndex,
		Enemies:          []*EnemyRef{},
		Obstacles:        []AABB{},
		Doors:            []*Door{},
		Tiles:            nil,
		TemplateID:       template.ID,
		Biome:            template.Biome,
		Rotation:         rotation,
		GridWidth:        template.Width,
		GridHeight:       template.Height,
		EventDuration:    0,
		EventTimeLeft:    0,
		BossSpawnX:       0,
		BossSpawnY:       0,
		HasBossSpawn:     false,
		Completed:        false,
	}

	room.Obstacles = templateToObstacles(template, room.X, room.Y)
	room.Doors = templateToDoors(template, room.X, room.Y)
	room.Enemies = buildEnemyRefsFromTemplate(template, room, rng, progressionIndex)
	room.Tiles = cloneTileGrid(template.Tiles)

	if room.Type == RoomTypeEvent {
		room.EventDuration = DungeonEventDuration
		room.EventTimeLeft = room.EventDuration
	}

	for _, marker := range template.SpawnMarkers {
		if marker.Type != SpawnTypeBoss {
			continue
		}
		room.BossSpawnX = room.X + float32(marker.X)*RoomTemplateTileSize + RoomTemplateTileSize/2
		room.BossSpawnY = room.Y + float32(marker.Y)*RoomTemplateTileSize + RoomTemplateTileSize/2
		room.HasBossSpawn = true
		break
	}

	return room
}

func templateToObstacles(template *RoomTemplate, originX, originY float32) []AABB {
	obstacles := make([]AABB, 0, template.Width*template.Height/3)
	for y, row := range template.Tiles {
		for x, tile := range row {
			if tile != TileWall && tile != TileHazard && tile != TileTrap {
				continue
			}
			obstacles = append(obstacles, AABB{
				X:      originX + float32(x)*RoomTemplateTileSize,
				Y:      originY + float32(y)*RoomTemplateTileSize,
				Width:  RoomTemplateTileSize,
				Height: RoomTemplateTileSize,
			})
		}
	}
	return obstacles
}

func cloneTileGrid(input [][]TileType) [][]TileType {
	if len(input) == 0 {
		return nil
	}
	result := make([][]TileType, len(input))
	for y := range input {
		result[y] = make([]TileType, len(input[y]))
		copy(result[y], input[y])
	}
	return result
}

func templateToDoors(template *RoomTemplate, originX, originY float32) []*Door {
	doors := make([]*Door, 0, len(template.Doors))
	for _, marker := range template.Doors {
		doors = append(doors, &Door{
			Bounds: AABB{
				X:      originX + float32(marker.X)*RoomTemplateTileSize,
				Y:      originY + float32(marker.Y)*RoomTemplateTileSize,
				Width:  RoomTemplateTileSize,
				Height: RoomTemplateTileSize,
			},
			Direction:       marker.Direction,
			MarkerX:         marker.X,
			MarkerY:         marker.Y,
			Locked:          true,
			TargetRoomIndex: -1,
		})
	}
	return doors
}

func buildEnemyRefsFromTemplate(template *RoomTemplate, room *Room, rng *rand.Rand, progressionIndex int) []*EnemyRef {
	if room == nil || template == nil || room.IsBoss() {
		return nil
	}

	normalMarkers := make([]SpawnMarker, 0)
	eliteMarkers := make([]SpawnMarker, 0)
	for _, marker := range template.SpawnMarkers {
		switch marker.Type {
		case SpawnTypeNormal:
			normalMarkers = append(normalMarkers, marker)
		case SpawnTypeElite:
			eliteMarkers = append(eliteMarkers, marker)
		}
	}

	totalMarkers := len(normalMarkers) + len(eliteMarkers)
	if totalMarkers == 0 {
		return generateEnemies(room, rng, progressionIndex)
	}

	composition := buildCompositionForRoom(rng, progressionIndex)
	if len(composition) == 0 {
		composition = []spawnBlueprint{{Type: gamedata.EnemyArchetypeRaider}}
	}

	enemies := make([]*EnemyRef, 0, totalMarkers)
	for i, marker := range normalMarkers {
		blueprint := composition[i%len(composition)]
		enemies = append(enemies, enemyFromMarker(room, marker, blueprint))
	}

	allowed := allowedArchetypes(progressionIndex)
	modifiers := gamedata.EliteModifierTypes()
	for _, marker := range eliteMarkers {
		archetype := allowed[rng.Intn(len(allowed))]
		blueprint := spawnBlueprint{
			Type:    archetype,
			IsElite: true,
		}
		if len(modifiers) > 0 {
			blueprint.EliteModifier = modifiers[rng.Intn(len(modifiers))]
		}
		enemies = append(enemies, enemyFromMarker(room, marker, blueprint))
	}

	if room.Type == RoomTypeElite {
		hasElite := false
		for _, enemy := range enemies {
			if enemy != nil && enemy.IsElite {
				hasElite = true
				break
			}
		}
		if !hasElite && len(enemies) > 0 {
			enemies[0].IsElite = true
			if len(modifiers) > 0 {
				enemies[0].EliteModifier = modifiers[rng.Intn(len(modifiers))]
			}
		}
	}

	if progressionIndex >= 2 && len(enemies) >= 2 {
		unique := map[gamedata.EnemyArchetypeType]struct{}{}
		for _, enemy := range enemies {
			if enemy == nil {
				continue
			}
			unique[enemy.Type] = struct{}{}
		}
		if len(unique) < 2 {
			allowed := allowedArchetypes(progressionIndex)
			for _, archetype := range allowed {
				if archetype == enemies[0].Type {
					continue
				}
				enemies[1].Type = archetype
				break
			}
		}
	}

	return enemies
}

func enemyFromMarker(room *Room, marker SpawnMarker, blueprint spawnBlueprint) *EnemyRef {
	spec := gamedata.GetEnemyArchetype(blueprint.Type)
	centerX := room.X + float32(marker.X)*RoomTemplateTileSize + RoomTemplateTileSize/2
	centerY := room.Y + float32(marker.Y)*RoomTemplateTileSize + RoomTemplateTileSize/2
	return &EnemyRef{
		X:             centerX - spec.Width/2,
		Y:             centerY - spec.Height/2,
		Type:          blueprint.Type,
		IsElite:       blueprint.IsElite,
		EliteModifier: blueprint.EliteModifier,
	}
}

func findDoorByMarker(room *Room, marker DoorMarker) *Door {
	if room == nil {
		return nil
	}
	for _, door := range room.Doors {
		if door == nil {
			continue
		}
		if door.MarkerX == marker.X && door.MarkerY == marker.Y && door.Direction == marker.Direction {
			return door
		}
	}
	return nil
}

func (d *Dungeon) GetCurrentRoom() *Room {
	if d.CurrentRoom >= 0 && d.CurrentRoom < len(d.Rooms) {
		return d.Rooms[d.CurrentRoom]
	}
	return nil
}

func (d *Dungeon) GetWorldBounds() (float32, float32) {
	maxX := float32(0)
	maxY := float32(0)
	for _, room := range d.Rooms {
		if room.X+room.Width > maxX {
			maxX = room.X + room.Width
		}
		if room.Y+room.Height > maxY {
			maxY = room.Y + room.Height
		}
	}
	return maxX, maxY
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
