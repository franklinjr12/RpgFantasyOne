# MVP Backlog — Single-player 2D Isometric ARPG (RO-inspired Dungeon Runs)

> Goal: A playable, enjoyable MVP that validates the core loop (run → rooms → boss → deterministic rewards → builds) with enough variety to feel “real”, not a tech demo.

---

## 0) Project Setup & Foundations
- [x] Initialize Go module + raylib-go, create window, fixed timestep (dt) + uncapped render
- [x] Basic asset pipeline (load/unload textures, fonts, audio; fallback placeholders)
- [x] Core game state machine (for mvp): Boot → Main Menu → Class Select → Run → Reward → Results → Main Menu
- [x] Simple debug overlay toggle (FPS, room index, entity counts)

---

## 1) Core Architecture (Systems + Data)
- [x] Implement base `Entity` model (pos/vel, HP, stats, hitbox, effects list, faction, alive)
- [x] Implement system scaffolding + update order:
  - Input → AI → Casting → Projectiles → Movement → Combat Resolve → Effects → Dungeon/Run → UI/Render
- [x] Central config/data layer for classes, skills, enemies, items, effects (code tables ok for MVP)
- [x] Save/load minimal settings (volume, fullscreen, keybind only display not configurable yet)

---

## 2) Isometric World & Camera
- [x] Isometric coordinate handling (grid/world to screen projection) + consistent depth sorting
- [x] Camera that follows player with small smoothing
- [x] Room bounds + collision with walls/obstacles (simple AABB)
- [x] Basic floor/room visuals (tile/flat textures) + door transitions (use placeholders for the textures as flat colorored rectangles for now)

---

## 3) Player Controller & Feel
- [x] Player movement (right mouse click) with speed, acceleration (light), and collision
- [x] Facing direction + sprite flip (all sprites are facing right, so need to flip if going left)
- [x] Basic attack (melee swing or ranged shot depending on class) with hit timing
- [x] Hurt/iframes tuning (short, readable) + hit feedback (flash/knockback)

---

## 4) Stats, Leveling, and Build Core
- [x] Six-stat system implemented (STR/AGI/VIT/INT/DEX/LUK or equivalents)
- [x] Derived stats computation (damage, attack speed, move speed, resist, crit, etc.)
- [x] XP + level ups during a run; allocate stat points via UI
- [x] Class identity baselines (starting stats, growth bias)

---

## 5) Skill System (Casting + Targeting + Delivery)
- [x] Skill definitions: cooldown, cost, targeting spec, delivery spec, damage spec, effect specs
- [x] Cooldown system + resource system (HP sustain rules, mana/energy where applicable)
- [x] Targeting system:
  - Self target
  - Single target (cursor hover/click or nearest-in-range)
  - Area target (ground circle)
  - Directional target (simple forward cone/line for MVP)
- [x] Delivery system:
  - Instant
  - Projectile
  - Delayed AoE (telegraphed ground effect)
- [x] Projectile system (lifetime, collision, hit callback, pierce optional)
- [x] Skill bar input (keys 1–4 initially), cast validation (silence/stun, cooldowns, resource)
- [x] Global cooldown (optional) and cast lockouts (optional) for readability

---

## 6) Status Effects & Combat Resolution
- [x] Effect instances with duration + optional tick rate (burn/poison)
- [x] Non-stacking MVP rule (refresh duration) + future-proof hooks for stacking later
- [x] Effect query helpers used by systems (move speed modifiers, can act, can cast)
- [x] Damage system:
  - Stat scaling (DamageSpec)
  - Mitigation (physical/magic resist)
  - Crit + hit feedback
- [x] Combat resolution system:
  - Apply damage + apply effects + on-hit hooks (lifesteal/mana drain)
- [x] Implement core effects for MVP: Slow, Stun, Freeze (as full slow), Silence, Burn

---

## 7) Classes & Skills (Playable Set)
> Each class must feel distinct and self-sufficient; 4 skills each implemented + tuned.

- [x] Class: **Melee**
  - [x] Power Strike (burst)
  - [x] Guard Stance (damage reduction tradeoff)
  - [x] Blood Oath (lifesteal window)
  - [x] Shockwave Slam (AoE + slow)
- [x] Class: **Ranged**
  - [x] Quick Shot (DPS burst)
  - [x] Retreat Roll (mobility)
  - [x] Focused Aim (damage up, move down)
  - [x] Poison Tip (DoT / %HP tick tuned for bosses)
- [x] Class: **Caster**
  - [x] Arcane Bolt (nuke)
  - [x] Mana Shield (mana absorbs damage)
  - [x] Frost Field (AoE slow zone)
  - [x] Arcane Drain (AoE damage + mana sustain)
- [x] Skill VFX placeholders (simple particles/circles) + SFX hooks
- [x] Skill tuning pass for readability and “one more run” feel

---

## 8) Enemies & AI Variety (Not a Tech Demo)
> MVP should have a small but meaningful roster per biome + elites.

- [x] Enemy framework: stats, move/attack, aggro radius, state machine (idle/chase/attack/retreat optional)
- [x] Implement at least **6 enemies** for the MVP biome:
  - [x] 2 melee chasers (different speeds/HP)
  - [x] 1 ranged attacker (projectiles)
  - [x] 1 “caster” enemy (AoE or debuff)
  - [x] 1 tanky bruiser (slow but heavy)
  - [x] 1 swarmer (low HP, fast)
- [x] Elite modifier system (simple): +HP/+damage + one extra effect (e.g., burn/slow aura)
- [x] Spawn director per room (controls composition, difficulty ramp)

