package world

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type layoutParseResult struct {
	width         int
	height        int
	tiles         [][]TileType
	doors         []DoorMarker
	spawnMarkers  []SpawnMarker
	propMarkers   []PropMarker
	eventMarkers  []EventMarker
	hazardMarkers []HazardMarker
	trapMarkers   []TrapMarker
}

type roomMetaFile struct {
	ID            string         `json:"id"`
	Biome         string         `json:"biome"`
	Type          string         `json:"type"`
	Difficulty    int            `json:"difficulty"`
	Weight        int            `json:"weight"`
	AllowRotation bool           `json:"allow_rotation"`
	Tags          []string       `json:"tags"`
	Doors         []roomMetaDoor `json:"doors"`
}

type roomMetaDoor struct {
	X   int    `json:"x"`
	Y   int    `json:"y"`
	Dir string `json:"dir"`
}

func parseRoomLayout(content string) (*layoutParseResult, error) {
	trimmed := strings.TrimRight(content, "\r\n")
	if trimmed == "" {
		return nil, fmt.Errorf("layout is empty")
	}

	rows := strings.Split(trimmed, "\n")
	height := len(rows)
	if height < RoomTemplateMinSize || height > RoomTemplateMaxSize {
		return nil, fmt.Errorf("layout height %d outside bounds %d..%d", height, RoomTemplateMinSize, RoomTemplateMaxSize)
	}

	width := -1
	tiles := make([][]TileType, height)
	doors := []DoorMarker{}
	spawns := []SpawnMarker{}
	props := []PropMarker{}
	events := []EventMarker{}
	hazards := []HazardMarker{}
	traps := []TrapMarker{}

	for y, raw := range rows {
		row := strings.TrimSuffix(raw, "\r")
		if width == -1 {
			width = len(row)
			if width < RoomTemplateMinSize || width > RoomTemplateMaxSize {
				return nil, fmt.Errorf("layout width %d outside bounds %d..%d", width, RoomTemplateMinSize, RoomTemplateMaxSize)
			}
		}
		if len(row) != width {
			return nil, fmt.Errorf("non-rectangular layout at row %d: got width %d, expected %d", y, len(row), width)
		}

		tiles[y] = make([]TileType, width)
		for x, ch := range row {
			switch ch {
			case '#', ' ':
				tiles[y][x] = TileWall
			case '.':
				tiles[y][x] = TileFloor
			case 'D':
				tiles[y][x] = TileDoor
				doors = append(doors, DoorMarker{X: x, Y: y})
			case 's':
				tiles[y][x] = TileFloor
				spawns = append(spawns, SpawnMarker{X: x, Y: y, Type: SpawnTypeNormal})
			case 'P':
				tiles[y][x] = TileFloor
				props = append(props, PropMarker{X: x, Y: y})
			case 'E':
				tiles[y][x] = TileFloor
				spawns = append(spawns, SpawnMarker{X: x, Y: y, Type: SpawnTypeElite})
			case 'B':
				tiles[y][x] = TileFloor
				spawns = append(spawns, SpawnMarker{X: x, Y: y, Type: SpawnTypeBoss})
			case 'R':
				tiles[y][x] = TileFloor
				events = append(events, EventMarker{X: x, Y: y})
			case 'H':
				tiles[y][x] = TileHazard
				hazards = append(hazards, HazardMarker{X: x, Y: y})
			case 'T':
				tiles[y][x] = TileTrap
				traps = append(traps, TrapMarker{X: x, Y: y})
			default:
				return nil, fmt.Errorf("unsupported layout character %q at (%d,%d)", ch, x, y)
			}
		}
	}

	return &layoutParseResult{
		width:         width,
		height:        height,
		tiles:         tiles,
		doors:         doors,
		spawnMarkers:  spawns,
		propMarkers:   props,
		eventMarkers:  events,
		hazardMarkers: hazards,
		trapMarkers:   traps,
	}, nil
}

