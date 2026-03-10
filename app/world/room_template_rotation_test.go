package world

import "testing"

func TestRotateRoomTemplate90(t *testing.T) {
	template := &RoomTemplate{
		ID:     "rotate_case",
		Biome:  "forest",
		Type:   RoomTypeCombat,
		Width:  4,
		Height: 3,
		Tiles: [][]TileType{
			{TileWall, TileFloor, TileFloor, TileWall},
			{TileDoor, TileFloor, TileFloor, TileDoor},
			{TileWall, TileFloor, TileFloor, TileWall},
		},
		Doors: []DoorMarker{
			{X: 0, Y: 1, Direction: DoorDirectionWest},
			{X: 3, Y: 1, Direction: DoorDirectionEast},
		},
		SpawnMarkers: []SpawnMarker{
			{X: 1, Y: 1, Type: SpawnTypeNormal},
		},
		AllowRotation: true,
	}

	rotated, err := RotateRoomTemplate(template, 90)
	if err != nil {
		t.Fatalf("rotate failed: %v", err)
	}
	if rotated.Width != 3 || rotated.Height != 4 {
		t.Fatalf("unexpected rotated size %dx%d", rotated.Width, rotated.Height)
	}

	var northDoor, southDoor bool
	for _, door := range rotated.Doors {
		if door.Direction == DoorDirectionNorth {
			northDoor = true
		}
		if door.Direction == DoorDirectionSouth {
			southDoor = true
		}
	}
	if !northDoor || !southDoor {
		t.Fatalf("expected west/east doors to rotate into north/south")
	}
}

func TestDoorCompatibilityHelpers(t *testing.T) {
	if !AreDoorDirectionsCompatible(DoorDirectionEast, DoorDirectionWest) {
		t.Fatalf("expected east/west compatibility")
	}
	if AreDoorDirectionsCompatible(DoorDirectionEast, DoorDirectionNorth) {
		t.Fatalf("unexpected east/north compatibility")
	}
	if !AreDoorMarkersAligned(
		DoorMarker{X: 7, Y: 3, Direction: DoorDirectionEast},
		DoorMarker{X: 0, Y: 3, Direction: DoorDirectionWest},
	) {
		t.Fatalf("expected east-west markers with same y to align")
	}
}
