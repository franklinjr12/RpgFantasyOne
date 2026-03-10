package world

import "strings"

const (
	RoomTemplateMinSize  = 8
	RoomTemplateMaxSize  = 30
	RoomTemplateTileSize = 96
)

type DoorMarker struct {
	X         int
	Y         int
	Direction DoorDirection
}

type SpawnMarker struct {
	X    int
	Y    int
	Type SpawnType
}

type PropMarker struct {
	X int
	Y int
}

type EventMarker struct {
	X int
	Y int
}

type HazardMarker struct {
	X int
	Y int
}

type TrapMarker struct {
	X int
	Y int
}

type RoomTemplate struct {
	ID            string
	Biome         string
	Type          RoomType
	Width         int
	Height        int
	Tiles         [][]TileType
	Doors         []DoorMarker
	SpawnMarkers  []SpawnMarker
	PropMarkers   []PropMarker
	EventMarkers  []EventMarker
	HazardMarkers []HazardMarker
	TrapMarkers   []TrapMarker
	Tags          []string
	Weight        int
	Difficulty    int
	AllowRotation bool
}

func (t *RoomTemplate) Clone() *RoomTemplate {
	if t == nil {
		return nil
	}

	clone := *t
	clone.Tiles = make([][]TileType, len(t.Tiles))
	for y := range t.Tiles {
		clone.Tiles[y] = make([]TileType, len(t.Tiles[y]))
		copy(clone.Tiles[y], t.Tiles[y])
	}
	clone.Doors = append([]DoorMarker(nil), t.Doors...)
	clone.SpawnMarkers = append([]SpawnMarker(nil), t.SpawnMarkers...)
	clone.PropMarkers = append([]PropMarker(nil), t.PropMarkers...)
	clone.EventMarkers = append([]EventMarker(nil), t.EventMarkers...)
	clone.HazardMarkers = append([]HazardMarker(nil), t.HazardMarkers...)
	clone.TrapMarkers = append([]TrapMarker(nil), t.TrapMarkers...)
	clone.Tags = append([]string(nil), t.Tags...)
	return &clone
}

func (t *RoomTemplate) HasDoorDirection(direction DoorDirection) bool {
	for _, door := range t.Doors {
		if door.Direction == direction {
			return true
		}
	}
	return false
}

func (t *RoomTemplate) DoorMarkersByDirection(direction DoorDirection) []DoorMarker {
	markers := make([]DoorMarker, 0, len(t.Doors))
	for _, door := range t.Doors {
		if door.Direction == direction {
			markers = append(markers, door)
		}
	}
	return markers
}

func (t *RoomTemplate) HasTag(tag string) bool {
	tag = strings.ToLower(strings.TrimSpace(tag))
	for _, value := range t.Tags {
		if strings.ToLower(strings.TrimSpace(value)) == tag {
			return true
		}
	}
	return false
}
