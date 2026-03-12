# Step 11 Refinement - Deterministic Rewards and Items (Curated Progression)

## Goal
- Implement backlog step 11 from `guidelines/backlog.md` with deterministic, curated rewards and item progression.
- Keep architecture aligned with project principles: entities store state, systems apply logic, and item behavior stays data-driven.

## Scope Boundaries
- In scope: equipment/reward data model, curated Biome 1 item pool, deterministic weighted selection with anti-repeat, boss and milestone reward flow, compare UI in reward screen, and focused tests.
- Out of scope: large UI redesign outside reward screen, new class skills, major dungeon generation changes, or full persistence systems.

## Current State Snapshot (from code)
- `ItemSlot` already has 4 slots (`Weapon`, `Head`, `Chest`, `Lower`) in `app/gamedata/items.go`.
- `Item` currently supports only `StatBonuses` + `ClassRestriction`; no special item proc/passive model exists.
- Reward generation uses randomness from `GenerateRewardOptions` and is not a weighted deterministic selector with run history.
- Reward confirm currently always calls `EnterResults(true, rewardPicked)` in `updateReward`, so there is no mid-run reward-return flow.
- Reward UI shows name/description/stat bonuses but no equipped-vs-offered delta compare.
- Current item definitions are below the target of 30 curated items.

## Assumptions for This Refinement
- Boss reward offers remain `choose 1 of 3`.
- Mid-run milestone reward is implemented as a smaller `choose 1 of 2` at room 4 (1-based index), once per run.
- Anti-repeat means avoiding identical offer sets across reward presentations in the same run when alternatives exist.
- Determinism target is run-deterministic with explicit seed/history inputs, not global random behavior.

## Implementation Backlog
- [x] 1) Refactor item data model for curated progression in `app/gamedata`.
  Add stable item identity and reward metadata to `Item` (example: `ID`, `Weight`, optional tier/source tags, optional biome tag).
  Add data-driven special effect model on items (example: typed proc/passive definitions) without name-based branching.
  Keep existing stat bonus behavior compatible with current `ComputeEffectiveStats`.

- [x] 2) Build a curated Biome 1 item pool with at least 30 items.
  Create/organize item tables under `app/gamedata` so total unique items for Biome 1 is `>= 30`.
  Ensure class flavor coverage target is met (`>= 10` flavor items per class, overlap allowed).
  Ensure mix includes minor stat upgrades and build-enabler effects (for example burn-on-hit or bonus crit vs slowed targets).

- [x] 3) Replace ad-hoc reward generation with deterministic weighted selection.
  Introduce a pure selector API in `app/gamedata` (or equivalent core data layer) that accepts explicit inputs:
  class type, biome, reward context, offer size, seed, and previous offer history.
  Implement weighted sampling without duplicates inside one offer.
  Implement anti-repeat for offer sets with bounded rerolls/fallback logic when pool size is small.
  Remove direct `math/rand` global usage from reward generation path.

- [x] 4) Update game reward state to support multiple reward contexts.
  Extend `Game` reward state in `app/game/game.go` to track reward context (boss vs milestone), run reward history, and per-run milestone trigger flag.
  Keep boss reward one-time gating behavior intact.
  Reset all reward-related run state correctly in `ResetState` and run start paths.

- [x] 5) Implement milestone reward flow in run progression.
  Trigger optional milestone reward once at room 4 clear (1-based), before continuing normal room progression.
  On milestone reward confirm, equip selected item and return to `StateRun` (not results).
  On boss reward confirm, keep current expected behavior of ending run into `StateResults`.

- [x] 6) Integrate item special effects into combat resolution.
  Hook item passive/proc application into centralized combat paths (`app/systems/combat_resolver.go` and/or `app/systems/damage_resolver.go`) using item effect types, not item names.
  Ensure effects that depend on target state (example: bonus crit vs slowed target) are resolved in a deterministic, testable way.
  Preserve existing skill/effect and damage pipelines.

- [x] 7) Implement compare UI in reward presentation.
  Update reward screen drawing (`app/game/game.go` and shared UI helpers if needed) to show:
  offered item slot/effects and equipped item in same slot.
  Show stat deltas for equip swap in a stable stat order with positive/negative readability.
  Keep existing keyboard controls (`1/2/3` + `Enter`) and selection highlight behavior.

- [x] 8) Add and update focused automated tests.
  `app/gamedata` tests:
  validate curated pool counts, deterministic selection for same inputs, weighted selection constraints, unique-offer constraint, and anti-repeat behavior.
  `app/game` tests:
  validate reward context transitions (milestone returns to run, boss goes to results), one-time milestone trigger, and reset behavior.
  `app/systems` or relevant package tests:
  validate item special effects are applied correctly through combat resolver paths.

- [x] 9) Keep data access API coherent after refactor.
  Update `app/gamedata/data_access.go` reward accessors so they expose pool and selector data cleanly (instead of returning already-randomized options).
  Update call sites in `app/game/game.go` accordingly.

- [x] 10) Final validation and checklist pass.
  Confirm no skill-name or item-name branching was introduced in core execution paths.
  Confirm step 11 backlog expectations are satisfied end-to-end in code and UI behavior.
  Mark completed boxes in this file as implementation progresses.

## Acceptance Criteria (Step 11 Done Definition)
- [x] Equipment flow supports all 4 slots (`Weapon`, `Head`, `Chest`, `Lower`) through reward equip flow.
- [x] Item model supports both stat modifiers and occasional typed special effects.
- [x] Biome 1 curated reward pool has at least 30 items with class flavor coverage and build-enabler presence.
- [x] Boss reward presents 3 deterministic weighted options, anti-repeat protected, choose 1.
- [x] Mid-run milestone reward (room 4) presents a smaller deterministic offer and returns to run after selection.
- [x] Reward UI includes equipped-vs-offered compare deltas.
- [x] New/updated tests cover deterministic selection, anti-repeat, reward flow transitions, and special effect application.

## Suggested Verification Commands
- `go test ./app/gamedata`
- `go test ./app/gameobjects`
- `go test ./app/world`
- `go test ./app/game -tags raylib` (if raylib runtime is available locally)
- `go test ./app/systems -tags raylib` (if raylib runtime is available locally)

## Notes for Implementing Agent
- Keep changes incremental and avoid broad rewrites.
- Prefer adding pure, testable selection/helpers in `app/gamedata` and thin orchestration in `app/game`.
- Preserve existing gameplay behavior outside reward/item scope.
