# 🧠 Core Design Principle (Re-stated)

> **Entities store state.
> Systems interpret and modify state.
> Skills and effects describe intent, not behavior.**

Every system below exists to enforce that rule.

---

# 1. Game / Run System

**Responsibility**

* Owns the lifetime of a run
* Coordinates update order
* Knows *when* systems should run

**What it does**

* Creates player, dungeon, enemies
* Calls Update() on all systems
* Handles win / loss conditions
* Resets state between runs

**What it does NOT do**

* Combat logic
* Rendering details
* Skill logic

---

# 2. Input System

**Responsibility**

* Convert raw input into player intent

**What it does**

* Reads keyboard/mouse
* Produces:

  * Movement vector
  * Skill cast requests
  * Targeting intent (cursor position, direction)

**What it does NOT do**

* Move player
* Cast skills
* Apply effects

---

# 3. Movement System

**Responsibility**

* Update entity positions based on intent and constraints

**What it does**

* Applies velocity to position
* Reads status effects (slow, freeze, stun)
* Enforces movement bounds
* Handles knockback

**What it does NOT do**

* Pathfinding
* AI decisions
* Animation logic

---

# 4. AI System

**Responsibility**

* Decide *what enemies want to do*

**What it does**

* Aggro detection
* State transitions (idle → chase → attack)
* Chooses when to attack or reposition

**What it does NOT do**

* Move enemies directly
* Apply damage
* Cast skills directly

(AI outputs intent, just like player input.)

---

# 5. Skill Casting System

**Responsibility**

* Orchestrate the skill execution pipeline

**What it does**

* Validates:

  * Cooldowns
  * Resources
  * Silence / stun
* Resolves targeting
* Chooses delivery method (instant / projectile / delayed)
* Starts skill execution

**What it does NOT do**

* Calculate damage
* Apply effects
* Implement skill-specific behavior

---

# 6. Targeting System

**Responsibility**

* Resolve *who* a skill can affect

**What it does**

* Spatial queries
* Distance checks
* Area-of-effect resolution
* Target filters (self, enemy, max targets)

**What it does NOT do**

* Apply damage
* Spawn projectiles
* Decide skill effects

---

# 7. Projectile System

**Responsibility**

* Manage all moving skill deliveries

**What it does**

* Updates projectile movement
* Detects collision
* Triggers skill application on hit
* Handles projectile lifetime

**What it does NOT do**

* Skill logic
* Damage calculation
* Effect logic

---

# 8. Damage System

**Responsibility**

* Calculate and apply damage numbers

**What it does**

* Interprets DamageSpec
* Applies stat scaling
* Handles mitigation (armor/resists)
* Applies critical hits
* Reduces HP

**What it does NOT do**

* Decide when damage happens
* Handle effects
* Know skill names

---

# 9. Status Effect System

**Responsibility**

* Manage all temporary effects on entities

**What it does**

* Updates effect timers
* Applies periodic ticks (burn, poison)
* Removes expired effects
* Exposes effect queries to other systems

**What it does NOT do**

* Move entities
* Calculate base damage
* Contain skill logic

---

# 10. Combat Resolution System

**Responsibility**

* Apply skill results to targets

**What it does**

* Applies damage via Damage System
* Applies effects via Status System
* Handles on-hit hooks (lifesteal, mana drain)

**What it does NOT do**

* Choose targets
* Move entities
* Execute skill casting

(Think of this as the “final common path.”)

---

# 11. Stats System

**Responsibility**

* Own all stat-derived values

**What it does**

* Holds base stats
* Applies equipment modifiers
* Exposes computed values (damage, speed, resist)

**What it does NOT do**

* Apply damage
* Decide skill outcomes

---

# 12. Equipment / Item System

**Responsibility**

* Modify stats and abilities via gear

**What it does**

* Equip / unequip items
* Apply stat modifiers
* Add passive effects

**What it does NOT do**

* Drop logic
* Skill casting
* Damage calculation

---

# 13. Dungeon / Room System

**Responsibility**

* Control dungeon flow and structure

**What it does**

* Generate rooms
* Spawn enemies per room
* Track room completion
* Trigger boss room

**What it does NOT do**

* Combat logic
* Enemy AI
* Rendering

---

# 14. Reward System

**Responsibility**

* Provide deterministic progression

**What it does**

* Present reward choices
* Apply selected item / stat upgrade
* Control reward pool per run

**What it does NOT do**

* Balance combat
* Generate enemies
* Modify dungeon layout

---

# 15. Rendering System

**Responsibility**

* Visualize current game state

**What it does**

* Draw entities
* Draw projectiles
* Draw UI (HP, skills, cooldowns)
* Visual feedback for effects

**What it does NOT do**

* Game logic
* State changes
* Input handling

---

# 16. Audio System (Optional MVP+)

**Responsibility**

* Play feedback sounds

**What it does**

* Attack sounds
* Hit feedback
* Skill cues

**What it does NOT do**

* Drive gameplay logic

---

# 17. Configuration / Data System

**Responsibility**

* Centralize tuning & definitions

**What it does**

* Defines:

  * Skills
  * Effects
  * Items
  * Classes
* Allows balance changes without logic edits

**What it does NOT do**

* Execute gameplay
* Store runtime state
