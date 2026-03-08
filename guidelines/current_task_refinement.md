# Step 7 Refinement - Classes & Skills (Playable Set)

## Objective
Refine and complete backlog step `7) Classes & Skills (Playable Set)` so each class has a distinct, reliable, and test-backed 4-skill kit that works with the existing runtime pipeline:
`Input -> AI -> Casting -> Projectiles -> Movement -> Combat Resolve -> Effects -> Dungeon/Run -> UI/Render Prep`.

## Scope
- In scope: skill behavior completion/tuning for Melee, Ranged, Caster; skill VFX placeholders; skill SFX hooks; tests.
- Out of scope: enemy roster expansion, dungeon generation redesign, reward system changes, full UI redesign.

## Guardrails
- [ ] Preserve current state flow and run progression behavior (`Boot -> MainMenu -> ClassSelect -> Run -> Reward -> Results`).
- [ ] Preserve Windows/raylib-go assumptions.
- [ ] Do not introduce new duplicate domain models outside `app/gamedata` for skills/effects.
- [ ] Prefer data-driven skill/effect behavior; avoid adding new `if skill.Name == ...` or broad `switch skill.Type` branching in core execution paths.

## Current Baseline (Observed)
- Implemented already:
  - [x] 3 class definitions and class skill assignment exist (`app/gamedata/classes.go`, `app/gamedata/skills.go`).
  - [x] Targeting, delivery, cooldown, mana checks, projectile handling, delayed telegraph handling, and skill bar cooldown rendering exist.
  - [x] Core status effects and combat resolver exist (`app/gamedata/effect_system.go`, `app/systems/combat_resolver.go`).
- Gaps to address for step 7 completion:
  - [ ] Skill-specific behavior is still partially hardcoded in runtime (`app/game/skills_handler.go`: pre/post cast switch logic).
  - [ ] `Poison Tip` is flat DoT only; not tuned for boss HP scaling intent.
  - [ ] `Frost Field` behaves as delayed one-shot apply, not a true temporary slow zone.
  - [ ] `Mana Shield` behavior lacks explicit tunable policy/duration semantics and coverage tests.
  - [ ] No skill SFX playback hooks are wired despite asset manager support.
  - [ ] Skill VFX placeholders are generic only; no per-skill readability differentiation.
  - [ ] Step-7-focused automated tests are missing for several class-defining behaviors.

## Implementation Backlog

### 1) Skill Runtime Data Hardening
- [ ] Extend `gamedata.Skill`/specs to encode remaining runtime modifiers currently hardcoded in `app/game/skills_handler.go` (retreat displacement, mana shield policy, arcane drain mana return, optional zone params).
- [ ] Refactor `TryCastSkill` execution path to consume data/spec fields instead of adding new per-skill branching.
- [ ] Keep existing casting order and validation semantics unchanged (cooldown, mana, silence/stun rules).

Acceptance criteria:
- [ ] Adding/changing a skill behavior (within current feature set) is possible by editing `app/gamedata/skills.go` and related specs, without adding new core-path skill name/type branching.

### 2) Melee Kit Completion and Tuning
- [ ] Verify `Power Strike` functions as clear burst skill (single-target, reliable hit outcome, readable cooldown window).
- [ ] Verify `Guard Stance` applies damage reduction + mobility tradeoff and expires cleanly.
- [ ] Verify `Blood Oath` grants temporary lifesteal window and does not persist beyond effect duration.
- [ ] Verify `Shockwave Slam` remains directional AoE with slow application and proper target cap/range behavior.

Acceptance criteria:
- [ ] Melee has distinct sustain/control/burst windows without infinite sustain loops.

### 3) Ranged Kit Completion and Tuning
- [ ] Keep `Quick Shot` as responsive projectile DPS burst with readable cooldown.
- [ ] Ensure `Retreat Roll` displacement is reliable, collision-safe, and consistent with targeting intent.
- [ ] Ensure `Focused Aim` applies damage-up/move-down tradeoff for configured duration.
- [ ] Upgrade `Poison Tip` to boss-viable DoT behavior (percent-HP or capped scaling model) using data-driven effect parameters.
- [ ] Add safeguards so poison tuning cannot trivialize bosses (cap/floor and duration rules documented in code constants/spec).

