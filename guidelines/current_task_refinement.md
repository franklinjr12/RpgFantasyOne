# Current Task Refinement - Backlog Item 13 Audio & Juice (Enough to Feel Good)

## Refined Scope
- Source backlog slice: `13) Audio & Juice (Enough to Feel Good)`
- Required outcomes:
- [x] Basic SFX set wired into runtime
- [x] Player hit / enemy hit feedback sounds
- [x] Skill cast / projectile impact sounds
- [x] Player healing sound with anti-annoyance rules
- [x] Player level up sound
- [x] Door open sound

## Asset Grounding (Use Existing Files Only)
- [x] Use `resources/sounds/player_damage_taken.wav` for player damage taken.
- [x] Use `resources/sounds/enemy_damage_taken.wav` for enemy/boss damage taken.
- [x] Use `resources/sounds/player_cast.wav` for player skill cast.
- [x] Use `resources/sounds/enemy_cast.wav` for enemy projectile cast windup/fire.
- [x] Use `resources/sounds/player_healing.mp3` for player healing events that match healing policy.
- [x] Use `resources/sounds/level_up.mp3` for player level-up.
- [x] Use `resources/sounds/door_open.mp3` for room door unlock/open moment.

## Current Code Reality (Grounding)
- [x] `updateBoot` now loads step-13 SFX from `resources/sounds/...`.
- [x] Player cast SFX is routed through `app/game/skill_feedback.go` using the new sound keys.
- [x] Explicit SFX hooks now exist for:
- [x] player hit
- [x] enemy hit
- [x] healing policy differentiation
- [x] level up
- [x] door open on unlock transition
- [x] Door-open playback uses unlock edge detection in `dungeonRunSystem.Update` to avoid per-frame replay.

## Constraints (Do Not Drift)
- [x] Preserve existing combat outcomes, XP flow, room progression, and state transitions.
- [x] Keep pipeline order unchanged.
- [x] Keep fallback-safe behavior when audio device is unavailable (`AssetManager` no-op behavior remains valid).
- [x] Do not introduce skill-name branching in core resolver logic beyond localized SFX policy helpers.
- [x] Follow healing sound rule from product note:
- [x] Do not stack passive per-hit lifesteal heal SFX with hit SFX.
- [x] Only play healing SFX for explicit active-skill healing moments (not passive on-hit class lifesteal ticks).

## Task Backlog

### 1) Create Audio Event Keys and Policy Helpers
- [x] Add a centralized SFX key list in `app/game` (for example: player hit, enemy hit, cast player, cast enemy, heal, level up, door open).
- [x] Add small helper functions that decide which SFX to play for:
- [x] damage target type (player vs enemy/boss)
- [x] healing source type (active-skill heal vs passive lifesteal/kill-heal)
- [x] Ensure helpers are deterministic and testable without requiring real audio playback.

### 2) Migrate Boot Sound Loading to `resources/sounds`
- [x] Replace current missing `resources/audio/*.wav` skill sound loads with the `resources/sounds/*` files listed above.
- [x] Keep non-step-13 audio (menu music/confirm) untouched unless it blocks compilation or startup.
- [x] Ensure all new SFX keys are loaded in `updateBoot`.

### 3) Wire Player/Enemy Hit and Projectile Impact SFX Through Combat Resolution Paths
- [x] Trigger player-hit SFX when `ApplyPlayerCombatHit` successfully applies damage (`AppliedDamage > 0`).
- [x] Trigger enemy-hit SFX when player attacks/skills/projectiles apply damage to enemy or boss.
- [x] Ensure projectile impacts are naturally covered by the same hit-based hooks (no duplicate impact+hurt layering on the same hit event).
- [x] Add a lightweight anti-spam guard (small per-key cooldown) if multi-target hits produce excessive simultaneous overlaps.

### 4) Keep Skill Cast Audio Readable
- [x] Keep cast SFX on successful player skill cast in `TryCastSkill`.
- [x] Use `player_cast.wav` for player casts.
- [x] Play `enemy_cast.wav` when enemy projectile attacks are spawned in combat resolution.

### 5) Implement Healing SFX Policy (Warrior Lifesteal Noise Control)
- [x] Classify healing feedback calls by source (at minimum: passive on-hit lifesteal, kill-heal, active-skill heal).
- [x] Do not play healing SFX for passive lifesteal ticks caused by normal attacks/on-hit hooks.
- [x] Do not play healing SFX for passive kill-heal unless design explicitly reclassifies it as active.
- [x] Play healing SFX only for explicit active-skill healing moments.
- [x] Keep floating heal numbers intact regardless of SFX policy.

### 6) Add Level-Up SFX Hook
- [x] Add a game-layer XP grant helper (or equivalent) that can detect level-up transitions safely.
- [x] Route all runtime XP grants through that helper.
- [x] Play `level_up.mp3` once per level-up event (handle multi-level gains by playing per level gained, capped if needed for sanity).

### 7) Add Door Open SFX Hook
- [x] Detect room door transition from locked -> unlocked in `dungeonRunSystem.Update`.
- [x] Play `door_open.mp3` once when doors become available after room clear.
- [x] Prevent replay every frame while doors remain unlocked.
- [x] Ensure no door-open sound plays for boss reward transition (where doors are not part of progression).

### 8) Tests and Verification
- [x] Add/extend tests in `app/game` for:
- [x] hit SFX routing (player target vs enemy target)
- [x] healing SFX suppression for passive lifesteal
- [x] level-up SFX trigger on XP threshold crossing
- [x] door-open SFX single-trigger on unlock edge
- [x] Use a test seam for sound playback capture (function injection/wrapper), instead of requiring actual audio hardware.
- [ ] Run targeted tests:
- [ ] `go test ./app/game -tags raylib`
- [x] `go test ./app/gamedata ./app/gameobjects ./app/world`
- Note: `go test ./app/game -tags raylib` is blocked in this environment because `raylib.dll` is not available at runtime.
- [ ] Run one manual smoke pass per class and verify audible outcomes for all backlog bullets.

## Definition of Done
- [ ] All step-13 bullets are behaviorally implemented and audible using `resources/sounds`.
- [x] Warrior/passive lifesteal no longer creates hit+heal sound spam on each hit.
- [x] Door open, level up, and cast/impact sounds trigger at correct gameplay moments without per-frame spam.
- [ ] Existing gameplay behavior remains unchanged beyond audio feedback.
- [ ] `guidelines/backlog.md` item `13) Audio & Juice` can be checked off.
