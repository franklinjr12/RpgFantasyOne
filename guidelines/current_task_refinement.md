# Step 3 Refinement: Player Controller and Feel

## Goal
Refine backlog step `3) Player Controller & Feel` into implementation-ready tasks that improve control readability and combat feel while preserving current run flow and class behavior.

## Scope Guardrails
- Keep Windows + raylib-go assumptions.
- Keep runtime update order in `app/game/runtime_pipeline.go`.
- Preserve room progression, XP gain, reward flow, and boss handling.
- Keep right click move and left click target attack controls.
- Do not start broad skill-system refactors in this step.

## Current State Snapshot (from code)
- Right click movement with click-to-move target exists in `inputSystem` + `movementSystem`.
- Player movement is instant speed (no acceleration curve), with collision handled by `systems.ResolvePlayerMovement`.
- Enemy has facing state, but player facing + sprite flip is not implemented.
- Auto attack exists per class in `Game.UpdateAutoAttack`, but damage/projectile is applied immediately when cooldown is ready.
- Player hit flash exists, but player iframes and knockback are not implemented.

## Implementation Backlog
- [x] 3.1 Add controller/feel tuning constants in `app/game/config.go`.
- [x] 3.2 Define player feel state in `app/gameobjects/player.go`:
  Add fields for facing, movement velocity smoothing, attack timing state, hurt iframe timer, and knockback velocity.
- [x] 3.3 Refactor `movementSystem.Update` in `app/game/runtime_pipeline.go` to use light acceleration/deceleration:
  Keep click-to-move target logic, apply acceleration toward desired velocity, apply deceleration near stop, and keep collision resolution.
- [x] 3.4 Ensure movement state and collision stay coherent:
  When collision blocks one axis, clear or damp that velocity axis to prevent jitter/sticking on walls and obstacles.
- [x] 3.5 Implement player facing updates:
  Update facing from movement and attack target direction in runtime systems so facing remains stable when idle.
- [x] 3.6 Implement sprite flip for player in `app/systems/renderer.go`:
  Keep same spritesheet rows/cols, flip horizontally when facing left, no asset changes.
- [x] 3.7 Introduce basic attack hit timing state machine in `app/game/game.go` and/or runtime systems:
  Start attack, wait windup, apply hit at hit frame, then recover/cooldown. Keep existing class-specific damage rules and XP/lifesteal outcomes.
- [x] 3.8 Preserve current class attack identity with timing:
  Melee hit at contact timing, ranged/caster spawn projectile or cast at hit frame (not immediately on click).
- [x] 3.9 Add player hurt iframes in `app/gameobjects/player.go`:
  Direct hits during active iframe should not re-apply normal hit damage.
- [x] 3.10 Add light readable knockback on player hit:
  Apply short knockback impulse from attacker/projectile direction and decay over time in movement update.
- [x] 3.11 Keep hit feedback readable:
  Retain hit flash and tune timer values so hit, iframe window, and knockback do not feel invulnerable or floaty.
- [x] 3.12 Add focused tests for feel logic:
  Unit tests for movement smoothing helper(s), iframe damage gating, and knockback decay/collision interaction where practical.
- [x] 3.13 Run validation commands and resolve regressions:
  `go test ./...` and `go build -o .\output\app.exe .\app`.

## Suggested Task Order for Implementation Agent
- [x] A. Movement foundation first: `3.1` to `3.4`.
- [x] B. Facing and rendering: `3.5` to `3.6`.
- [x] C. Attack timing and class behavior parity: `3.7` to `3.8`.
- [x] D. Hurt response and feedback: `3.9` to `3.11`.
- [x] E. Tests and final validation: `3.12` to `3.13`.

## Acceptance Criteria
- [ ] Right click movement feels smoother than instant start/stop and still respects room/obstacle collision.
- [ ] Player sprite flips correctly left/right and does not flicker when nearly stationary.
- [ ] Left click basic attacks have visible hit timing (windup -> hit -> recover), not instant application.
- [ ] Melee/ranged/caster basic attacks still work with the same core class intent and progression rewards.
- [ ] Consecutive enemy hits are gated by short iframes; player still takes damage when iframe expires.
- [ ] Knockback is visible but small, does not break room progression, and resolves safely with collision.
- [ ] No regressions in room transitions, boss flow, level up menu gating, or reward entry.

## Manual Smoke Checklist
- [ ] Start each class and verify right click movement response in open space and near obstacles.
- [ ] Verify player sprite orientation while moving, attacking, and idle.
- [ ] Click an enemy at max range and confirm chase -> timed attack -> damage application.
- [ ] Stand in enemy melee range and verify iframe behavior prevents rapid multi-hit spikes.
- [ ] Verify knockback does not push player out of room bounds or through obstacles.
- [ ] Clear a normal room, transition through door, and complete boss room to reward screen.

## Out of Scope for Step 3
- [ ] Full animation system or animation blending.
- [ ] New enemy archetypes or AI state machine redesign.
- [ ] Skill/effect architecture migration beyond what is required for basic attack timing.
- [ ] UI redesign outside required hit feedback readability.
