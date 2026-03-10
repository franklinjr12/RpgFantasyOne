package world

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRoomTemplateRegistryDeterministicOrder(t *testing.T) {
	tempDir := t.TempDir()
	biomeDir := filepath.Join(tempDir, "forest")
	if err := os.MkdirAll(biomeDir, 0o755); err != nil {
		t.Fatalf("mkdir biome dir: %v", err)
	}

	writePair := func(base, layout, meta string) {
		if err := os.WriteFile(filepath.Join(biomeDir, base+".layout"), []byte(layout), 0o644); err != nil {
			t.Fatalf("write layout: %v", err)
		}
		if err := os.WriteFile(filepath.Join(biomeDir, base+".meta.json"), []byte(meta), 0o644); err != nil {
			t.Fatalf("write metadata: %v", err)
		}
	}

	writePair("z_room", "########\n#......#\n#......#\nD......D\n#......#\n#......#\n#......#\n########",
		`{"id":"z_room","biome":"forest","type":"combat","weight":1,"doors":[{"x":0,"y":3,"dir":"west"},{"x":7,"y":3,"dir":"east"}]}`)
	writePair("a_room", "########\n#......#\n#......#\nD......D\n#......#\n#......#\n#......#\n########",
		`{"id":"a_room","biome":"forest","type":"combat","weight":1,"doors":[{"x":0,"y":3,"dir":"west"},{"x":7,"y":3,"dir":"east"}]}`)

	registry, err := LoadRoomTemplateRegistry(tempDir)
	if err != nil {
		t.Fatalf("unexpected registry error: %v", err)
	}
	if len(registry.Templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(registry.Templates))
	}
	if registry.Templates[0].ID != "a_room" || registry.Templates[1].ID != "z_room" {
		t.Fatalf("expected deterministic sorted order, got %q then %q", registry.Templates[0].ID, registry.Templates[1].ID)
	}
}

func TestLoadRoomTemplateRegistryMissingPairFails(t *testing.T) {
	tempDir := t.TempDir()
	biomeDir := filepath.Join(tempDir, "forest")
	if err := os.MkdirAll(biomeDir, 0o755); err != nil {
		t.Fatalf("mkdir biome dir: %v", err)
	}

	layout := "########\n#......#\n#......#\nD......D\n#......#\n#......#\n#......#\n########"
	if err := os.WriteFile(filepath.Join(biomeDir, "orphan.layout"), []byte(layout), 0o644); err != nil {
		t.Fatalf("write layout: %v", err)
	}

	if _, err := LoadRoomTemplateRegistry(tempDir); err == nil {
		t.Fatalf("expected missing metadata error")
	}
}
