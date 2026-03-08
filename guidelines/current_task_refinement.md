# Step 6 Refinement: Status Effects and Combat Resolution

## Objective
Refine backlog step 6 into incremental code tasks that deliver:
- Effect instances with duration and optional tick behavior
- Non-stacking refresh rule
- Effect query helpers used by movement/casting/combat
- Centralized damage and combat resolution
- Core MVP effects: Slow, Stun, Freeze (full slow), Silence, Burn

## Guardrails
- Preserve current room flow, XP, reward, and boss progression behavior.
- Keep skill definitions data-driven (no branching by skill name).
- Reuse `app/gamedata` types; do not duplicate effect/damage structs in other packages.
- Keep Windows and raylib assumptions unchanged.

## Current Code Snapshot (2026-03-08)
- Runtime pipeline order already exists in `app/game/runtime_pipeline.go`.
- `gamedata.UpdateEffects` and `ApplyEffect` already exist, but effect semantics are not centralized for movement/acting/casting.
- Damage is currently split across `systems.ApplySkill`, `player.takeDamageInternal`, and `game/auto_attack.go`.
- Crit fields in `DamageSpec` exist but are not resolved in the current skill damage path.
- Enemy and boss movement/combat do not fully interpret control effects (slow/stun/freeze).
- Burn exists as data but lacks focused runtime validation tests.

## Task Backlog

### A) Effect semantics and helper API
- [x] Add a small helper API in `app/gamedata/effect_system.go` (or `app/gamedata/effect_queries.go`) for:
  - [x] `CanAct(effects)` (false on Stun)
  - [x] `CanCast(effects)` (false on Stun or Silence)
  - [x] `MoveSpeedMultiplier(effects)` (Slow and MoveSpeedReduction interactions, Freeze as full slow => multiplier `0`)
  - [x] `HasCrowdControl(effects)` (future hook)
- [x] Keep existing `HasEffect` and `GetEffectMagnitude` available for compatibility, then migrate call sites gradually.
- [x] Ensure Freeze semantics are data-driven through helper logic instead of scattered hardcoded checks.

### B) Effect lifecycle hardening
- [x] Keep MVP non-stacking behavior: reapplying the same effect type refreshes duration.
- [x] Preserve strongest magnitude on refresh for now and document this as MVP policy.
- [x] Update periodic tick handling to support `dt > tickRate` safely (no skipped ticks).
- [x] Restrict periodic damage application to intended DOT effects (Burn and Poison), avoiding accidental ticks on non-DOT effects.
- [x] Add short TODO hooks for future stack policies (count-based stacking, diminishing returns, immunity).

### C) Centralized damage resolver
- [x] Create a dedicated damage resolver in `app/systems` (example: `damage_resolver.go`) with typed request/result structs.
- [x] Move damage math from `systems.ApplySkill` into the resolver:
  - [x] stat scaling from `DamageSpec`
  - [x] mitigation by damage type (physical/magical resist, true damage bypass)
  - [x] crit chance and crit multiplier resolution
- [x] Keep current hit feedback behavior intact (player/enemy flash timing should not regress).

### D) Combat resolution as the final common path
- [x] Create a combat resolver in `app/systems` (example: `combat_resolver.go`) that applies, in order:
  - [x] damage (via damage resolver)
  - [x] status effects (`EffectSpec` to `EffectInstance`)
  - [x] on-hit hooks (class lifesteal, effect lifesteal, mana-drain style hooks)
- [x] Refactor `systems.ApplySkill` to call this resolver instead of performing inline damage/effect logic.
- [x] Ensure no skill-name branching is introduced in resolver code.

### E) Integrate resolver into all hit sources
- [x] Skill hits:
  - [x] instant, projectile, and delayed skill paths in `app/game/skills_handler.go` and `app/game/runtime_pipeline.go` use combat resolver logic
- [x] Auto attacks:
  - [x] melee/ranged/caster auto attacks in `app/game/auto_attack.go` route damage and on-hit hooks through the same resolver
- [x] Enemy and boss hits:
  - [x] `ApplyPlayerDirectHit` keeps i-frame/knockback behavior, but damage computation path is centralized
- [x] Keep XP gain, kill-heal, and target-clearing behavior unchanged.

### F) Apply effect queries to runtime behavior
- [x] Player:
  - [x] replace direct effect checks in `Game.GetPlayerMoveSpeed` with helper API
  - [x] block auto-attack windup/resolve when `CanAct == false` (stun lockout)
  - [x] keep cast lockout based on centralized `CanCast`
- [x] Enemies/Boss:
  - [x] respect slow/freeze/stun in `gameobjects.Enemy.Update` and `Enemy.Attack`
  - [x] keep boss phase logic and add-spawn behavior unchanged

### G) Core MVP effects validation
- [x] Verify these effects are fully functional in runtime and tests:
  - [x] Slow: movement reduction
  - [x] Freeze: full slow (movement zero) without separate hardcoded paths
  - [x] Stun: blocks movement and actions
  - [x] Silence: blocks skill casts only
  - [x] Burn: periodic damage over time
- [x] Keep Poison behavior intact while implementing Burn checks.

### H) Tests and smoke checklist
- [x] Add or expand unit tests in `app/gamedata` for:
  - [x] effect refresh policy
  - [x] periodic tick correctness with large `dt`
  - [x] helper query semantics (`CanAct`, `CanCast`, `MoveSpeedMultiplier`)
- [x] Add or expand tests in `app/game` and/or `app/systems` (with `//go:build raylib` when needed) for:
  - [x] stun blocks player actions
  - [x] silence blocks casting but not movement
  - [x] freeze sets effective movement to zero
  - [x] burn ticks and expires correctly
  - [x] skill/projectile/auto-attack paths share combat resolution behavior
- [x] Run `go test ./...` and document tagged test commands separately if raylib-tagged tests are added.

## Definition of Done
- [x] All Step 6 backlog bullets are implemented without changing run flow state transitions.
- [x] Damage/effect application has one primary shared path (no duplicate combat math spread across files).
- [x] Core MVP effects are behaviorally correct and covered by tests.
- [x] `go test ./...` passes.

## Verification Notes
- Default test suite: `go test ./...` (pass).
- Raylib-tagged suite command: `go test -tags raylib ./...` (fails in current environment because `raylib.dll` is missing).
