# SingleFantasy - AGENTS Guide

This file is project context for AI coding agents working in this repository.

## 1. Project Summary

- Project: `SingleFantasy`
- Genre: single-player 2D isometric ARPG, inspired by Ragnarok Online class/build feel
- Core loop target: `run -> rooms -> boss -> deterministic rewards -> results -> next run`
- Current maturity: MVP in progress, architecture refactor backlog is active
- Platform constraint: **Windows only** (this is intentional and should not be generalized)

## 2. Stack and Runtime

- Language: Go `1.25.5`
- Graphics/input/audio: `github.com/gen2brain/raylib-go/raylib v0.55.1`
- Main loop: fixed update + uncapped render
  - `FixedDeltaTime = 1/60`
  - Frame clamp: `MaxFrameTime = 0.25`
  - Max fixed updates per frame: `8`

## 3. Current repo Layout

- `app/main.go`: app entrypoint, window/audio init, fixed timestep loop
- `app/game/`: game state machine and run orchestration
  - `game.go`: state transitions, main run update/draw
  - `skills_handler.go`: cast entry for skill execution paths
  - `config.go`: window, debug, and gameplay constants
- `app/gameobjects/`: runtime actor models
  - `player.go`, `enemy.go`, `boss.go`
- `app/gamedata/`: shared domain data/types (classes, skills, effects, stats, items)
- `app/systems/`: input, renderer, skill execution helpers, duplicated effects/types (technical debt exists)
- `app/world/`: dungeon and room generation structures
- `guidelines/`: design and backlog source of truth for architecture/gameplay direction
- `prompts/refine_backlog_task.md`: prompt template for refining backlog tasks
- `resources/`: art/audio assets (some references currently fallback if missing)
- `build.bat`, `build_docker.bat`, `Dockerfile.builder`: build tooling

## 4. Current Game States and Controls

State flow currently implemented:

- `Boot -> MainMenu -> ClassSelect -> Run -> Reward -> Results -> MainMenu`

Input bindings currently implemented:

- Movement: right mouse click (move-to)
- Target attack: left mouse click on enemy/boss
- Skills: `Q`, `W`, `E`, `R`
- Class select: `1`, `2`, `3` then `Enter`/`Space`
- Reward choice: `1`, `2`, `3` then `Enter`
- Stat allocation menu: `1`..`6`
- Debug overlay: `F3`

## 5. Build, Test, Run

Preferred local build:

```powershell
.\build.bat
```

Direct Go build:

```powershell
go build -o .\output\app.exe .\app
```

Tests/build check:

```powershell
go test ./...
```

Cross-build via Docker (Windows executable):

```powershell
.\build_docker.bat
```

## 6. Important Current Issues (As of 2026-03-05)

- `go build ./app` and `go test ./...` currently fail due to unused import:
  - `app/game/game.go`: `log` imported but not used.
- Audio files referenced in boot (`resources/audio/...`) are not present in repo.
  - This does not crash because `AssetManager` falls back to no-op sound/music with warning logs.
- `app/gamedata/enemies.go` is currently empty; enemy base values are still hardcoded in `app/gameobjects/enemy.go`.

## 7. Design Principles from Guidelines (Do Not Drift)

These principles are repeated across `guidelines/*.md` and should drive architecture changes:

- Entities store state.
- Systems interpret and modify state.
- Skills describe intent and data, not custom branching behavior.
- Effects should be data-driven, time-bound, and managed centrally.
- Damage math should be centralized (`DamageSpec -> damage system`), not embedded in skill logic.
- Avoid skill-name branching like `if skill.Name == "Fireball"` in core execution paths.

## 8. Gameplay/Balance Direction

Class identity target (4 active skills per class):

- Melee: Power Strike, Guard Stance, Blood Oath, Shockwave Slam
- Ranged: Quick Shot, Retreat Roll, Focused Aim, Poison Tip
- Caster: Arcane Bolt, Mana Shield, Frost Field, Arcane Drain

Global direction:

- Cooldown-driven readability over mana spam
- Sustain tied to engagement, not passive abuse
- No class should ignore positioning
- No invulnerability-based design

## 9. Architecture Reality vs Target

Current reality:

- `updateRun` in `app/game/game.go` is still monolithic and mixes input, AI, movement, projectiles, combat, room progression, and camera.
- Domain type duplication exists between `app/gamedata` and `app/systems` (skills/effects-related structs).
- No shared base entity model across player/enemy/boss yet.

Target direction (from `guidelines/current_task_refinement.md` backlog):

- Shared base entity model in a core package
- Explicit system pipeline with this order:
  - Input -> AI -> Casting -> Projectiles -> Movement -> Combat Resolve -> Effects -> Dungeon/Run -> UI/Render prep
- Centralized config/data layer under `app/gamedata` for classes/skills/enemies/items/effects
- Minimal settings persistence (volume/fullscreen/keybind display)
- Focused tests + smoke checklist

## 10. Guidance for Agents Editing This Repo

- Preserve current gameplay behavior unless the task explicitly requests behavior changes.
- Prefer incremental migrations over big-bang rewrites.
- Keep architecture work scoped to the active backlog step.
- When touching skills/effects/damage, prioritize using `app/gamedata` types over creating new duplicates.
- Keep Windows assumptions and raylib-go compatibility.
- If refactoring run logic, keep room progression, XP gain, reward flow, and boss handling behaviorally consistent.
- Add/update tests when adding data or persistence behavior.

## 11. Key Reference Files

- Design and backlog:
  - `guidelines/backlog.md`
  - `guidelines/current_task_refinement.md`
  - `guidelines/game_systems.md`
  - `guidelines/skills_system.md`
  - `guidelines/skill_effects_system.md`
  - `guidelines/skill_stat_upgrade.md`
  - `guidelines/initial_character_skills.md`
  - `guidelines/repository.md`
- Execution-critical code:
  - `app/game/game.go`
  - `app/game/skills_handler.go`
  - `app/systems/skills.go`
  - `app/gamedata/skills.go`
  - `app/gamedata/effect_system.go`
