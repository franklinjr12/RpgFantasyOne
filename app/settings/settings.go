package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const DefaultPath = "settings.json"

type KeybindDisplay struct {
	Move   string `json:"move"`
	Attack string `json:"attack"`
	Skill1 string `json:"skill_1"`
	Skill2 string `json:"skill_2"`
	Skill3 string `json:"skill_3"`
	Skill4 string `json:"skill_4"`
}

type Settings struct {
	MasterVolume   float32        `json:"master_volume"`
	Fullscreen     bool           `json:"fullscreen"`
	KeybindDisplay KeybindDisplay `json:"keybind_display"`
}

func Default() Settings {
	return Settings{
		MasterVolume: 1.0,
		Fullscreen:   false,
		KeybindDisplay: KeybindDisplay{
			Move:   "RMB",
			Attack: "LMB",
			Skill1: "Q",
			Skill2: "W",
			Skill3: "E",
			Skill4: "R",
		},
	}
}

func (s Settings) SkillLabels() []string {
	return []string{s.KeybindDisplay.Skill1, s.KeybindDisplay.Skill2, s.KeybindDisplay.Skill3, s.KeybindDisplay.Skill4}
}

func Load() Settings {
	cfg, err := LoadFromPath(DefaultPath)
	if err != nil {
		return Default()
	}
	return cfg
}

func LoadFromPath(path string) (Settings, error) {
	defaults := Default()

	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaults, nil
		}
		return defaults, err
	}

	var cfg Settings
	if err := json.Unmarshal(content, &cfg); err != nil {
		return defaults, nil
	}

	applyDefaults(&cfg, defaults)
	cfg.MasterVolume = clamp(cfg.MasterVolume, 0, 1)

	return cfg, nil
}

func Save(cfg Settings) error {
	return SaveToPath(DefaultPath, cfg)
}

func SaveToPath(path string, cfg Settings) error {
	cfg.MasterVolume = clamp(cfg.MasterVolume, 0, 1)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, payload, 0o644); err != nil {
		return err
	}

	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(path)
		if renameErr := os.Rename(tempPath, path); renameErr != nil {
			_ = os.Remove(tempPath)
			return renameErr
		}
	}

	return nil
}

func applyDefaults(cfg *Settings, defaults Settings) {
	if cfg.KeybindDisplay.Move == "" {
		cfg.KeybindDisplay.Move = defaults.KeybindDisplay.Move
	}
	if cfg.KeybindDisplay.Attack == "" {
		cfg.KeybindDisplay.Attack = defaults.KeybindDisplay.Attack
	}
	if cfg.KeybindDisplay.Skill1 == "" {
		cfg.KeybindDisplay.Skill1 = defaults.KeybindDisplay.Skill1
	}
	if cfg.KeybindDisplay.Skill2 == "" {
		cfg.KeybindDisplay.Skill2 = defaults.KeybindDisplay.Skill2
	}
	if cfg.KeybindDisplay.Skill3 == "" {
		cfg.KeybindDisplay.Skill3 = defaults.KeybindDisplay.Skill3
	}
	if cfg.KeybindDisplay.Skill4 == "" {
		cfg.KeybindDisplay.Skill4 = defaults.KeybindDisplay.Skill4
	}
}

func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
