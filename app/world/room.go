package world

import (
	"math"
	"math/rand"

	"singlefantasy/app/gamedata"
)

const (
	RoomMinWidth   = 800
	RoomMaxWidth   = 1200
	RoomMinHeight  = 600
	RoomMaxHeight  = 900
	BossRoomWidth  = 800
	BossRoomHeight = 600

	ObstacleCountMin = 2
	ObstacleCountMax = 4
	DoorWidth        = 60
	DoorHeight       = 140
)

type RoomType int

const (
	RoomTypeNormal RoomType = iota
	RoomTypeBoss
)

type AABB struct {
	X      float32
	Y      float32
	Width  float32
	Height float32
}

func (a AABB) ContainsPoint(x, y float32) bool {
	return x >= a.X && x <= a.X+a.Width && y >= a.Y && y <= a.Y+a.Height
}

type Door struct {
	Bounds          AABB
	Locked          bool
	TargetRoomIndex int
}

type Room struct {
	X                float32
	Y                float32
	Width            float32
	Height           float32
	Type             RoomType
	ProgressionIndex int
	Enemies          []*EnemyRef
	Obstacles        []AABB
	Doors            []*Door
	Completed        bool
}

type EnemyRef struct {
	X             float32
	Y             float32
	Type          gamedata.EnemyArchetypeType
	IsElite       bool
	EliteModifier gamedata.EliteModifierType
}

func NewRoom(x, y float32, roomType RoomType, progressionIndex int) *Room {
	rng := newRoomRNG(x, y, roomType)

	var width, height float32
	if roomType == RoomTypeBoss {
		width = BossRoomWidth
		height = BossRoomHeight
	} else {
		width = float32(rng.Intn(RoomMaxWidth-RoomMinWidth+1) + RoomMinWidth)
		height = float32(rng.Intn(RoomMaxHeight-RoomMinHeight+1) + RoomMinHeight)
	}

	room := &Room{
		X:                x,
		Y:                y,
		Width:            width,
		Height:           height,
		Type:             roomType,
		ProgressionIndex: progressionIndex,
		Enemies:          []*EnemyRef{},
		Obstacles:        []AABB{},
		Doors:            []*Door{},
		Completed:        false,
	}

	if roomType == RoomTypeNormal {
		room.Obstacles = generateObstacles(room, rng)
		room.Enemies = generateEnemies(room, rng, progressionIndex)
	}

	return room
}

func (r *Room) IsBoss() bool {
	return r.Type == RoomTypeBoss
}

func (r *Room) SpawnPoint() (float32, float32) {
	return r.X + r.Width/2, r.Y + r.Height/2
}

func (r *Room) EntryPoint() (float32, float32) {
	return r.X + 80, r.Y + r.Height/2
}

func (r *Room) SetDoorsLocked(locked bool) {
	for _, door := range r.Doors {
		if door == nil {
			continue
		}
		door.Locked = locked
	}
}

func newRoomRNG(x, y float32, roomType RoomType) *rand.Rand {
	seed := int64(math.Round(float64(x*13 + y*17)))
	seed += int64(roomType+1) * 1_000_003
	return rand.New(rand.NewSource(seed))
}

type spawnBlueprint struct {
	Type          gamedata.EnemyArchetypeType
	IsElite       bool
	EliteModifier gamedata.EliteModifierType
}

