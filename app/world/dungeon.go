package world

const (
	DungeonLength = 5
)

type Dungeon struct {
	Rooms       []*Room
	CurrentRoom int
}

func NewDungeon() *Dungeon {
	dungeon := &Dungeon{
		Rooms:       []*Room{},
		CurrentRoom: 0,
	}

	x := float32(0)
	y := float32(0)
	roomSpacing := float32(RoomMaxWidth + 200)

	for i := 0; i < DungeonLength; i++ {
		room := NewRoom(x, y, RoomTypeNormal)
		dungeon.Rooms = append(dungeon.Rooms, room)
		x += roomSpacing
	}

	bossRoom := NewRoom(x, y, RoomTypeBoss)
	dungeon.Rooms = append(dungeon.Rooms, bossRoom)

	return dungeon
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
