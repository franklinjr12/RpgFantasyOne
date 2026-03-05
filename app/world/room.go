package world

import (
	"math"
	"math/rand"
)

const (
	RoomMinWidth      = 800
	RoomMaxWidth      = 1200
	RoomMinHeight     = 600
	RoomMaxHeight     = 900
	BossRoomWidth     = 800
	BossRoomHeight    = 600
	EnemiesPerRoomMin = 3
	EnemiesPerRoomMax = 5

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
	X         float32
	Y         float32
	Width     float32
	Height    float32
	Type      RoomType
	Enemies   []*EnemyRef
	Obstacles []AABB
	Doors     []*Door
	Completed bool
}

type EnemyRef struct {
	X       float32
	Y       float32
	IsElite bool
}

func NewRoom(x, y float32, roomType RoomType) *Room {
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
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Type:      roomType,
		Enemies:   []*EnemyRef{},
		Obstacles: []AABB{},
		Doors:     []*Door{},
		Completed: false,
	}

	if roomType == RoomTypeNormal {
		room.Obstacles = generateObstacles(room, rng)
		room.Enemies = generateEnemies(room, rng)
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

func generateEnemies(room *Room, rng *rand.Rand) []*EnemyRef {
	enemyCount := rng.Intn(EnemiesPerRoomMax-EnemiesPerRoomMin+1) + EnemiesPerRoomMin
	result := make([]*EnemyRef, 0, enemyCount)

	spawnPadding := float32(40)
	enemySize := float32(60)
	centerSafe := AABB{
		X:      room.X + room.Width/2 - 90,
		Y:      room.Y + room.Height/2 - 90,
		Width:  180,
		Height: 180,
	}

	for i := 0; i < enemyCount; i++ {
		placed := false
		for attempt := 0; attempt < 20; attempt++ {
			x := room.X + spawnPadding + rng.Float32()*(room.Width-2*spawnPadding-enemySize)
			y := room.Y + spawnPadding + rng.Float32()*(room.Height-2*spawnPadding-enemySize)
			candidate := AABB{X: x, Y: y, Width: enemySize, Height: enemySize}
			if overlapsAny(candidate, room.Obstacles) || overlaps(candidate, centerSafe) {
				continue
			}
			result = append(result, &EnemyRef{X: x, Y: y, IsElite: i == enemyCount-1})
			placed = true
			break
		}
		if placed {
			continue
		}

		fallbackX := room.X + spawnPadding + float32(i%2)*80
		fallbackY := room.Y + spawnPadding + float32(i/2)*80
		result = append(result, &EnemyRef{X: fallbackX, Y: fallbackY, IsElite: i == enemyCount-1})
	}

	return result
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
