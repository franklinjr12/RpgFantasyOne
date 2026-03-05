package world

import "testing"

func TestDungeonGenerationInvariants(t *testing.T) {
	dungeon := NewDungeon()
	if len(dungeon.Rooms) != DungeonLength+1 {
		t.Fatalf("unexpected room count: %d", len(dungeon.Rooms))
	}

	for index, room := range dungeon.Rooms {
		if room == nil {
			t.Fatalf("room %d is nil", index)
		}

		for _, obstacle := range room.Obstacles {
			if obstacle.X < room.X || obstacle.Y < room.Y {
				t.Fatalf("obstacle out of bounds in room %d", index)
			}
			if obstacle.X+obstacle.Width > room.X+room.Width || obstacle.Y+obstacle.Height > room.Y+room.Height {
				t.Fatalf("obstacle out of room bounds in room %d", index)
			}
		}

		if room.IsBoss() {
			continue
		}

		if len(room.Doors) == 0 {
			t.Fatalf("normal room %d has no doors", index)
		}

		for _, door := range room.Doors {
			if door == nil {
				t.Fatalf("nil door in room %d", index)
			}
			if door.TargetRoomIndex != index+1 {
				t.Fatalf("door target mismatch in room %d: got %d", index, door.TargetRoomIndex)
			}
			if door.Bounds.X < room.X || door.Bounds.Y < room.Y {
				t.Fatalf("door out of bounds in room %d", index)
			}
			if door.Bounds.X+door.Bounds.Width > room.X+room.Width || door.Bounds.Y+door.Bounds.Height > room.Y+room.Height {
				t.Fatalf("door dimensions out of room bounds in room %d", index)
			}
		}
	}
}

func TestDungeonGenerationIsDeterministic(t *testing.T) {
	left := NewDungeon()
	right := NewDungeon()

	if len(left.Rooms) != len(right.Rooms) {
		t.Fatalf("room count mismatch: %d vs %d", len(left.Rooms), len(right.Rooms))
	}

	for i := range left.Rooms {
		l := left.Rooms[i]
		r := right.Rooms[i]
		if l.Width != r.Width || l.Height != r.Height || l.X != r.X || l.Y != r.Y {
			t.Fatalf("room geometry mismatch at %d", i)
		}
		if len(l.Obstacles) != len(r.Obstacles) {
			t.Fatalf("obstacle count mismatch at %d", i)
		}
		if len(l.Enemies) != len(r.Enemies) {
			t.Fatalf("enemy count mismatch at %d", i)
		}
	}
}
