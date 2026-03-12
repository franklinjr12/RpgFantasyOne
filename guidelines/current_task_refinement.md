# Current Task Refinement - Backlog Item 12 Combat Feedback

## Refined Scope
- Source backlog slice: `12) UI/UX (Must Be Playable and Clear) -> Combat Feedback`
- Required outcomes:
- [x] Floating damage numbers (crit styling)
- [x] Floating heal numbers
- [x] Status text popups (for example: `Stunned`, `Silenced`)
- [x] Telegraph indicators (AoE circles and line attacks)
- [x] Hit flashes

## Current Code Reality (Grounding)
- Existing and already usable:
- [x] Entity hit flashes are rendered (`DrawPlayer`, `DrawEnemy`, `DrawBoss`) via `HitFlashTimer`.
- [x] AoE telegraphs exist for delayed skills and boss mechanics (`DrawDelayedTelegraph`, boss zone/heavy telegraph rendering).
- [x] Skill cast/impact pulse visuals exist (`SkillVisualEffects`, `DrawSkillCastPulse`).
- Missing or incomplete for backlog completion:
- [x] No floating combat text system for damage/heal/status.
- [x] No crit-specific floating text styling.
- [x] No line telegraph visualization for directional attacks.
- [x] Combat feedback spawning is not centralized; damage application happens across multiple paths.

## Constraints (Do Not Drift)
- [x] Preserve current gameplay outcomes (damage, cooldowns, room flow, XP/reward behavior).
- [x] Keep runtime order intact (`Input -> AI -> Casting -> Projectiles -> Movement -> Combat Resolve -> Effects -> Dungeon/Run -> UI/Render Prep`).
- [x] Keep Windows/raylib assumptions.
- [x] Prefer `gamedata` effect/damage types; do not introduce duplicate domain structs in `systems`.

## Task Backlog

### 1) Create Combat Feedback Runtime Model
- [x] Add a dedicated combat feedback model in `app/game` for transient UI events (damage text, heal text, status text, directional telegraph visuals).
- [x] Add `Game` state fields for active feedback events and clear them in all run/reset transitions (`NewGame`, `ResetState`, `StartRun`, `AdvanceToNextRoom`, debug room load).
- [x] Add update lifecycle for feedback events in fixed update path (timer decay, upward drift, fade-out, cleanup).
- [x] Add draw lifecycle in `drawRun` after world entities and before HUD so text is readable but not hidden by terrain.
- [x] Add combat feedback tuning constants in `app/game/config.go` (durations, rise speed, alpha fade, crit scale, status text duration).

### 2) Centralize Damage/Heal Feedback Emission
- [x] Add `game`-layer wrappers around hit application so combat outcomes and UI feedback are emitted together.
- [x] Route all player->enemy/boss hit paths through wrappers:
- [x] auto-attacks in `app/game/auto_attack.go`
- [x] skill instant/projectile/delayed application paths (`skills_handler.go`, `runtime_pipeline.go`)
- [x] Route enemy/boss->player hit paths through wrapper (`ApplyPlayerCombatHit` already central; extend it for feedback emission).
- [x] Use applied HP delta (before/after) to spawn damage/heal text accurately after mitigation and shields.
- [x] Use resolver crit result to style crit damage text (size/color/prefix like `CRIT` or `!`).

### 3) Implement Floating Damage Numbers (Crit Styling)
- [x] Spawn floating damage text on every successful combat hit with `AppliedDamage > 0`.
- [x] Differentiate friendly/enemy damage colors for readability (player taking damage vs enemies taking damage).
- [x] Apply crit styling when `DamageResult.IsCrit == true` (larger scale, stronger color, optional bounce).
- [x] Prevent overlap clutter: apply small deterministic per-event horizontal jitter and stack offset.
- [x] Add deterministic tests for event spawn count/value/style flags from representative hit paths.

### 4) Implement Floating Heal Numbers
- [x] Spawn heal text on player healing from:
- [x] class/item lifesteal during combat hit resolution
- [x] ranged kill-heal bonuses in projectile kill branches
- [x] Keep displayed value clamped to effective heal gained (not attempted heal amount).
- [x] Style heal text distinctly (green palette and `+` prefix).
- [x] Add tests covering lifesteal and kill-heal popup emission.

### 5) Implement Status Text Popups
- [x] Add effect-to-label mapper in `gamedata` or `game` (for example: `EffectStun -> STUNNED`, `EffectSilence -> SILENCED`, `EffectFreeze -> FROZEN`).
- [x] On successful application of control/debuff effects, spawn short-lived status popups at target position.
- [x] Deduplicate repeated status labels per hit event (do not emit duplicate identical labels from one application batch).
- [x] Ensure both enemy targets and player target can receive status popups.
- [x] Add tests for at least stun/silence/slow popup emission and dedupe behavior.

### 6) Complete Telegraph Indicators (AoE + Line)
- [x] Keep current AoE telegraph rendering intact and move shared styling knobs to config constants.
- [x] Add directional/line telegraph visual support for directional attacks (minimum: player directional skills such as `Shockwave Slam`).
- [x] Represent line telegraph using world-space start/end derived from cast intent and targeting range.
- [x] Add renderer helper(s) in `app/systems/renderer.go` for line telegraph draw and optional endpoint marker.
- [x] Spawn directional telegraph events with short lifetime in cast feedback path (`skill_feedback.go`).

### 7) Hit Flash Completion and Validation
- [x] Audit all damage sources and confirm hit flash timer is triggered for player/enemy/boss across direct hits, projectiles, skill hits, and DoT ticks where intended.
- [x] Normalize flash durations/colors through constants to avoid hardcoded duplicates (`0.2`, etc.).
- [x] Add/adjust tests to validate flash timer activation for representative damage paths.

### 8) QA, Testing, and Backlog Closeout
- [x] Add targeted tests in `app/game` (raylib-tagged) for combat feedback event emission logic.
- [x] Add targeted tests in `app/systems` only if resolver contracts are extended (resolver contracts unchanged; no additional `app/systems` tests required).
- [ ] Validate package tests with:
- [ ] `go test ./app/game -tags raylib`
- [ ] `go test ./app/systems -tags raylib`
- [x] `go test ./app/gamedata ./app/gameobjects ./app/world`
- Note: `go test` runs for raylib-tagged packages are currently blocked in this environment because `raylib.dll` is not available at runtime.
- [ ] Perform manual smoke checklist in one run per class:
- [ ] Crit and non-crit damage numbers are visually distinct.
- [ ] Heal numbers appear for lifesteal/kill-heal.
- [ ] Stun/silence/slow popups are readable and short-lived.
- [ ] AoE and directional line telegraphs are clear and non-blocking.
- [ ] Hit flashes remain readable without overwhelming sprites.

## Definition of Done
- [ ] All five combat feedback backlog bullets are implemented and visibly verifiable in-run.
- [ ] No gameplay regressions in core loop/state transitions.
- [ ] New tests pass in supported raylib-tagged workflow.
- [ ] `guidelines/backlog.md` item `12 -> Combat Feedback` can be checked off.