Acceptance criteria:
- [ ] Ranged can pressure bosses via poison without outclassing all other damage options.

### 4) Caster Kit Completion and Tuning
- [ ] Keep `Arcane Bolt` as projectile nuke with mana/cooldown readability.
- [ ] Formalize `Mana Shield` behavior (absorb amount source, expiry condition, optional duration) in data/spec + runtime handling.
- [ ] Implement `Frost Field` as a temporary AoE control zone (not only delayed single apply), including periodic reapplication cadence.
- [ ] Keep `Arcane Drain` as close-range AoE damage with mana sustain tied to actual targets hit and proper clamping.

Acceptance criteria:
- [ ] Caster sustain requires engagement and positioning; no passive infinite safety.

### 5) Skill VFX Placeholder Pass
- [ ] Add distinct placeholder visuals per skill category (burst/projectile/zone/buff) for readability using existing renderer primitives.
- [ ] Ensure delayed and zone telegraphs are visually different from projectile impacts.
- [ ] Keep visuals lightweight and compatible with existing render queue/camera logic.

Acceptance criteria:
- [ ] During combat, each of the 12 skills is visually identifiable within ~1 second by placeholder telegraph/impact style.

### 6) Skill SFX Hook Pass
- [ ] Define minimal skill SFX keys and load attempts in boot/asset setup (fallback-safe).
- [ ] Trigger cast and/or impact SFX from casting/projectile/zone resolution points.
- [ ] Ensure missing files continue to degrade gracefully via current no-op asset behavior.

Acceptance criteria:
- [ ] Skill casts generate stable SFX hook calls with no crashes when audio files are absent.

### 7) Test Coverage for Step 7
- [ ] Add/extend unit tests for class skill definitions and essential tuning invariants in `app/gamedata`.
- [ ] Add/extend runtime tests for:
  - [ ] Retreat Roll displacement + collision-safe movement.
  - [ ] Mana Shield absorb/expiry semantics.
  - [ ] Frost Field zone application over time.
  - [ ] Poison Tip boss-scaling/cap behavior.
  - [ ] Arcane Drain mana return clamping by targets hit.
- [ ] Keep deterministic tests by controlling crit rolls and avoiding frame-rate-dependent assertions.

Acceptance criteria:
- [ ] `go test ./...` passes.
- [ ] `go test -tags raylib ./...` passes when `raylib.dll` is available in runtime path.

### 8) Final Tuning and Documentation Sync
- [ ] Perform a focused numeric tuning pass (cooldowns, magnitudes, ranges, durations) in `app/gamedata/skills.go` and related constants.
- [ ] Add a concise tuning rationale section in this file after implementation (final values + intended role per skill).
- [ ] Confirm AGENTS step-7 checklist intent is satisfied (distinct class identity, self-sufficient kits, readability-first cooldown pacing).

Acceptance criteria:
- [ ] All 12 skills are implemented, test-backed, and behaviorally aligned with step-7 class fantasy.

## Suggested Execution Order
- [ ] 1. Runtime data hardening/spec extension.
- [ ] 2. Ranged and Caster special-mechanic completion (`Poison Tip`, `Frost Field`, `Mana Shield`, `Arcane Drain`).
- [ ] 3. Melee/Ranged/Caster numeric tuning pass.
- [ ] 4. VFX placeholders.
- [ ] 5. SFX hooks.
- [ ] 6. Tests + final verification.

## Verification Commands
- [ ] `go test ./...`
- [ ] `go test -tags raylib ./...` (requires `raylib.dll` available; currently fails in this environment when missing)
- [ ] Optional manual run: `go build -o .\output\app.exe .\app` then play one run per class for feel/readability checks.