func generateEnemies(room *Room, rng *rand.Rand, progressionIndex int) []*EnemyRef {
	composition := buildCompositionForRoom(rng, progressionIndex)
	result := make([]*EnemyRef, 0, len(composition))

	spawnPadding := float32(40)
	centerSafe := AABB{
		X:      room.X + room.Width/2 - 90,
		Y:      room.Y + room.Height/2 - 90,
		Width:  180,
		Height: 180,
	}
	placed := make([]AABB, 0, len(composition))

	for i, blueprint := range composition {
		archetype := gamedata.GetEnemyArchetype(blueprint.Type)
		candidateWidth := archetype.Width + 8
		candidateHeight := archetype.Height + 8
		didPlace := false
		for attempt := 0; attempt < 20; attempt++ {
			x := room.X + spawnPadding + rng.Float32()*(room.Width-2*spawnPadding-candidateWidth)
			y := room.Y + spawnPadding + rng.Float32()*(room.Height-2*spawnPadding-candidateHeight)
			candidate := AABB{X: x, Y: y, Width: candidateWidth, Height: candidateHeight}
			if overlapsAny(candidate, room.Obstacles) || overlaps(candidate, centerSafe) || overlapsAny(candidate, placed) {
				continue
			}
			result = append(result, &EnemyRef{
				X:             x,
				Y:             y,
				Type:          blueprint.Type,
				IsElite:       blueprint.IsElite,
				EliteModifier: blueprint.EliteModifier,
			})
			placed = append(placed, candidate)
			didPlace = true
			break
		}
		if didPlace {
			continue
		}

		fallbackX := room.X + spawnPadding + float32(i%2)*80
		fallbackY := room.Y + spawnPadding + float32(i/2)*80
		result = append(result, &EnemyRef{
			X:             fallbackX,
			Y:             fallbackY,
			Type:          blueprint.Type,
			IsElite:       blueprint.IsElite,
			EliteModifier: blueprint.EliteModifier,
		})
		placed = append(placed, AABB{X: fallbackX, Y: fallbackY, Width: candidateWidth, Height: candidateHeight})
	}

	return result
}

func buildCompositionForRoom(rng *rand.Rand, progressionIndex int) []spawnBlueprint {
	allowed := allowedArchetypes(progressionIndex)
	if len(allowed) == 0 {
		allowed = []gamedata.EnemyArchetypeType{gamedata.EnemyArchetypeRaider}
	}

	minCount := 3 + progressionIndex/2
	if minCount > 6 {
		minCount = 6
	}
	maxCount := minCount + 2
	if maxCount > 8 {
		maxCount = 8
	}

	threatBudget := 28 + progressionIndex*10
	roster := make([]spawnBlueprint, 0, maxCount)
	usedTypes := map[gamedata.EnemyArchetypeType]int{}

	for len(roster) < maxCount {
		if len(roster) >= minCount && threatBudget <= 0 {
			break
		}

		selected := pickWeightedArchetype(rng, allowed, progressionIndex, usedTypes)
		spec := gamedata.GetEnemyArchetype(selected)

		if spec.ThreatValue > threatBudget && len(roster) >= minCount {
			break
		}

		roster = append(roster, spawnBlueprint{Type: selected})
		usedTypes[selected]++
		threatBudget -= spec.ThreatValue
	}

	ensureMinimumRoleMix(rng, progressionIndex, allowed, roster, usedTypes)
	assignEliteModifiers(rng, progressionIndex, roster)
	return roster
}

func allowedArchetypes(progressionIndex int) []gamedata.EnemyArchetypeType {
	base := []gamedata.EnemyArchetypeType{
		gamedata.EnemyArchetypeRaider,
		gamedata.EnemyArchetypeSwarmling,
		gamedata.EnemyArchetypePikeman,
	}
	if progressionIndex >= 2 {
		base = append(base, gamedata.EnemyArchetypeArcher)
	}
	if progressionIndex >= 3 {
		base = append(base, gamedata.EnemyArchetypeHexCaller)
	}
	if progressionIndex >= 4 {
		base = append(base, gamedata.EnemyArchetypeBrute)
	}
	return base
}

