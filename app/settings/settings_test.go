package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromPathMissingReturnsDefaults(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "missing_settings.json")

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}

	defaults := Default()
	if cfg.MasterVolume != defaults.MasterVolume {
		t.Fatalf("expected default volume %v, got %v", defaults.MasterVolume, cfg.MasterVolume)
	}
	if cfg.Fullscreen != defaults.Fullscreen {
		t.Fatalf("expected default fullscreen %v, got %v", defaults.Fullscreen, cfg.Fullscreen)
	}
	if cfg.KeybindDisplay.Skill1 != defaults.KeybindDisplay.Skill1 {
		t.Fatalf("expected default skill1 key %q, got %q", defaults.KeybindDisplay.Skill1, cfg.KeybindDisplay.Skill1)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "settings.json")
	in := Settings{
		MasterVolume: 0.55,
		Fullscreen:   true,
		KeybindDisplay: KeybindDisplay{
			Move:   "RMB",
			Attack: "LMB",
			Skill1: "1",
			Skill2: "2",
			Skill3: "3",
			Skill4: "4",
		},
	}

	if err := SaveToPath(path, in); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	out, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if out.MasterVolume != in.MasterVolume {
		t.Fatalf("expected volume %v, got %v", in.MasterVolume, out.MasterVolume)
	}
	if out.Fullscreen != in.Fullscreen {
		t.Fatalf("expected fullscreen %v, got %v", in.Fullscreen, out.Fullscreen)
	}
	if out.KeybindDisplay.Skill3 != in.KeybindDisplay.Skill3 {
		t.Fatalf("expected skill3 key %q, got %q", in.KeybindDisplay.Skill3, out.KeybindDisplay.Skill3)
	}
}

func TestLoadFromPathInvalidJSONFallsBackToDefaults(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "settings.json")

	if err := os.WriteFile(path, []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("expected nil error for invalid file fallback, got %v", err)
	}

	defaults := Default()
	if cfg.MasterVolume != defaults.MasterVolume {
		t.Fatalf("expected default volume %v, got %v", defaults.MasterVolume, cfg.MasterVolume)
	}
	if cfg.KeybindDisplay.Attack != defaults.KeybindDisplay.Attack {
		t.Fatalf("expected default attack key %q, got %q", defaults.KeybindDisplay.Attack, cfg.KeybindDisplay.Attack)
	}
}
