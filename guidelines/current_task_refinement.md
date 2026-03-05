# Step 1 Refinement: Core Architecture (Systems + Data)

## Scope
Refine and complete backlog item `1) Core Architecture (Systems + Data)` from `guidelines/backlog.md`.

Source backlog items:
- Implement base `Entity` model (pos/vel, HP, stats, hitbox, effects list, faction, alive)
- Implement system scaffolding + update order:
  - Input -> AI -> Casting -> Projectiles -> Movement -> Combat Resolve -> Effects -> Dungeon/Run -> UI/Render
- Central config/data layer for classes, skills, enemies, items, effects (code tables ok for MVP)
- Save/load minimal settings (volume, fullscreen, keybind only display not configurable yet)

## Current Code Reality (Important For Planning)
- `app/game/game.go` has a large monolithic `updateRun` that mixes input, AI, movement, projectiles, combat, room flow, and camera.
- There is no shared base entity type; `Player`, `Enemy`, and `Boss` duplicate state fields.
- Effects and skill spec types are duplicated across packages (`app/gamedata/*` and `app/systems/*`).
- Class/skill/item data exists in `app/gamedata`, but enemy data is not centralized (`app/gameobjects/enemy.go` hardcodes stats).
- Settings persistence does not exist yet.

## Constraints
- Preserve current gameplay behavior while refactoring architecture.
- Keep step 1 focused on architecture and data foundation only (no feature expansion beyond what is required).
- Favor incremental migration over a risky big-bang rewrite.

## Task Backlog

### A. Base Entity Model
- [ ] A1. Add shared entity core model in new `app/core` package.
  - Add `Faction` enum (`Player`, `Enemy`, `Neutral` at minimum).
  - Add `Hitbox` model (`Width`, `Height` for current AABB usage).
  - Add `Entity` model with required minimum fields:
    - `PosX`, `PosY`
    - `VelX`, `VelY`
    - `HP`, `MaxHP`
    - `Stats` (pointer/reference to `gamedata.Stats`)
    - `Hitbox`
    - `Effects` (`[]gamedata.EffectInstance`)
    - `Faction`
    - `Alive`
- [ ] A2. Migrate `Player`, `Enemy`, and `Boss` to embed or compose the new base entity.
  - Remove duplicated fields where possible.
  - Keep domain-specific fields (mana, AI state, boss phase, etc.) in their concrete structs.
- [ ] A3. Add shared helpers for common entity operations.
  - Examples: center position, is alive check, basic damage/heal guard rails.
  - Keep helper placement consistent (either on `core.Entity` or small utility file in `app/core`).

Acceptance criteria:
- All runtime actors (`Player`, `Enemy`, `Boss`) are backed by the common entity model.
- Build passes with no duplicated effect/type definitions introduced during migration.
- Existing gameplay loop still runs (menu -> class select -> run -> reward/results).

### B. System Scaffolding And Ordered Runtime Pipeline
- [ ] B1. Create explicit runtime context object for systems.
  - Contains references currently read/written by `updateRun` (player, enemies, boss, projectiles, room/dungeon, camera, timers, inputs, transient intents).
  - Keep this in `app/game` or `app/systems/runtime` to avoid circular imports.
- [ ] B2. Create system interfaces/scaffolding.
  - Minimum: `Name() string` and `Update(ctx *RuntimeContext, dt float32)`.
  - Add a `Pipeline` runner that executes systems in fixed order.
- [ ] B3. Implement systems with the required order (even if some are thin wrappers initially):
  1. Input
  2. AI
  3. Casting
  4. Projectiles
  5. Movement
  6. Combat Resolve
  7. Effects
  8. Dungeon/Run
  9. UI/Render prep
- [ ] B4. Replace monolithic `updateRun` body with pipeline invocation.
  - `Game.updateRun` should become orchestration only.
  - Move logic blocks into corresponding systems with minimal behavior drift.
- [ ] B5. Add short system-order comments and one debug line showing active pipeline order (debug overlay only).