func loadRoomTemplatePair(layoutPath, metaPath string) (*RoomTemplate, error) {
	layoutBytes, err := os.ReadFile(layoutPath)
	if err != nil {
		return nil, fmt.Errorf("read layout %q: %w", layoutPath, err)
	}
	layoutResult, err := parseRoomLayout(string(layoutBytes))
	if err != nil {
		return nil, fmt.Errorf("parse layout %q: %w", layoutPath, err)
	}

	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("read metadata %q: %w", metaPath, err)
	}
	var meta roomMetaFile
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return nil, fmt.Errorf("parse metadata %q: %w", metaPath, err)
	}

	meta.ID = strings.TrimSpace(meta.ID)
	meta.Biome = strings.TrimSpace(meta.Biome)
	meta.Type = strings.TrimSpace(meta.Type)
	if meta.ID == "" {
		return nil, fmt.Errorf("metadata %q missing required field id", metaPath)
	}
	if meta.Biome == "" {
		return nil, fmt.Errorf("metadata %q missing required field biome", metaPath)
	}
	if meta.Type == "" {
		return nil, fmt.Errorf("metadata %q missing required field type", metaPath)
	}
	if len(meta.Doors) == 0 {
		return nil, fmt.Errorf("metadata %q missing required field doors", metaPath)
	}

	roomType, err := ParseRoomType(meta.Type)
	if err != nil {
		return nil, fmt.Errorf("metadata %q: %w", metaPath, err)
	}

	layoutBase := strings.TrimSuffix(filepath.Base(layoutPath), filepath.Ext(layoutPath))
	metaBase := strings.TrimSuffix(filepath.Base(metaPath), filepath.Ext(metaPath))
	metaBase = strings.TrimSuffix(metaBase, ".meta")
	if layoutBase != metaBase {
		return nil, fmt.Errorf("file pair mismatch layout=%q metadata=%q", layoutBase, metaBase)
	}
	if meta.ID != layoutBase {
		return nil, fmt.Errorf("metadata id %q does not match filename base %q", meta.ID, layoutBase)
	}

	layoutDoorSet := map[string]struct{}{}
	for _, door := range layoutResult.doors {
		layoutDoorSet[fmt.Sprintf("%d,%d", door.X, door.Y)] = struct{}{}
	}

	doors := make([]DoorMarker, 0, len(meta.Doors))
	for _, metaDoor := range meta.Doors {
		if metaDoor.X < 0 || metaDoor.X >= layoutResult.width || metaDoor.Y < 0 || metaDoor.Y >= layoutResult.height {
			return nil, fmt.Errorf("door out of bounds at (%d,%d)", metaDoor.X, metaDoor.Y)
		}
		key := fmt.Sprintf("%d,%d", metaDoor.X, metaDoor.Y)
		if _, ok := layoutDoorSet[key]; !ok {
			return nil, fmt.Errorf("door (%d,%d) missing matching D marker in layout", metaDoor.X, metaDoor.Y)
		}
		delete(layoutDoorSet, key)

		direction, err := ParseDoorDirection(metaDoor.Dir)
		if err != nil {
			return nil, fmt.Errorf("door (%d,%d): %w", metaDoor.X, metaDoor.Y, err)
		}

		doors = append(doors, DoorMarker{
			X:         metaDoor.X,
			Y:         metaDoor.Y,
			Direction: direction,
		})
	}

	if len(layoutDoorSet) > 0 {
		return nil, fmt.Errorf("layout has D marker(s) missing from metadata")
	}

	weight := meta.Weight
	if weight <= 0 {
		weight = 1
	}
	difficulty := meta.Difficulty
	if difficulty <= 0 {
		difficulty = 1
	}

	template := &RoomTemplate{
		ID:            meta.ID,
		Biome:         strings.ToLower(meta.Biome),
		Type:          roomType,
		Width:         layoutResult.width,
		Height:        layoutResult.height,
		Tiles:         layoutResult.tiles,
		Doors:         doors,
		SpawnMarkers:  layoutResult.spawnMarkers,
		PropMarkers:   layoutResult.propMarkers,
		EventMarkers:  layoutResult.eventMarkers,
		HazardMarkers: layoutResult.hazardMarkers,
		TrapMarkers:   layoutResult.trapMarkers,
		Tags:          append([]string(nil), meta.Tags...),
		Weight:        weight,
		Difficulty:    difficulty,
		AllowRotation: meta.AllowRotation,
	}

	if err := validateTemplateForRoomType(template); err != nil {
		return nil, err
	}

	return template, nil
}

