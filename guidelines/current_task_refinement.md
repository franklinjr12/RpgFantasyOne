# Step 2 Refinement: Isometric World and Camera

Source backlog step: `2) Isometric World & Camera` from `guidelines/backlog.md`.

## Objective
Implement a usable isometric presentation layer for the run state while keeping gameplay logic in world-space and preserving current room progression/combat behavior. Deliver camera smoothing, depth-correct rendering, room wall and obstacle AABB collision, and placeholder floor/door visuals with a basic door transition flow.

## Scope Constraints
- Keep Windows and `raylib-go` compatibility.
- Keep the runtime pipeline order unchanged.
- Keep existing core loop behavior (run -> rooms -> boss -> reward -> results).
- Do not expand into Step 3 movement feel tuning (acceleration/facing polish) beyond what is required for collision correctness.
- Use placeholders (flat colors/shapes) for floor, obstacles, and door visuals.

## Current State Snapshot (Code Reality)
- Rendering is top-down world-to-screen translation, not isometric projection (`app/systems/renderer.go`).
- `IsometricToScreen` and `ScreenToIsometric` exist but are not integrated into render/input flow.
- Camera snaps directly to player without smoothing (`systems.UpdateCamera`), clamped to world bounds.
- Player movement collision is only room-edge clamping in `movementSystem.Update` (`app/game/runtime_pipeline.go`).
- No room obstacle or door data model in `app/world/room.go`.
- Room transition is immediate on clear via `CheckRoomCompletion` + `AdvanceToNextRoom`, not door-driven.

## Refined Task Backlog

### [x] T2.1 Add a shared isometric projection layer used by render and input
Goal: one authoritative conversion path between world-space and screen-space.

Implementation:
- Add/normalize helpers in `app/systems/renderer.go` or a new focused file (example: `app/systems/projection.go`):
  - `WorldToScreenIso(worldX, worldY, camera)`
  - `ScreenToWorldIso(screenX, screenY, camera)`
  - optional helper for projected depth key (Y-sort basis).
- Keep gameplay simulation in world coordinates; only view/input mapping changes.
- Replace direct usage of current top-down transforms in:
  - `app/systems/input.go`
  - `app/game/runtime_pipeline.go` (attack target pick)
  - `app/game/skills_handler.go` (cursor targeting)
  - draw paths in `app/systems/renderer.go`

Acceptance criteria:
- Right-click move and left-click target selection still work after projection switch.
- Skill targeting still aligns with cursor position in run state.
- No duplicate conversion math outside the shared helpers.

### [x] T2.2 Implement consistent depth sorting for world objects
Goal: prevent incorrect overlap when entities cross in isometric view.

Implementation:
- Build a render queue for run-world drawables (player, enemies, boss, projectiles, optional obstacle overlays).
- Sort by a stable depth key derived from world position (typically world Y plus optional tie-breaker on X or entity id).
- Keep UI/HUD drawing after world queue.
- Update `Game.drawRun()` (`app/game/game.go`) to use the sorted draw path.

Acceptance criteria:
- Player/enemy/boss overlap order changes correctly when moving up/down the map.
- Sorting is stable (no frame-to-frame flicker when depth is equal).

### [x] T2.3 Upgrade camera follow to smoothed motion with safe clamping
Goal: camera follows player with small smoothing instead of hard snap.

Implementation:
- Extend camera state (`app/systems/renderer.go`): smoothing factor and optional velocity/target fields.
- Update `systems.UpdateCamera(...)` to interpolate toward target each fixed update.
- Keep clamping so camera does not expose outside playable world bounds.
- Ensure compatibility with isometric projection math introduced in T2.1.

Acceptance criteria:
- Camera visibly lags slightly and catches up smoothly.
- No jitter when player is idle.
- Camera remains bounded across full dungeon length.

### [x] T2.4 Add room obstacle and door data structures in world model
Goal: define collision/transition primitives at room level.

Implementation:
- Add simple AABB types for world geometry in `app/world/room.go` (or `app/world/collision.go`):
  - obstacle rectangle list per room
  - door rectangle list per room with minimal state (`locked/open`, target room index or direction).
- Generate placeholder obstacles/doors when creating rooms (`NewRoom` / `NewDungeon`).
- Ensure obstacle placement does not block room spawn center and is within room bounds.

Acceptance criteria:
- Each non-boss room has at least one door definition for progression.
- Obstacles are deterministic enough for testing and never outside room bounds.

### [x] T2.5 Implement AABB collision resolution for player vs room walls/obstacles
Goal: replace simple edge clamp with collision-resolved movement constraints.

Implementation:
- Add collision helpers in `app/systems` (example: `collision.go`):
  - AABB overlap check
  - movement resolution against room bounds and obstacle list.
- Integrate into `movementSystem.Update` in `app/game/runtime_pipeline.go`.
- Keep current click-to-move flow; adjust final position only through collision resolution.

Acceptance criteria:
- Player cannot leave room bounds.
- Player cannot pass through obstacles.
- Movement remains smooth while sliding around obstacle edges.

### [x] T2.6 Render basic isometric floor/room/obstacle placeholders
Goal: make rooms read as isometric spaces with minimal art.

Implementation:
- Update room drawing in `app/systems/renderer.go`:
  - render floor as isometric-projected shapes/tiles (flat color placeholders).
  - render room border/walls in a readable way.
  - render obstacle placeholders from room obstacle data.
- Keep boss-room visual distinction with alternate colors.

Acceptance criteria:
- Room floor, walls, and obstacles are visible and readable in isometric view.
- Visuals remain performant and deterministic under current room count.

### [x] T2.7 Add placeholder door visuals and transition trigger
Goal: support basic door transitions tied to room progression.

Implementation:
- Render door placeholders (locked vs open state) in `app/systems/renderer.go`.
- In run logic (`app/game/runtime_pipeline.go` and/or `app/game/game.go`):
  - lock doors while room objective is incomplete.
  - once room is cleared, allow transition when player enters door AABB.
  - transition to next room and reposition player at entry point.
- Add a minimal transition effect (short fade or short lockout timer) using placeholder visuals.

Acceptance criteria:
- Room does not advance before clear.
- After clear, entering door transitions once (no double-trigger).
- Boss room flow still reaches reward state correctly.

### [x] T2.8 Add focused automated tests for projection and collision primitives
Goal: protect new math and collision logic from regressions.

Implementation:
- Add tests in `app/systems` and `app/world` for pure logic:
  - projection round-trip tolerance (`world -> screen -> world`).
  - depth sort ordering helper behavior.
  - AABB collision resolution against bounds/obstacles.
  - room obstacle/door generation invariants.
- Keep tests independent from raylib runtime where possible.

Acceptance criteria:
- `go test ./...` passes.
- New tests fail if projection/collision helpers regress.

## Suggested Execution Order
1. T2.1
2. T2.3
3. T2.4
4. T2.5
5. T2.6
6. T2.2
7. T2.7
8. T2.8

## Verification Checklist (End of Step)
- `go test ./...`
- `go build -o .\\output\\app.exe .\\app`
- Manual smoke in run state:
  - click-to-move reaches intended world positions under isometric projection.
  - camera smoothly follows player.
  - collision blocks room edges and obstacles.
  - depth ordering is visually correct when entities overlap.
  - doors lock until clear and transition correctly after clear.

## Out of Scope for This Step
- Full movement feel tuning (acceleration/facing polish), except collision correctness.
- Minimap door UX and advanced transition polish (handled in later backlog steps).
- Final art assets, VFX polish, or audio pass.
