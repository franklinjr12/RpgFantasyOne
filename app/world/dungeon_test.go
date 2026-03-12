package world

import "testing"

func TestDungeonGenerationInvariants(t *testing.T) {
	dungeon := NewDungeon()
	if dungeon == nil {
		t.Fatalf("expected dungeon, got nil")
	}
	if len(dungeon.Rooms) < DefaultRunMinRooms || len(dungeon.Rooms) > DefaultRunMaxRooms {
		t.Fatalf("unexpected room count: %d", len(dungeon.Rooms))
	}
	if dungeon.Rooms[0].Type != RoomTypeStart {
		t.Fatalf("first room should be start, got %s", dungeon.Rooms[0].Type.String())
	}
	if !dungeon.Rooms[len(dungeon.Rooms)-1].IsBoss() {
		t.Fatalf("last room should be boss")
	}

	eventCount := 0
	eliteCount := 0
	for index, room := range dungeon.Rooms {
		if room == nil {
			t.Fatalf("room %d is nil", index)
		}
		if room.ProgressionIndex != index {
			t.Fatalf("room %d progression mismatch: %d", index, room.ProgressionIndex)
		}

		if room.Type == RoomTypeEvent {
			eventCount++
			if room.EventTimeLeft <= 0 {
				t.Fatalf("event room %d missing timer", index)
			}
		}
		if room.Type == RoomTypeElite {
			eliteCount++
		}

		for _, obstacle := range room.Obstacles {
			if obstacle.X < room.X || obstacle.Y < room.Y {
				t.Fatalf("obstacle out of bounds in room %d", index)
			}
			if obstacle.X+obstacle.Width > room.X+room.Width || obstacle.Y+obstacle.Height > room.Y+room.Height {
				t.Fatalf("obstacle out of room bounds in room %d", index)
			}
		}

		if index < len(dungeon.Rooms)-1 {
			hasProgressionDoor := false
			for _, door := range room.Doors {
				if door == nil {
					t.Fatalf("nil door in room %d", index)
				}
				if door.Bounds.X < room.X || door.Bounds.Y < room.Y {
					t.Fatalf("door out of bounds in room %d", index)
				}
				if door.Bounds.X+door.Bounds.Width > room.X+room.Width || door.Bounds.Y+door.Bounds.Height > room.Y+room.Height {
					t.Fatalf("door dimensions out of room bounds in room %d", index)
				}
				if door.TargetRoomIndex == index+1 {
					hasProgressionDoor = true
				}
			}
			if !hasProgressionDoor {
				t.Fatalf("room %d missing progression door to %d", index, index+1)
			}
		}
	}

	if eventCount < DefaultRunMinEventRooms {
		t.Fatalf("expected at least %d event room(s), got %d", DefaultRunMinEventRooms, eventCount)
	}
	if eliteCount < 1 {
		t.Fatalf("expected at least one elite room")
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

		if l.TemplateID != r.TemplateID || l.Type != r.Type || l.Rotation != r.Rotation {
			t.Fatalf("room identity mismatch at %d", i)
		}
		if l.X != r.X || l.Y != r.Y || l.Width != r.Width || l.Height != r.Height {
			t.Fatalf("room geometry mismatch at %d", i)
		}
		if len(l.Obstacles) != len(r.Obstacles) {
			t.Fatalf("obstacle count mismatch at %d", i)
		}
		if len(l.Enemies) != len(r.Enemies) {
			t.Fatalf("enemy count mismatch at %d", i)
		}
		for j := range l.Enemies {
			le := l.Enemies[j]
			re := r.Enemies[j]
			if le.Type != re.Type || le.IsElite != re.IsElite || le.EliteModifier != re.EliteModifier {
				t.Fatalf("enemy %d mismatch in room %d", j, i)
			}
			if le.X != re.X || le.Y != re.Y {
				t.Fatalf("enemy spawn mismatch at room %d enemy %d", i, j)
			}
		}
	}
}

func TestDungeonBossRoomHasBossSpawnMarkerAndWestDoor(t *testing.T) {
	dungeon := NewDungeon()
	if dungeon == nil || len(dungeon.Rooms) == 0 {
		t.Fatalf("expected dungeon rooms")
	}

	bossRoom := dungeon.Rooms[len(dungeon.Rooms)-1]
	if bossRoom == nil || !bossRoom.IsBoss() {
		t.Fatalf("expected final room to be boss")
	}
	if !bossRoom.HasBossSpawn {
		t.Fatalf("expected boss room to expose boss spawn point")
	}

	hasWestDoor := false
	for _, door := range bossRoom.Doors {
		if door != nil && door.Direction == DoorDirectionWest {
			hasWestDoor = true
			break
		}
	}
	if !hasWestDoor {
		t.Fatalf("expected boss room to have at least one west entry door")
	}
}
