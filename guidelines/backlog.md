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
- [ ] Isometric coordinate handling (grid/world to screen projection) + consistent depth sorting
- [ ] Camera that follows player with small smoothing
- [ ] Room bounds + collision with walls/obstacles (simple AABB)
- [ ] Basic floor/room visuals (tile/flat textures) + door transitions (use placeholders for the textures as flat colorored rectangles for now)

---

## 3) Player Controller & Feel
- [ ] Player movement (right mouse click) with speed, acceleration (light), and collision
- [ ] Facing direction + sprite flip (all sprites are facing right, so need to flip if going left)
- [ ] Basic attack (melee swing or ranged shot depending on class) with hit timing
- [ ] Hurt/iframes tuning (short, readable) + hit feedback (flash/knockback)

---

## 4) Stats, Leveling, and Build Core
- [ ] Six-stat system implemented (STR/AGI/VIT/INT/DEX/LUK or equivalents)
- [ ] Derived stats computation (damage, attack speed, move speed, resist, crit, etc.)
- [ ] XP + level ups during a run; allocate stat points via UI
- [ ] Class identity baselines (starting stats, growth bias)

---

## 5) Skill System (Casting + Targeting + Delivery)
- [ ] Skill definitions: cooldown, cost, targeting spec, delivery spec, damage spec, effect specs
- [ ] Cooldown system + resource system (HP sustain rules, mana/energy where applicable)
- [ ] Targeting system:
  - Self target
  - Single target (cursor hover/click or nearest-in-range)
  - Area target (ground circle)
  - Directional target (simple forward cone/line for MVP)
- [ ] Delivery system:
  - Instant
  - Projectile
  - Delayed AoE (telegraphed ground effect)
- [ ] Projectile system (lifetime, collision, hit callback, pierce optional)
- [ ] Skill bar input (keys 1–4 initially), cast validation (silence/stun, cooldowns, resource)
- [ ] Global cooldown (optional) and cast lockouts (optional) for readability

---

## 6) Status Effects & Combat Resolution
- [ ] Effect instances with duration + optional tick rate (burn/poison)
- [ ] Non-stacking MVP rule (refresh duration) + future-proof hooks for stacking later
- [ ] Effect query helpers used by systems (move speed modifiers, can act, can cast)
- [ ] Damage system:
  - Stat scaling (DamageSpec)
  - Mitigation (physical/magic resist)
  - Crit + hit feedback
- [ ] Combat resolution system:
  - Apply damage + apply effects + on-hit hooks (lifesteal/mana drain)
- [ ] Implement core effects for MVP: Slow, Stun, Freeze (as full slow), Silence, Burn

---

## 7) Classes & Skills (Playable Set)
> Each class must feel distinct and self-sufficient; 4 skills each implemented + tuned.

- [ ] Class: **Melee**
  - [ ] Power Strike (burst)
  - [ ] Guard Stance (damage reduction tradeoff)
  - [ ] Blood Oath (lifesteal window)
  - [ ] Shockwave Slam (AoE + slow)
- [ ] Class: **Ranged**
  - [ ] Quick Shot (DPS burst)
  - [ ] Retreat Roll (mobility)
  - [ ] Focused Aim (damage up, move down)
  - [ ] Poison Tip (DoT / %HP tick tuned for bosses)
- [ ] Class: **Caster**
  - [ ] Arcane Bolt (nuke)
  - [ ] Mana Shield (mana absorbs damage)
  - [ ] Frost Field (AoE slow zone)
  - [ ] Arcane Drain (AoE damage + mana sustain)
- [ ] Skill VFX placeholders (simple particles/circles) + SFX hooks
- [ ] Skill tuning pass for readability and “one more run” feel

---

## 8) Enemies & AI Variety (Not a Tech Demo)
> MVP should have a small but meaningful roster per biome + elites.

