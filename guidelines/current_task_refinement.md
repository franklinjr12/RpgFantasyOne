
# Step 9 Refinement - Dungeon Generation and Room Flow (Replayable Runs)

## Goal
Implement a template-driven dungeon flow for Biome 1 (forest) using:
- ASCII `.layout` room files for geometry and markers.
- `.meta.json` files for room metadata and selection rules.
- Procedural run assembly (8-12 rooms, 1-2 event rooms, boss at end).
- Deterministic generation with replayable variation.
- Room completion rules and door transition UX integrated into current runtime pipeline.

## Scope Boundaries
- In scope: room templates, loader/validation, template selection, room graph/flow, completion rules, door transitions, minimap/debug updates, tests, sample room files.
- Out of scope: boss AI redesign (step 10), reward/item redesign (step 11), major render style overhaul, multi-biome production content.

## Current Code Reality (Read Before Implementing)
- `app/world/dungeon.go` currently builds a fixed linear chain: 5 normal + 1 boss.
- `app/world/room.go` currently creates random rectangle rooms with procedural obstacles/enemy refs.
- `app/game/runtime_pipeline.go` handles room clear checks, door lock/unlock, transition timer, and room advance.
- `app/game/ui_hud.go` already has a minimap and room index HUD.
- `app/systems/renderer.go` already draws rooms, obstacles, and doors in isometric projection.

## Implementation Backlog

### 1) Room Template Domain and Parsing
- [x] Add room template domain types in `app/world` (or `app/gamedata` if shared broadly):
- [x] `RoomTemplate`, `DoorMarker`, `SpawnMarker`, `PropMarker`, `EventMarker`, `RoomType`, `DoorDirection`, `SpawnType`, `TileType`.
- [x] Add metadata DTO + parsing structs for `.meta.json`.
- [x] Implement `.layout` parser:
- [x] Enforce rectangular grid.
- [x] Enforce supported character legend.
- [x] Enforce bounds (recommended 8x8 min, 30x30 max via constants).
- [x] Convert ASCII to tile grid + marker lists.
- [x] Implement metadata loader and validation:
- [x] Required fields: `id`, `biome`, `type`, `doors`.
- [x] Optional/defaulted fields: `difficulty`, `weight`, `allow_rotation`, `tags`.
- [x] Validate door coordinates in bounds.
- [x] Validate metadata doors match `D` markers in layout.
- [x] Fail hard on invalid template pairs at load time.
- [x] Add unit tests for parser + validator edge cases.

### 2) Template Registry and Startup Loading
- [x] Add template registry loader that scans `assets/rooms/<biome>/`.
- [x] Pair `*.layout` with matching `*.meta.json` by base filename.
- [x] Return actionable errors for missing pair files.
- [x] Make load deterministic (stable file ordering before parse).
- [x] Expose query API by biome/type/tags/difficulty for generator use.
- [x] Add tests for directory scan/pairing and deterministic registry ordering.

### 3) Rotation and Door Compatibility
- [x] Implement optional room rotation for templates where `allow_rotation=true`:
- [x] Support 0, 90, 180, 270 degrees.
- [x] Rotate tile grid, doors, and all marker coordinates consistently.
- [x] Add door direction rotation mapping (`north/east/south/west`).
- [x] Add tests that verify rotated coordinates and door directions.
- [x] Add door compatibility helpers:
- [x] Compatibility pairs: east<->west, north<->south.
- [x] Coordinate alignment checks for connected doors.

### 4) Run Shape and Procedural Selection
- [x] Introduce a dungeon generation config struct (seed, biome, run length range, event count range).
- [x] Replace fixed `DungeonLength` chain with generated room sequence:
- [x] Start room first.
- [x] Combat-heavy middle.
- [x] 1-2 event rooms inserted in middle slots.
- [x] At least 1 elite room before boss.
- [x] Boss room last.
- [x] Implement weighted template selection filtered by:
- [x] Biome.
- [x] Room type.
- [x] Difficulty band by progression index.
- [x] Required tags (if used by generation rule).
- [x] Door compatibility.
- [x] Prevent immediate template repeats.
- [x] Preserve deterministic generation for same seed.
- [x] Add generator tests for sequence invariants and determinism.

### 5) Runtime Room Instantiation (Template -> `world.Room`)
- [x] Add conversion from `RoomTemplate` to runtime `world.Room` while keeping current systems compatible.
- [x] Compute room world bounds from tile dimensions and tile size constants.
- [x] Build `Room.Obstacles` from wall/hazard/trap tiles used for collision.
- [x] Build room doors from template door markers (with lock defaults).
- [x] Build enemy spawn refs from `s`, `E`, `B` markers:
- [x] Normal rooms: map `s` to spawn director assignments.
- [x] Elite rooms: ensure at least one elite assignment from `E` if present.
- [x] Boss room: consume `B` marker as boss spawn anchor.
- [x] Keep existing flow fallback when markers are missing (do not panic).

### 6) Room Completion Rules and Flow
- [x] Add explicit completion rule per room type:
- [x] `combat`: kill all alive enemies.
- [x] `elite`: kill elite target(s), then clear leftovers or auto-complete per chosen rule.
- [x] `event`: support at least one non-kill rule (survive timer or interaction marker).
- [x] `boss`: boss death completes room.
- [x] Keep door lock/unlock logic centralized in dungeon/run runtime system.
- [x] Ensure transition continues to use existing fade timer and respects input lock during transition.
- [ ] Add tests for completion rules and transition triggers.

### 7) UX and Debug Visibility
- [x] Update debug overlay with template-focused data:
- [x] Current template ID.
- [x] Room type.
- [x] Rotation.
- [x] Remaining enemies / completion status.
- [x] Update minimap coloring/icons by room type (`start/combat/elite/event/boss`).
- [x] Add optional debug quick-load hook for a room template ID (dev-only path).

### 8) Integration and Migration Safety
- [x] Keep `RuntimePipeline` system order unchanged.
- [x] Avoid behavior regressions in XP gain, reward state entry, and boss clear flow.
- [x] Keep Windows-only assumptions and raylib compatibility.
- [x] Do not introduce skill-name branching in dungeon flow logic.

### 9) Test Plan and Acceptance
- [x] Unit tests: parsing, validation, rotation, selection filters, compatibility.
- [x] World tests: deterministic dungeon generation, no nil rooms, valid door targets.
- [ ] Runtime tests: clear rules, lock/unlock transitions, boss-to-reward path.
- [ ] Smoke check (manual):
- [ ] Start run from class select.
- [ ] Progress through mixed room types.
- [ ] Event room appears in run (when configured).
- [ ] Boss room is final.
- [ ] Reward screen appears after boss clear.
- [ ] Minimap and debug overlay reflect room progression.

## Suggested File Touch Map (For Implementing Agent)
- `app/world/dungeon.go`
- `app/world/room.go`
- `app/world/*_test.go` (expand existing tests)
- `app/game/game.go`
- `app/game/runtime_pipeline.go`
- `app/game/ui_hud.go`
- `app/systems/renderer.go` (only if room visual markers/debug are added)
- New files likely in `app/world/`:
- `room_template.go`
- `room_template_loader.go`
- `room_template_parser.go`
- `room_template_registry.go`
- `room_template_rotation.go`

## Seed Content Added In This Refinement Pass
- [x] Added starter sample templates under `assets/rooms/forest`:
- [x] `forest_start_01`
- [x] `forest_combat_small_01`
- [x] `forest_combat_open_02`
- [x] `forest_elite_01`
- [x] `forest_event_shrine_01`
- [x] `forest_boss_01`
