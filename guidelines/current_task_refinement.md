# Step 5 Refinement - Skill System (Casting + Targeting + Delivery)

## Goal
Complete backlog step 5 by finishing missing skill pipeline capabilities in a data-driven way while preserving current gameplay feel for existing skills.

## Scope Guardrails
- Keep runtime order unchanged: `Input -> AI -> Casting -> Projectiles -> Movement -> Combat Resolve -> Effects -> Dungeon/Run -> UI/Render Prep`.
- Preserve behavior of current class skills unless a checklist item explicitly calls for a behavior change.
- Prefer extending `app/gamedata` specs and shared systems over adding skill-specific branches in core casting paths.
- Keep Windows-only and raylib-go compatibility.

## Current Baseline (Already Present)
- Skill data model already includes cooldown, mana cost, targeting spec, delivery spec, damage spec, and effect specs.
- Skill cooldown ticking, mana spending/regeneration, and cast validation (`cooldown/mana/silence/stun`) already exist.
- Self/enemy/area targeting and instant/projectile delivery exist.
- Skill bar and skill key input currently exist with `Q/W/E/R`.

## Gaps This Refinement Must Close
- Directional targeting is not implemented.
- Single-target resolution is not cursor-prioritized / nearest fallback based on intent.
- Delayed AoE delivery with telegraph is not implemented.
- Player projectiles have no explicit lifetime/pierce model.
- Step requirement mentions `1..4` skill input; code currently supports only `Q/W/E/R`.
- Optional global cooldown/cast lockout support is missing.

## Expected Touch Points
- `app/gamedata/skill_specs.go`
- `app/gamedata/skills.go`
- `app/systems/skills.go`
- `app/systems/input.go`
- `app/systems/renderer.go`
- `app/game/skills_handler.go`
- `app/game/runtime_pipeline.go`
- `app/game/game.go`
- new tests in `app/systems` and/or `app/game` (plus `app/gamedata` if needed)

## Implementation Backlog (Mandatory)

### 1) Targeting Data and Intent Contract
- [x] Extend targeting spec to support directional targeting parameters needed for MVP cone/line behavior (for example: angle/width and forward range).
- [x] Add/standardize cast intent data passed into targeting/delivery (cursor world position and directional vector), so targeting is not inferred ad hoc in multiple places.
- [x] Ensure zero-values keep existing self/enemy/area skills behaviorally unchanged.

### 2) Target Resolution System
- [x] Refactor target resolution so `TargetSelf`, `TargetEnemy`, `TargetArea`, and new directional targeting all run through one resolver.
- [x] Implement single-target behavior: prefer hovered/clicked enemy under cursor if valid; otherwise fallback to nearest valid target in range.
- [x] Keep boss participation consistent with enemy targeting rules.
- [x] Make target selection deterministic when distances tie (stable ordering).
- [x] Add unit tests for self/single(area intent)/nearest fallback/area/directional resolution.

### 3) Delivery System Completion
- [x] Centralize delivery handling by `Delivery.Type` (`Instant`, `Projectile`, `Delayed`) in one execution path.
- [x] Keep instant delivery applying through shared `ApplySkill` logic.
- [x] Implement delayed delivery runtime object(s): delay timer, target point, radius, owning skill/caster, one-shot application on expiry.
- [x] Add simple telegraph rendering for delayed AoE instances (placeholder circle is enough for MVP).
- [x] Ensure delayed events are cleaned up safely on room/reset transitions.

### 4) Projectile System Completion
- [x] Extend player projectile runtime data to include lifetime and optional pierce count.
- [x] Decrement lifetime in projectile updates and expire on timeout/out-of-room bounds.
- [x] Keep collision behavior unified so hit processing calls shared skill application path.
- [x] Implement optional pierce behavior (if pierce > 0 continue; else despawn).
- [x] Add unit tests for projectile lifetime expiry, collision apply, and pierce behavior.

### 5) Input + Cast Validation Finalization
- [x] Support skill-slot numeric aliases (`1..4`) in addition to existing `Q/W/E/R` inputs (do not remove `Q/W/E/R` defaults).
- [x] Keep cast validation centralized (cooldown, resource, silence, stun) and prevent duplicate checks across call sites.
- [x] Add tests for cast validation edge cases (insufficient mana, active cooldown, silence/stun blocked casts).

### 6) Skill Data Wiring
- [x] Populate/normalize `Delivery` fields (`Speed`, `Delay`, `Lifetime`, optional pierce-related value if introduced) in skill definitions where relevant.
- [x] Preserve current behavior of existing skills; if any skill is intentionally moved to delayed/directional delivery, document the intentional change in this file during implementation.
- [x] Remove or isolate ad hoc skill-type branching in core casting flow where feasible within this step, without breaking current skill outcomes.

### 7) Verification and Done Gates
- [x] `go test ./...` passes.
- [x] `go build -o .\\output\\app.exe .\\app` passes.
- [ ] Manual smoke checklist completed:
- [ ] Cast all 4 skills for each class without runtime errors.
- [ ] Projectile skills expire by hit, bounds, or lifetime (no lingering dead projectiles).
- [ ] Delayed AoE telegraph appears and applies at delay expiry.
- [ ] Directional targeting affects only targets in front of caster.

## Optional Enhancements (Do Not Block Step 5)
- [ ] Add configurable global cooldown (default OFF) applied across all skill casts.
- [ ] Add configurable cast lockout window (default OFF) for readability after cast start.
- [ ] Add tests for optional GCD/lockout behavior when enabled.

## Step Completion Rule
Step 5 is complete when all mandatory checkboxes are checked and optional items are either completed or explicitly left disabled with rationale in implementation notes.

## Implementation Notes (2026-03-06)
- Intentional behavior change: `Shockwave Slam` now uses directional targeting (`TargetDirection`) with a forward cone.
- Intentional behavior change: `Frost Field` now uses delayed ground-targeted delivery (`DeliveryDelayed`) with cast range and telegraphed delay.
- Projectile skills now define explicit `Lifetime`, and `Poison Tip` is configured with `Pierce: 1`.
- Added focused tests under `app/systems/skills_test.go` and `app/game/projectiles_system_test.go` with `//go:build raylib` so they can run in raylib-enabled environments.
