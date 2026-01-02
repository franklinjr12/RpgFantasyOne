package world

import "math/rand"

const (
	RoomMinWidth      = 400
	RoomMaxWidth      = 600
	RoomMinHeight     = 300
	RoomMaxHeight     = 450
	BossRoomWidth     = 800
	BossRoomHeight    = 600
	EnemiesPerRoomMin = 3
	EnemiesPerRoomMax = 5
)

type RoomType int

const (
	RoomTypeNormal RoomType = iota
	RoomTypeBoss
)

type Room struct {
	X         float32
	Y         float32
	Width     float32
	Height    float32
	Type      RoomType
	Enemies   []*EnemyRef
	Completed bool
}

type EnemyRef struct {
	X       float32
	Y       float32
	IsElite bool
}

func NewRoom(x, y float32, roomType RoomType) *Room {
	var width, height float32
	if roomType == RoomTypeBoss {
		width = BossRoomWidth
		height = BossRoomHeight
	} else {
		width = float32(rand.Intn(RoomMaxWidth-RoomMinWidth+1) + RoomMinWidth)
		height = float32(rand.Intn(RoomMaxHeight-RoomMinHeight+1) + RoomMinHeight)
	}

	room := &Room{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Type:      roomType,
		Enemies:   []*EnemyRef{},
		Completed: false,
	}

	if roomType == RoomTypeNormal {
		enemyCount := rand.Intn(EnemiesPerRoomMax-EnemiesPerRoomMin+1) + EnemiesPerRoomMin
		for i := 0; i < enemyCount; i++ {
			enemyX := x + float32(rand.Float64())*(width-60)
			enemyY := y + float32(rand.Float64())*(height-60)
			isElite := i == enemyCount-1
			room.Enemies = append(room.Enemies, &EnemyRef{
				X:       enemyX,
				Y:       enemyY,
				IsElite: isElite,
			})
		}
	}

	return room
}

func (r *Room) IsBoss() bool {
	return r.Type == RoomTypeBoss
}
