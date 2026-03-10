package world

import "fmt"

func AreDoorDirectionsCompatible(left, right DoorDirection) bool {
	if !left.IsValid() || !right.IsValid() {
		return false
	}
	return left.Opposite() == right
}

func AreDoorMarkersAligned(a, b DoorMarker) bool {
	if !AreDoorDirectionsCompatible(a.Direction, b.Direction) {
		return false
	}
	if (a.Direction == DoorDirectionEast && b.Direction == DoorDirectionWest) || (a.Direction == DoorDirectionWest && b.Direction == DoorDirectionEast) {
		return a.Y == b.Y
	}
	return a.X == b.X
}

func RotateRoomTemplate(template *RoomTemplate, degrees int) (*RoomTemplate, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil")
	}

	normalized := ((degrees % 360) + 360) % 360
	if normalized%90 != 0 {
		return nil, fmt.Errorf("rotation must be multiple of 90, got %d", degrees)
	}

	result := template.Clone()
	turns := normalized / 90
	for i := 0; i < turns; i++ {
		result = rotateRoomTemplate90(result)
	}
	return result, nil
}

func rotateRoomTemplate90(template *RoomTemplate) *RoomTemplate {
	width := template.Width
	height := template.Height

	rotatedTiles := make([][]TileType, width)
	for y := 0; y < width; y++ {
		rotatedTiles[y] = make([]TileType, height)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			nx, ny := rotateCoordCW(x, y, width, height)
			rotatedTiles[ny][nx] = template.Tiles[y][x]
		}
	}

	rotated := template.Clone()
	rotated.Width = height
	rotated.Height = width
	rotated.Tiles = rotatedTiles
	rotated.Doors = rotateDoorsCW(template.Doors, width, height)
	rotated.SpawnMarkers = rotateSpawnsCW(template.SpawnMarkers, width, height)
	rotated.PropMarkers = rotatePropsCW(template.PropMarkers, width, height)
	rotated.EventMarkers = rotateEventsCW(template.EventMarkers, width, height)
	rotated.HazardMarkers = rotateHazardsCW(template.HazardMarkers, width, height)
	rotated.TrapMarkers = rotateTrapsCW(template.TrapMarkers, width, height)
	return rotated
}

func rotateCoordCW(x, y, width, height int) (int, int) {
	return height - 1 - y, x
}

func rotateDoorDirectionCW(direction DoorDirection) DoorDirection {
	switch direction {
	case DoorDirectionNorth:
		return DoorDirectionEast
	case DoorDirectionEast:
		return DoorDirectionSouth
	case DoorDirectionSouth:
		return DoorDirectionWest
	case DoorDirectionWest:
		return DoorDirectionNorth
	default:
		return direction
	}
}

func rotateDoorsCW(input []DoorMarker, width, height int) []DoorMarker {
	result := make([]DoorMarker, 0, len(input))
	for _, marker := range input {
		nx, ny := rotateCoordCW(marker.X, marker.Y, width, height)
		result = append(result, DoorMarker{
			X:         nx,
			Y:         ny,
			Direction: rotateDoorDirectionCW(marker.Direction),
		})
	}
	return result
}

func rotateSpawnsCW(input []SpawnMarker, width, height int) []SpawnMarker {
	result := make([]SpawnMarker, 0, len(input))
	for _, marker := range input {
		nx, ny := rotateCoordCW(marker.X, marker.Y, width, height)
		result = append(result, SpawnMarker{
			X:    nx,
			Y:    ny,
			Type: marker.Type,
		})
	}
	return result
}

func rotatePropsCW(input []PropMarker, width, height int) []PropMarker {
	result := make([]PropMarker, 0, len(input))
	for _, marker := range input {
		nx, ny := rotateCoordCW(marker.X, marker.Y, width, height)
		result = append(result, PropMarker{
			X: nx,
			Y: ny,
		})
	}
	return result
}

func rotateEventsCW(input []EventMarker, width, height int) []EventMarker {
	result := make([]EventMarker, 0, len(input))
	for _, marker := range input {
		nx, ny := rotateCoordCW(marker.X, marker.Y, width, height)
		result = append(result, EventMarker{
			X: nx,
			Y: ny,
		})
	}
	return result
}

func rotateHazardsCW(input []HazardMarker, width, height int) []HazardMarker {
	result := make([]HazardMarker, 0, len(input))
	for _, marker := range input {
		nx, ny := rotateCoordCW(marker.X, marker.Y, width, height)
		result = append(result, HazardMarker{
			X: nx,
			Y: ny,
		})
	}
	return result
}

func rotateTrapsCW(input []TrapMarker, width, height int) []TrapMarker {
	result := make([]TrapMarker, 0, len(input))
	for _, marker := range input {
		nx, ny := rotateCoordCW(marker.X, marker.Y, width, height)
		result = append(result, TrapMarker{
			X: nx,
			Y: ny,
		})
	}
	return result
}