func validateTemplateForRoomType(template *RoomTemplate) error {
	if template == nil {
		return fmt.Errorf("room template is nil")
	}
	if template.Type == RoomTypeBoss {
		return validateBossTemplate(template)
	}
	return nil
}

func validateBossTemplate(template *RoomTemplate) error {
	westDoors := make([]DoorMarker, 0, len(template.Doors))
	for _, door := range template.Doors {
		if door.Direction != DoorDirectionWest {
			continue
		}
		westDoors = append(westDoors, door)

		interiorX := door.X + 1
		interiorY := door.Y
		if !isTemplateCoordinateInBounds(template, interiorX, interiorY) {
			return fmt.Errorf("boss template %q west door (%d,%d) has out-of-bounds interior tile", template.ID, door.X, door.Y)
		}
		if !isTemplateWalkableTile(template.Tiles[interiorY][interiorX]) {
			return fmt.Errorf("boss template %q west door (%d,%d) has blocked interior tile", template.ID, door.X, door.Y)
		}
	}
	if len(westDoors) == 0 {
		return fmt.Errorf("boss template %q requires at least one west door", template.ID)
	}

	bossMarkers := make([]SpawnMarker, 0, 1)
	for _, marker := range template.SpawnMarkers {
		if marker.Type == SpawnTypeBoss {
			bossMarkers = append(bossMarkers, marker)
		}
	}
	if len(bossMarkers) != 1 {
		return fmt.Errorf("boss template %q must contain exactly one boss spawn marker, got %d", template.ID, len(bossMarkers))
	}

	bossMarker := bossMarkers[0]
	if !hasTemplateWalkableNeighbor(template, bossMarker.X, bossMarker.Y) {
		return fmt.Errorf("boss template %q boss spawn marker at (%d,%d) is blocked", template.ID, bossMarker.X, bossMarker.Y)
	}

	minDistanceToDoor := -1
	for _, door := range westDoors {
		distance := absInt(bossMarker.X-door.X) + absInt(bossMarker.Y-door.Y)
		if minDistanceToDoor == -1 || distance < minDistanceToDoor {
			minDistanceToDoor = distance
		}
	}
	if minDistanceToDoor >= 0 && minDistanceToDoor < 3 {
		return fmt.Errorf("boss template %q boss spawn marker too close to west entry", template.ID)
	}

	return nil
}

func isTemplateCoordinateInBounds(template *RoomTemplate, x, y int) bool {
	return template != nil && y >= 0 && y < len(template.Tiles) && x >= 0 && x < len(template.Tiles[y])
}

func isTemplateWalkableTile(tile TileType) bool {
	return tile == TileFloor || tile == TileDoor
}

func hasTemplateWalkableNeighbor(template *RoomTemplate, x, y int) bool {
	neighbors := [][2]int{
		{x + 1, y},
		{x - 1, y},
		{x, y + 1},
		{x, y - 1},
	}
	for _, neighbor := range neighbors {
		nx := neighbor[0]
		ny := neighbor[1]
		if !isTemplateCoordinateInBounds(template, nx, ny) {
			continue
		}
		if isTemplateWalkableTile(template.Tiles[ny][nx]) {
			return true
		}
	}
	return false
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