- [ ] Enemy framework: stats, move/attack, aggro radius, state machine (idle/chase/attack/retreat optional)
- [ ] Implement at least **6 enemies** for the MVP biome:
  - [ ] 2 melee chasers (different speeds/HP)
  - [ ] 1 ranged attacker (projectiles)
  - [ ] 1 “caster” enemy (AoE or debuff)
  - [ ] 1 tanky bruiser (slow but heavy)
  - [ ] 1 swarmer (low HP, fast)
- [ ] Elite modifier system (simple): +HP/+damage + one extra effect (e.g., burn/slow aura)
- [ ] Spawn director per room (controls composition, difficulty ramp)

---

## 9) Dungeon Generation & Room Flow (Replayable Runs)
> Fixed-length, room-based, procedural, themed biome.

- [ ] Biome 1: define visuals + enemy pool + reward pool theme
- [ ] Procedural room generator (rect rooms + doors):
  - [ ] 8–12 room run length (configurable)
  - [ ] 1–2 mini-event rooms (e.g., shrine/heal or challenge) for variety
  - [ ] Boss room at end
- [ ] Room completion rules (kill all / kill elites / survive timer in one room type)
- [ ] Door transition UX (fade, lock until clear, minimap update)

---

## 10) Boss Encounter (Real Validation Fight)
- [ ] Boss entity + arena layout
- [ ] Boss AI with at least:
  - [ ] Telegraphed heavy attack
  - [ ] Area denial mechanic (AoE zones)
  - [ ] Add spawn phase OR enrage phase (at 50% HP)
- [ ] Boss-specific tuning for solo play (fair, readable, class-agnostic)
- [ ] Boss reward trigger (curated selection)

---

## 11) Deterministic Rewards & Items (Curated Progression)
> No low-chance drops; reward choices feel good and enable builds.

- [ ] Equipment system with 4 slots: Weapon, Head, Chest, Lower
- [ ] Item model: stat modifiers + occasional special effect (proc/passive)
- [ ] Curated reward pool for Biome 1:
  - [ ] At least **30 items** total (10 per class flavor, overlap allowed)
  - [ ] Mix of minor upgrades and build-enablers (e.g., “burn on hit”, “+crit on slowed targets”)
- [ ] Reward presentation:
  - [ ] Choose 1 of 3 after boss
  - [ ] Optional smaller reward after mid-run milestone (e.g., room 4)
- [ ] Ensure deterministic feel: weighted selection + anti-repeat (avoid showing identical choice sets)
- [ ] Item compare UI (show equipped vs offered deltas)

---

## 12) UI/UX (Must Be Playable and Clear)
### Menus
- [ ] Main menu (Start Run, Settings, Quit)
- [ ] Class select screen (summary stats + skill previews)
- [ ] Pause menu (Resume, Restart Run, Settings, Exit to Menu)
- [ ] Results screen (time, rooms cleared, boss killed, build summary)

### In-Run HUD
- [ ] Player HP bar + resource (mana/energy) bar
- [ ] Skill bar (4 slots):
  - [ ] Icons
  - [ ] Cooldown overlays
  - [ ] Cost indicators
  - [ ] Keybind labels
- [ ] XP/level display + “stat points available” indicator
- [ ] Minimap (simple: room nodes + current room + boss room)
- [ ] Buff/Debuff tray (icons + remaining duration)
- [ ] Target frame when hovering/locked:
  - [ ] Enemy HP bar
  - [ ] Name/type + elite indicator
  - [ ] Active debuffs icons

### Combat Feedback
- [ ] Floating damage numbers (crit styling)
- [ ] Floating heal numbers
- [ ] Status text popups (e.g., “Stunned”, “Silenced”)
- [ ] Telegraph indicators (AoE circles, line attacks)
- [ ] Hit flashes + small screenshake toggle in settings

### Reward UI
- [ ] Reward selection screen (3 cards):
  - [ ] Item icon, stats, special effect text
  - [ ] Compare with current gear
- [ ] “Build recap” panel (key stats, active effects, chosen rewards)

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