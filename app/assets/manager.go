package assets

import (
	"log"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Asset keys used by the MVP foundation.
const (
	TextureHumanoidSpriteSheet = "texture.humanoid_sprite_sheet"
	FontDefault                = "font.default"
)

type AssetManager struct {
	textures     map[string]rl.Texture2D
	fonts        map[string]rl.Font
	sounds       map[string]rl.Sound
	music        map[string]rl.Music
	loadedFonts  map[string]bool
	loadedSounds map[string]bool
	loadedMusic  map[string]bool
}

var defaultManager = NewAssetManager()

func NewAssetManager() *AssetManager {
	return &AssetManager{
		textures:     make(map[string]rl.Texture2D),
		fonts:        make(map[string]rl.Font),
		sounds:       make(map[string]rl.Sound),
		music:        make(map[string]rl.Music),
		loadedFonts:  make(map[string]bool),
		loadedSounds: make(map[string]bool),
		loadedMusic:  make(map[string]bool),
	}
}

func Get() *AssetManager {
	return defaultManager
}

func (m *AssetManager) LoadTexture(key, path string, fallbackWidth, fallbackHeight int32, fallbackColor rl.Color) {
	if existing, ok := m.textures[key]; ok && existing.ID != 0 {
		rl.UnloadTexture(existing)
	}

	if fallbackWidth <= 0 {
		fallbackWidth = 64
	}
	if fallbackHeight <= 0 {
		fallbackHeight = 64
	}

	if path != "" && fileExists(path) {
		texture := rl.LoadTexture(path)
		if texture.ID != 0 {
			m.textures[key] = texture
			return
		}
	}

	log.Printf("asset warning: texture key=%q path=%q missing/invalid, using generated fallback", key, path)
	m.textures[key] = generateFallbackTexture(fallbackWidth, fallbackHeight, fallbackColor)
}

func (m *AssetManager) LoadFont(key, path string) {
	if path != "" && fileExists(path) {
		font := rl.LoadFont(path)
		if font.Texture.ID != 0 {
			m.fonts[key] = font
			m.loadedFonts[key] = true
			return
		}
	}

	if path != "" {
		log.Printf("asset warning: font key=%q path=%q missing/invalid, using default font", key, path)
	}
	m.fonts[key] = rl.GetFontDefault()
	m.loadedFonts[key] = false
}

func (m *AssetManager) LoadSound(key, path string) {
	if !rl.IsAudioDeviceReady() {
		log.Printf("asset warning: audio device not ready, sound key=%q set to no-op", key)
		m.sounds[key] = rl.Sound{}
		m.loadedSounds[key] = false
		return
	}

	if path != "" && fileExists(path) {
		sound := rl.LoadSound(path)
		if sound.FrameCount > 0 {
			m.sounds[key] = sound
			m.loadedSounds[key] = true
			return
		}
	}

	log.Printf("asset warning: sound key=%q path=%q missing/invalid, using no-op", key, path)
	m.sounds[key] = rl.Sound{}
	m.loadedSounds[key] = false
}

func (m *AssetManager) LoadMusic(key, path string) {
	if !rl.IsAudioDeviceReady() {
		log.Printf("asset warning: audio device not ready, music key=%q set to no-op", key)
		m.music[key] = rl.Music{}
		m.loadedMusic[key] = false
		return
	}

	if path != "" && fileExists(path) {
		music := rl.LoadMusicStream(path)
		if music.CtxType != 0 {
			m.music[key] = music
			m.loadedMusic[key] = true
			return
		}
	}

	log.Printf("asset warning: music key=%q path=%q missing/invalid, using no-op", key, path)
	m.music[key] = rl.Music{}
	m.loadedMusic[key] = false
}

func (m *AssetManager) GetTexture(key string) rl.Texture2D {
	if texture, ok := m.textures[key]; ok {
		return texture
	}
	return rl.Texture2D{}
}

func (m *AssetManager) GetFont(key string) rl.Font {
	if font, ok := m.fonts[key]; ok {
		return font
	}
	return rl.GetFontDefault()
}

func (m *AssetManager) GetSound(key string) rl.Sound {
	if sound, ok := m.sounds[key]; ok {
		return sound
	}
	return rl.Sound{}
}

func (m *AssetManager) PlaySound(key string) {
	if m == nil || !rl.IsAudioDeviceReady() {
		return
	}
	sound, ok := m.sounds[key]
	if !ok || sound.FrameCount <= 0 {
		return
	}
	rl.PlaySound(sound)
}

func (m *AssetManager) GetMusic(key string) rl.Music {
	if music, ok := m.music[key]; ok {
		return music
	}
	return rl.Music{}
}

func (m *AssetManager) UnloadAll() {
	for key, texture := range m.textures {
		if texture.ID != 0 {
			rl.UnloadTexture(texture)
		}
		delete(m.textures, key)
	}

	for key, font := range m.fonts {
		if m.loadedFonts[key] && font.Texture.ID != 0 {
			rl.UnloadFont(font)
		}
		delete(m.fonts, key)
		delete(m.loadedFonts, key)
	}

	for key, sound := range m.sounds {
		if m.loadedSounds[key] && sound.FrameCount > 0 {
			rl.UnloadSound(sound)
		}
		delete(m.sounds, key)
		delete(m.loadedSounds, key)
	}

	for key, music := range m.music {
		if m.loadedMusic[key] && music.CtxType != 0 {
			rl.UnloadMusicStream(music)
		}
		delete(m.music, key)
		delete(m.loadedMusic, key)
	}
}

func generateFallbackTexture(width, height int32, color rl.Color) rl.Texture2D {
	image := rl.GenImageColor(int(width), int(height), color)
	defer rl.UnloadImage(image)
	return rl.LoadTextureFromImage(image)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
