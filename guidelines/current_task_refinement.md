
# Step 8 Refinement - Enemies & AI Variety (Not a Tech Demo)

## Goal
- Deliver a small but meaningful enemy roster for the MVP biome with distinct behavior, elite variety, and deterministic per-room compositions.
- Keep implementation aligned with project principles: entities store state, systems decide behavior, data lives in `app/gamedata`.

## Scope And Constraints
- Scope is only backlog step 8 from `guidelines/backlog.md`.
- Preserve current run loop behavior: room clear flow, door lock/unlock, XP gain, reward/results transitions.
- Keep Windows + raylib-go assumptions.
- Prefer incremental migration; do not perform a full architecture rewrite.
- Avoid skill-name/string branching for enemy behavior; drive behavior by typed data/profile fields.

## Current Baseline (Observed)
- `app/gamedata/enemies.go` only defines 3 templates (`Normal`, `Elite`, `Boss`) with no archetype variety.
- `app/gameobjects/enemy.go` mixes enemy state machine + movement + attack in `Enemy.Update`.
- `app/world/room.go` spawns enemies with `EnemyRef{X,Y,IsElite}` only; no type/composition director.
- `app/game/runtime_pipeline.go` AI system calls `enemy.Update`; combat system consumes `enemy.Attack()`.
- Elites are currently just stronger stats (`IsElite`), no extra modifier behavior.

## Task Backlog

### 8.1 Enemy Data Foundation (gamedata-first)
- [x] Add typed enemy archetype definitions in `app/gamedata/enemies.go`:
  - `EnemyArchetypeType` enum with at least 6 non-boss archetypes (2 melee chasers, 1 ranged, 1 caster, 1 bruiser, 1 swarmer).
  - `EnemyArchetype` struct containing combat and AI profile fields (HP, damage, move speed, cooldown, attack/aggro range, attack mode, preferred/retreat range, hitbox, XP reward).
  - Keep boss template path compatible with current boss setup.
- [x] Add elite modifier data model in `app/gamedata/enemies.go`:
  - `EliteModifierType` enum.
  - `EliteModifier` struct with baseline stat multipliers (`HP`, `Damage`) plus one extra effect payload (on-hit effect and/or aura spec).
- [x] Extend accessors in `app/gamedata/data_access.go` for enemy archetypes and elite modifiers.
- [x] Add/extend tests in `app/gamedata` validating:
  - Exactly 6+ non-boss archetypes are registered.
  - Required fields are valid (`> 0` where expected).
  - Elite modifiers exist and are retrievable deterministically.

### 8.2 Enemy Spawn Reference + Construction Path
- [x] Extend `app/world/room.go` `EnemyRef` to include enemy type and elite modifier metadata (not only `IsElite`).
- [x] Introduce constructor path in `app/gameobjects/enemy.go` that builds enemy instances from typed spawn refs (archetype + elite modifier), while keeping existing codepaths compiling.
- [x] Ensure spawned enemies carry enough metadata for AI, rendering, HUD label, and combat effect logic.
- [x] Update spawn callsites in `app/game/game.go` (`SpawnRoomEnemies`) to use the new typed enemy refs.

### 8.3 Enemy AI Framework (System-driven decisions)
- [x] Refactor enemy update responsibilities so AI decision logic is owned by the AI system (or AI helper invoked by it), not embedded movement logic inside `Enemy.Update`.
- [x] Add intent/state fields on enemy entities as needed (state, desired movement vector, attack intent, cooldown/timing trackers).
- [x] Keep state machine explicit and testable: `idle -> chase -> attack` with optional `retreat` for ranged/caster profiles.
- [x] Update `app/game/runtime_pipeline.go`:
  - AI step computes intent per archetype/profile.
  - Movement step applies enemy movement using that intent.
  - Combat/projectile steps resolve enemy attacks from intent and cooldown readiness.
- [x] Maintain effect integration (`gamedata.UpdateEffects`, `CanAct`, move-speed modifiers) for enemies after refactor.

### 8.4 Implement MVP Roster Variety (6 Enemy Roles)
- [x] Implement 2 melee chaser archetypes with distinct speed/HP profiles.
- [x] Implement 1 ranged enemy that attacks via projectile delivery.
- [x] Implement 1 caster enemy that uses AoE or debuff behavior (telegraphed if AoE).
- [x] Implement 1 tanky bruiser (slow move, heavy hit, longer cadence).
- [x] Implement 1 swarmer (low HP, high speed, pressure role).
- [x] Ensure role behavior is profile/data-driven (`AttackMode`, ranges, timing, effect specs), not keyed by display name checks.

### 8.5 Elite Modifier System (Beyond Stat Inflation)
- [x] Keep baseline elite stat bump (+HP/+damage) data-driven through elite modifiers.
- [x] Implement at least one extra elite behavior effect (example: burn-on-hit, slow aura, or equivalent) and apply it through centralized combat/effect paths.
- [x] Ensure elite modifiers are visible to player-facing labels (e.g., target frame name or debug text) without breaking existing HUD flow.
- [x] Keep deterministic modifier assignment for a given room seed/index.

### 8.6 Spawn Director Per Room (Composition + Ramp)
- [x] Replace current enemy generation in `app/world/room.go` with a spawn director that decides composition before position placement.
- [x] Spawn director inputs must include room progression context (room index or equivalent) and deterministic RNG source.
- [x] Add difficulty ramp rules by room progression:
  - Threat budget increases over rooms.
  - Composition shifts from mostly basic melee to mixed roles.
  - Elite frequency/modifier pressure ramps later in run.
- [x] Preserve room spawn safety constraints already present (avoid center safe zone and obstacle overlap).
- [x] Keep dungeon generation deterministic; same seed/setup must yield the same room compositions.

### 8.7 Visual/Debug Readability For Variety
- [x] Provide clear visual differentiation by archetype using existing placeholder rendering approach (tint and/or sprite cell selection).
- [x] Preserve elite readability layered over archetype visuals.
- [x] Add debug overlay support for room enemy composition summary (counts by type, elite count/modifier) to aid tuning.

### 8.8 Test Coverage And Verification Gates
- [x] Add/extend deterministic tests in `app/world` for spawn director composition/ramp invariants.
- [x] Add unit tests in `app/gameobjects` (or pure helper package if introduced) for enemy AI state transitions and intent decisions.
- [x] Add tests for elite behavior application through combat/effect pipeline (at least one positive case).
- [x] Ensure these package tests pass locally:
  - `go test ./app/gamedata ./app/world ./app/gameobjects ./app/core`
- [ ] If `raylib.dll` is available, also run:
  - `go test -tags raylib ./app/game`

## Manual Smoke Checklist (Post-Implementation)
- [ ] Run starts and room progression still works (doors lock until clear, transitions unchanged).
- [ ] By mid-run, at least 4 distinct enemy archetypes are observed.
- [ ] Ranged enemy projectiles and caster control behavior are readable and functional.
- [ ] Elite enemies show a modifier-driven behavior beyond raw HP/damage.
- [ ] A full run to boss/reward/results completes without crash.

## Out Of Scope For This Step
- Boss encounter redesign (backlog step 10).
- Reward/item system expansion (backlog step 11).
- Full dungeon/biome expansion beyond required enemy variety work for current biome.