---

## 9) Dungeon Generation & Room Flow (Replayable Runs)
> Fixed-length, room-based, procedural, themed biome.

- [x] Biome 1: define visuals + enemy pool + reward pool theme
- [x] Procedural room generator (rect rooms + doors):
  - [x] 8–12 room run length (configurable)
  - [x] 1–2 mini-event rooms (e.g., shrine/heal or challenge) for variety
  - [x] Boss room at end
- [x] Room completion rules (kill all / kill elites / survive timer in one room type)
- [x] Door transition UX (fade, lock until clear, minimap update)

---

## 10) Boss Encounter (Real Validation Fight)
- [x] Boss entity + arena layout
- [x] Boss AI with at least:
  - [x] Telegraphed heavy attack
  - [x] Area denial mechanic (AoE zones)
  - [x] Add spawn phase OR enrage phase (at 50% HP)
- [x] Boss-specific tuning for solo play (fair, readable, class-agnostic)
- [x] Boss reward trigger (curated selection)

---

## 11) Deterministic Rewards & Items (Curated Progression)
> No low-chance drops; reward choices feel good and enable builds.

- [x] Equipment system with 4 slots: Weapon, Head, Chest, Lower
- [x] Item model: stat modifiers + occasional special effect (proc/passive)
- [x] Curated reward pool for Biome 1:
  - [x] At least **30 items** total (10 per class flavor, overlap allowed)
  - [x] Mix of minor upgrades and build-enablers (e.g., “burn on hit”, “+crit on slowed targets”)
- [x] Reward presentation:
  - [x] Choose 1 of 3 after boss
  - [x] Optional smaller reward after mid-run milestone (e.g., room 4)
- [x] Ensure deterministic feel: weighted selection + anti-repeat (avoid showing identical choice sets)
- [x] Item compare UI (show equipped vs offered deltas)

---

## 12) UI/UX (Must Be Playable and Clear)
### Menus
- [x] Main menu (Start Run, Settings, Quit)
- [x] Class select screen (summary stats + skill previews)
- [x] Pause menu (Resume, Restart Run, Settings, Exit to Menu)
- [x] Results screen (time, rooms cleared, boss killed, build summary)
- [x] Icons
- [x] Skill effects

### In-Run HUD
- [x] Player HP bar + resource (mana/energy) bar
- [x] Skill bar (4 slots):
  - [x] Icons
  - [x] Cooldown overlays
  - [x] Cost indicators
  - [x] Keybind labels
- [x] XP/level display + “stat points available” indicator
- [x] Minimap (simple: room nodes + current room + boss room)
- [x] Buff/Debuff tray (icons + remaining duration)
- [x] Target frame when hovering/locked:
  - [x] Enemy HP bar
  - [x] Name/type + elite indicator
  - [x] Active debuffs icons

### Combat Feedback
- [ ] Floating damage numbers (crit styling)
- [ ] Floating heal numbers
- [ ] Status text popups (e.g., “Stunned”, “Silenced”)
- [ ] Telegraph indicators (AoE circles, line attacks)
- [ ] Hit flashes

### Reward UI
- [x] Reward selection screen (3 cards):
  - [x] Item icon, stats, special effect text
  - [x] Compare with current gear
- [x] “Build recap” panel (key stats, active effects, chosen rewards)

---

## 13) Audio & Juice (Enough to Feel Good)
- [ ] Basic SFX set:
  - [ ] Player hit / enemy hit
  - [ ] Skill cast / projectile impact
  - [ ] UI click/confirm
  - [ ] Boss warning cue
- [ ] Music: menu + biome track + boss track (placeholder ok)
- [ ] Basic particle/VFX for skills and impacts (simple circles/lines ok)

---

## 14) Balance Passes (Make It Enjoyable)
- [ ] Difficulty curve across rooms (enemy HP/damage ramp + spawn density ramp)
- [ ] Sustain tuning for each class (no infinite sustain; no potion dependency)
- [ ] Reward pacing (meaningful upgrades within 1–2 choices)
- [ ] Time-to-kill targets:
  - [ ] Trash pack clear time
  - [ ] Elite time
  - [ ] Boss duration target
- [ ] Ensure 3 distinct viable build archetypes per class using items + stats

---

## 15) Quality & Stability
- [ ] Collision edge cases (doors, corners, knockback)
- [ ] Projectile edge cases (tunneling, lifetime cleanup)
- [ ] Effect edge cases (refresh, removal, immunity hooks optional)
- [ ] No memory leaks on asset reload / run reset
- [ ] Keyboard + mouse rebinding optional (MVP: fixed keys acceptable)
- [ ] “Restart Run” works from pause and after death without crashing

---

## 16) MVP Validation Checklist (Exit Criteria)
- [ ] Complete 10 full runs without crashes
- [ ] Each class can win a run with at least 2 distinct build paths
- [ ] Boss fight feels readable and fair (telegraphs, no cheap hits)
- [ ] Rewards feel satisfying (no long “dead” stretches)
- [ ] Combat feedback is clear (damage numbers, bars, debuffs, skill cooldown UI)
- [ ] Player reports “one more run” feeling in at least 50% of test sessions

---

## Suggested Implementation Order (Practical)
1. Foundations + Game/Run state machine
2. Movement + combat basics + 2 enemies
3. Full skill pipeline (targeting/delivery/projectiles) + status effects
4. 3 classes + 4 skills each
5. Dungeon rooms + spawn director + minimap
6. Enemy roster expansion + elites
7. Boss fight
8. Deterministic rewards + items + reward UI
9. Balance + juice + stability