func pickWeightedArchetype(rng *rand.Rand, allowed []gamedata.EnemyArchetypeType, progressionIndex int, used map[gamedata.EnemyArchetypeType]int) gamedata.EnemyArchetypeType {
	weights := make([]int, len(allowed))
	total := 0

	for i, typ := range allowed {
		spec := gamedata.GetEnemyArchetype(typ)
		weight := 100 - spec.ThreatValue
		if weight < 12 {
			weight = 12
		}

		if progressionIndex < 2 && (typ == gamedata.EnemyArchetypeHexCaller || typ == gamedata.EnemyArchetypeBrute) {
			weight /= 2
		}

		if used[typ] == 0 {
			weight += 20
		}

		if typ == gamedata.EnemyArchetypeSwarmling {
			weight += 15
		}

		weights[i] = weight
		total += weight
	}

	if total <= 0 {
		return allowed[0]
	}

	roll := rng.Intn(total)
	acc := 0
	for i, weight := range weights {
		acc += weight
		if roll < acc {
			return allowed[i]
		}
	}
	return allowed[len(allowed)-1]
}

func ensureMinimumRoleMix(rng *rand.Rand, progressionIndex int, allowed []gamedata.EnemyArchetypeType, roster []spawnBlueprint, used map[gamedata.EnemyArchetypeType]int) {
	if progressionIndex < 2 || len(roster) < 2 || len(used) >= 2 {
		return
	}

	firstType := roster[0].Type
	alternatives := make([]gamedata.EnemyArchetypeType, 0, len(allowed))
	for _, typ := range allowed {
		if typ != firstType {
			alternatives = append(alternatives, typ)
		}
	}
	if len(alternatives) == 0 {
		return
	}

	replaceIndex := len(roster) - 1
	roster[replaceIndex].Type = alternatives[rng.Intn(len(alternatives))]
}

func assignEliteModifiers(rng *rand.Rand, progressionIndex int, roster []spawnBlueprint) {
	if len(roster) == 0 {
		return
	}

	baseChance := 0.15 + float32(progressionIndex)*0.1
	if baseChance > 0.65 {
		baseChance = 0.65
	}
	if rng.Float32() > baseChance {
		return
	}

	maxElites := 1
	if progressionIndex >= 4 && len(roster) >= 5 {
		maxElites = 2
	}
	eliteCount := 1
	if maxElites > 1 && rng.Float32() < 0.35 {
		eliteCount = 2
	}

	modifiers := gamedata.EliteModifierTypes()
	for i := 0; i < eliteCount; i++ {
		index := rng.Intn(len(roster))
		for attempts := 0; attempts < len(roster) && roster[index].IsElite; attempts++ {
			index = (index + 1) % len(roster)
		}
		roster[index].IsElite = true
		roster[index].EliteModifier = modifiers[rng.Intn(len(modifiers))]
	}
}

func generateObstacles(room *Room, rng *rand.Rand) []AABB {
	count := rng.Intn(ObstacleCountMax-ObstacleCountMin+1) + ObstacleCountMin
	result := make([]AABB, 0, count)

	margin := float32(90)
	centerSafe := AABB{
		X:      room.X + room.Width/2 - 110,
		Y:      room.Y + room.Height/2 - 110,
		Width:  220,
		Height: 220,
	}

	for len(result) < count {
		placed := false
		for attempt := 0; attempt < 30; attempt++ {
			w := 70 + rng.Float32()*90
			h := 70 + rng.Float32()*90
			maxX := room.X + room.Width - margin - w
			maxY := room.Y + room.Height - margin - h
			if maxX <= room.X+margin || maxY <= room.Y+margin {
				break
			}
			x := room.X + margin + rng.Float32()*(maxX-(room.X+margin))
			y := room.Y + margin + rng.Float32()*(maxY-(room.Y+margin))
			candidate := AABB{X: x, Y: y, Width: w, Height: h}
			if overlaps(candidate, centerSafe) || overlapsAny(candidate, result) {
				continue
			}
			result = append(result, candidate)
			placed = true
			break
		}

		if !placed {
			break
		}
	}

	return result
}

func overlapsAny(candidate AABB, existing []AABB) bool {
	for _, value := range existing {
		if overlaps(candidate, value) {
			return true
		}
	}
	return false
}

func overlaps(a, b AABB) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}