Acceptance criteria:
- `Game.updateRun` no longer directly implements the full run logic inline.
- Pipeline order is explicit in one place and matches backlog order exactly.
- No major behavior regression in movement, attacks, skill casts, room progression.

### C. Central Data/Config Layer
- [ ] C1. Remove duplicate gameplay type definitions from `app/systems`.
  - `systems/skills.go` should consume `gamedata` types for targeting, delivery, damage, and effects.
  - `systems/effects.go` duplicates should be removed or converted into thin wrappers that delegate to `gamedata`.
- [ ] C2. Introduce central data access entry points in `app/gamedata`.
  - Add explicit table access functions for classes, skills, enemies, items, effects.
  - Normalize read APIs so systems/game code does not reach into random constructors directly.
- [ ] C3. Add enemy data definitions to `app/gamedata` and stop hardcoding base values in `NewEnemy`.
  - Include at least normal + elite base templates used by current flow.
- [ ] C4. Keep current code-table approach (no external files required), but isolate all balancing constants in data layer.

Acceptance criteria:
- Data definitions for classes/skills/enemies/items/effects are all centralized under `app/gamedata`.
- `app/systems` no longer owns duplicated domain types already present in `gamedata`.
- Enemy constructor pulls from `gamedata` templates/config instead of hardcoded numbers.

### D. Minimal Settings Save/Load
- [ ] D1. Add settings model and defaults in new `app/settings` package.
  - Required fields:
    - `MasterVolume` (float)
    - `Fullscreen` (bool)
    - `KeybindDisplay` map/struct for labels only (read-only behavior for now)
- [ ] D2. Add JSON persistence with sane fallback behavior.
  - `Load()` reads from a local settings file and falls back to defaults if missing/invalid.
  - `Save()` writes atomically (temp file + rename) to reduce corruption risk.
- [ ] D3. Integrate settings at boot time.
  - Apply volume setting to raylib audio.
  - Apply fullscreen preference on startup.
  - Expose keybind labels to HUD rendering (display only, no rebinding input behavior).
- [ ] D4. Add minimal save trigger.
  - Save on explicit settings changes and/or graceful shutdown path in `main.go`.

Acceptance criteria:
- Restarting the game preserves fullscreen + volume + keybind display data.
- Missing/corrupt settings file does not crash startup.
- Keybind display source comes from settings model, not hardcoded literals in renderer.

### E. Safety Nets (Tests + Smoke)
- [ ] E1. Add focused unit tests for:
  - Settings load/save defaults and invalid-file fallback.
  - Core entity helper behavior (at least alive/damage guards).
  - Data layer retrieval for at least one class, one skill, and one enemy template.
- [ ] E2. Add a manual smoke checklist in this file (or adjacent note) and run it after refactor.
  - Boot to menu
  - Start run with each class
  - Basic attack and at least one skill cast
  - Clear room and progress
  - Reach reward/results
  - Close and reopen game to confirm settings persistence

Acceptance criteria:
- Added tests pass locally.
- Smoke checklist passes without crash.

## Suggested Implementation Sequence
1. A1-A3 (entity foundation)
2. C1-C4 (type/data consolidation)
3. B1-B5 (pipeline extraction)
4. D1-D4 (settings persistence)
5. E1-E2 (tests and smoke validation)

## Out Of Scope For This Refinement
- New skills, enemies, items, or content tuning beyond migration needs.
- Full key rebinding UI/logic (display only in this step).
- Deep combat redesign (only architecture extraction and behavior parity).

## Smoke Checklist (Post-Refactor)
- [ ] Boot to menu
- [ ] Start run with each class
- [ ] Basic attack and at least one skill cast
- [ ] Clear room and progress
- [ ] Reach reward/results
- [ ] Close and reopen game to confirm settings persistence

Run status:
- Automated verification completed with `go test ./...` and `go build -o .\output\app.exe .\app`.
- Manual in-game smoke steps were not executed in this non-interactive environment.
