package world

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const DefaultRoomsRoot = "assets/rooms"

type RoomTemplateRegistry struct {
	Templates []*RoomTemplate
	byID      map[string]*RoomTemplate
}

type TemplateQuery struct {
	Biome         string
	RoomType      RoomType
	MinDifficulty int
	MaxDifficulty int
	RequiredTags  []string
	RequiredDoors []DoorDirection
}

func LoadRoomTemplateRegistry(root string) (*RoomTemplateRegistry, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		root = DefaultRoomsRoot
	}
	root = resolveRoomsRoot(root)

	rootEntries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("scan room template root %q: %w", root, err)
	}

	biomeDirs := make([]string, 0, len(rootEntries))
	for _, entry := range rootEntries {
		if entry.IsDir() {
			biomeDirs = append(biomeDirs, entry.Name())
		}
	}
	sort.Strings(biomeDirs)

	templates := make([]*RoomTemplate, 0, 32)
	byID := map[string]*RoomTemplate{}

	for _, biome := range biomeDirs {
		biomePath := filepath.Join(root, biome)
		files, err := os.ReadDir(biomePath)
		if err != nil {
			return nil, fmt.Errorf("scan biome directory %q: %w", biomePath, err)
		}

		layoutBases := map[string]string{}
		metaBases := map[string]string{}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			name := file.Name()
			fullPath := filepath.Join(biomePath, name)
			if strings.HasSuffix(name, ".layout") {
				base := strings.TrimSuffix(name, ".layout")
				layoutBases[base] = fullPath
				continue
			}
			if strings.HasSuffix(name, ".meta.json") {
				base := strings.TrimSuffix(name, ".meta.json")
				metaBases[base] = fullPath
			}
		}

		keys := map[string]struct{}{}
		for base := range layoutBases {
			keys[base] = struct{}{}
		}
		for base := range metaBases {
			keys[base] = struct{}{}
		}

		bases := make([]string, 0, len(keys))
		for base := range keys {
			bases = append(bases, base)
		}
		sort.Strings(bases)

		for _, base := range bases {
			layoutPath, hasLayout := layoutBases[base]
			metaPath, hasMeta := metaBases[base]
			if !hasLayout {
				return nil, fmt.Errorf("room template %q in biome %q missing .layout file", base, biome)
			}
			if !hasMeta {
				return nil, fmt.Errorf("room template %q in biome %q missing .meta.json file", base, biome)
			}

			template, err := loadRoomTemplatePair(layoutPath, metaPath)
			if err != nil {
				return nil, err
			}
			if _, exists := byID[template.ID]; exists {
				return nil, fmt.Errorf("duplicate room template id %q", template.ID)
			}

			templates = append(templates, template)
			byID[template.ID] = template
		}
	}

	return &RoomTemplateRegistry{
		Templates: templates,
		byID:      byID,
	}, nil
}

func (r *RoomTemplateRegistry) GetByID(id string) *RoomTemplate {
	if r == nil {
		return nil
	}
	return r.byID[id]
}

func (r *RoomTemplateRegistry) Query(query TemplateQuery) []*RoomTemplate {
	if r == nil {
		return nil
	}

	biome := strings.ToLower(strings.TrimSpace(query.Biome))
	result := make([]*RoomTemplate, 0, len(r.Templates))

	for _, template := range r.Templates {
		if template == nil {
			continue
		}
		if biome != "" && template.Biome != biome {
			continue
		}
		if template.Type != query.RoomType {
			continue
		}
		if query.MinDifficulty > 0 && template.Difficulty < query.MinDifficulty {
			continue
		}
		if query.MaxDifficulty > 0 && template.Difficulty > query.MaxDifficulty {
			continue
		}
		if !templateHasTags(template, query.RequiredTags) {
			continue
		}
		if !templateHasDoors(template, query.RequiredDoors) {
			continue
		}
		result = append(result, template)
	}

	return result
}

func templateHasTags(template *RoomTemplate, required []string) bool {
	for _, tag := range required {
		if !template.HasTag(tag) {
			return false
		}
	}
	return true
}

func templateHasDoors(template *RoomTemplate, required []DoorDirection) bool {
	for _, direction := range required {
		if !template.HasDoorDirection(direction) {
			return false
		}
	}
	return true
}

func resolveRoomsRoot(root string) string {
	candidates := []string{
		root,
		filepath.Join("..", root),
		filepath.Join("..", "..", root),
		filepath.Join("..", "..", "..", root),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return root
}
