# Step 4 Refinement - Stats, Leveling, and Build Core

## Goal
Implement backlog step `4) Stats, Leveling, and Build Core` in a way that preserves the current run loop and controls while making stats and leveling ready for later combat and item expansion.

## Scope and Constraints
- Keep Windows + raylib-go assumptions unchanged.
- Preserve current gameplay flow: `Run -> rooms -> boss -> Reward -> Results`.
- Keep existing stat allocation inputs (`1..6`) and level-up menu behavior as the base UX.
- Do not redesign the full combat damage pipeline from step 6; only prepare and integrate stat/level foundations needed by step 4.

## Current Code Snapshot (for implementer)
- Six stats already exist in `app/gamedata/stats.go` and are used by `Player.ApplyStats()`.
- XP and level-up flow already exists in `app/gameobjects/player.go` (`GainXP`, `AddStatPoint`) and level-up UI wiring exists in `app/game/runtime_pipeline.go`.
- Class data exists in `app/gamedata/classes.go`, but class-specific starting stats and growth bias are not defined.
- Derived stats are partially computed ad hoc (damage, hp, mana, move speed, attack speed), crit/resist are not fully integrated.

## Implementation Backlog (Ordered)

### 1) Class Baselines Data
- [x] Add class baseline stat profiles to `app/gamedata/classes.go` (starting STR/AGI/VIT/INT/DEX/LUK per class).
- [x] Add class growth bias metadata to `Class` (at minimum one deterministic growth stat per class).
- [x] Define explicit baseline values for MVP identity:
- [x] `Melee`: high STR/VIT, low INT.
- [x] `Ranged`: high DEX/AGI, moderate LUK.
- [x] `Caster`: high INT, moderate VIT/DEX, low STR.
- [x] Add/extend data access helpers so callers can fetch class baseline + growth info without touching `classTable` directly.

### 2) Stats and Derived Stats Core
- [x] Introduce a single derived-stats model in `app/gamedata` (or `app/core` if more appropriate) that includes: base attack output, attack speed multiplier, move speed, crit chance, physical resist, magical resist, max HP, max mana.
- [x] Move formula ownership into one place (no duplicated stat math across `Player` methods).
- [x] Keep current effective formulas where already established (hp, mana, move speed, attack speed) unless required for consistency.
- [x] Add crit and resist formulas with clamps/caps to avoid invalid values.
- [x] Add a helper to compute effective stats from base stats + equipped item bonuses before derived calculations.

### 3) Player Integration
- [x] Update `app/gameobjects/player.go` to initialize stats from selected class baseline instead of uniform `NewStats()` defaults.
- [x] Refactor `ApplyStats`, `GetAttackCooldown`, and `GetAutoAttackDamage` to use shared effective/derived computation.
- [x] Ensure equipping items and spending stat points always triggers a consistent derived stats refresh.
- [x] Keep current class auto-attack identity behavior (melee physical focus, ranged dex focus, caster int focus) while sourcing numbers from shared derived stats.

### 4) Leveling and Growth Bias
- [x] Centralize progression constants (xp curve, stat points per level, optional baseline values) in a data/config location instead of hardcoding in `GainXP`.
- [x] Extend `GainXP` to apply class growth bias deterministically on level-up.
- [x] Preserve manual allocation points each level (current behavior is 3 points/level) unless explicitly re-tuned in config.
- [x] Guard edge cases in leveling flow: ignore non-positive XP grants, support multi-level gains in one XP event, and keep XP carry-over behavior explicit.

### 5) UI/HUD Updates for Step 4
- [x] Keep existing level-up menu controls (`1..6`) and add clear on-screen feedback for unspent stat points during run HUD.
- [x] Show class baseline/growth identity in class select or level-up panel (short text is enough for MVP).
- [x] Ensure level-up menu still blocks gameplay actions as it does today.

### 6) Tests
- [x] Add unit tests for stats/derived formulas in `app/gamedata` (including crit/resist clamping and effective stats from equipment bonuses).
- [x] Add unit tests for player progression in `app/gameobjects`:
- [x] class baseline initialization by selected class.
- [x] XP leveling loop behavior (single and multi-level cases).
- [x] stat points grant and growth bias application.
- [x] Add regression tests for `ApplyStats` recalculation after stat allocation and item equip.

### 7) Verification Checklist
- [x] `go test ./...` passes.
- [ ] Manual smoke check in a run:
- [ ] each class starts with distinct baseline stats.
- [ ] killing enemies/boss grants XP and levels correctly.
- [ ] level-up menu allows point spending and closes when points reach zero.
- [ ] move speed / attack cadence change when AGI changes.
- [ ] hp/mana and damage outputs update when VIT/INT/STR/DEX are changed.

## Acceptance Criteria
- [x] Six-stat system is fully class-aware (starting baselines + growth bias).
- [x] Derived stats are computed from a single authoritative path and reused by player runtime logic.
- [x] XP, level-up, and stat allocation are stable during runs and covered by tests.
- [ ] Step 4 scope is complete without coupling to later backlog items beyond required foundations.
