package world

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRoomLayoutValid(t *testing.T) {
	layout := strings.Join([]string{
		"########",
		"#..s...#",
		"#..D...#",
		"#..E...#",
		"#..B...#",
		"#..R...#",
		"#..P..H#",
		"###T####",
	}, "\n")

	parsed, err := parseRoomLayout(layout)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if parsed.width != 8 || parsed.height != 8 {
		t.Fatalf("unexpected size %dx%d", parsed.width, parsed.height)
	}
	if len(parsed.doors) != 1 {
		t.Fatalf("expected 1 door, got %d", len(parsed.doors))
	}
	if len(parsed.spawnMarkers) != 3 {
		t.Fatalf("expected 3 spawn markers, got %d", len(parsed.spawnMarkers))
	}
	if len(parsed.propMarkers) != 1 || len(parsed.eventMarkers) != 1 || len(parsed.hazardMarkers) != 1 || len(parsed.trapMarkers) != 1 {
		t.Fatalf("expected prop/event/hazard/trap markers")
	}
}

func TestParseRoomLayoutRejectsInvalidShapeAndChars(t *testing.T) {
	_, err := parseRoomLayout("########\n#.....#\n########")
	if err == nil {
		t.Fatalf("expected non-rectangular parse error")
	}

	_, err = parseRoomLayout(strings.Join([]string{
		"########",
		"#......#",
		"#..x...#",
		"#......#",
		"#......#",
		"#......#",
		"#......#",
		"########",
	}, "\n"))
	if err == nil {
		t.Fatalf("expected unsupported character error")
	}
}

func TestLoadRoomTemplatePairValidation(t *testing.T) {
	tempDir := t.TempDir()
	layoutPath := filepath.Join(tempDir, "test_room.layout")
	metaPath := filepath.Join(tempDir, "test_room.meta.json")

	layout := strings.Join([]string{
		"########",
		"#......#",
		"#......#",
		"D......D",
		"#......#",
		"#......#",
		"#......#",
		"########",
	}, "\n")
	meta := `{
  "id":"test_room",
  "biome":"forest",
  "type":"combat",
  "difficulty":1,
  "weight":2,
  "allow_rotation":true,
  "tags":["small"],
  "doors":[
    {"x":0,"y":3,"dir":"west"},
    {"x":7,"y":3,"dir":"east"}
  ]
}`

	if err := os.WriteFile(layoutPath, []byte(layout), 0o644); err != nil {
		t.Fatalf("write layout: %v", err)
	}
	if err := os.WriteFile(metaPath, []byte(meta), 0o644); err != nil {
		t.Fatalf("write metadata: %v", err)
	}

	template, err := loadRoomTemplatePair(layoutPath, metaPath)
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if template.ID != "test_room" || template.Type != RoomTypeCombat {
		t.Fatalf("unexpected template data: %+v", template)
	}
	if !template.AllowRotation {
		t.Fatalf("expected allow_rotation=true")
	}

	badMetaPath := filepath.Join(tempDir, "test_room_bad.meta.json")
	badMeta := `{
  "id":"test_room_bad",
  "biome":"forest",
  "type":"combat",
  "doors":[{"x":1,"y":1,"dir":"north"}]
}`
	if err := os.WriteFile(badMetaPath, []byte(badMeta), 0o644); err != nil {
		t.Fatalf("write bad metadata: %v", err)
	}
	if _, err := loadRoomTemplatePair(layoutPath, badMetaPath); err == nil {
		t.Fatalf("expected door mismatch validation error")
	}
}
