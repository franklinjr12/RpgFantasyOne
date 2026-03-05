> ❌ **Do NOT create one class per skill type**
> ❌ **Do NOT put casting logic inside skills**
> ❌ **Do NOT branch on skill names**

Instead, you want **composition + a small execution pipeline**.

---

# 1. The Correct Mental Model

Think of **casting** as a *pipeline*, not a function:

```
Input → Targeting → Validation → Execution → Effects
```

Each step is **generic**, reusable, and data-driven.

Skills don’t *do things*.
They **describe how they should be executed**.

---

# 2. Skill = Data + Hooks (Very Small)

```go
type Skill struct {
    ID          SkillID
    Name        string
    Cooldown    float32
    Cost        ResourceCost

    Targeting   TargetingSpec
    Delivery    DeliverySpec
    Effects     []EffectSpec
    Damage      *DamageSpec
}
```

That’s it.

No behavior. No logic.

---

# 3. TargetingSpec — “Who can this hit?”

This handles:

* Self
* Single enemy
* Area
* Directional cone
* Ground target

```go
type TargetType int

const (
    TargetSelf TargetType = iota
    TargetEnemy
    TargetArea
    TargetDirection
)
```

```go
type TargetingSpec struct {
    Type        TargetType
    Range       float32
    Radius      float32
    MaxTargets  int
}
```

Examples:

* Melee strike → Enemy, short range
* Heal → Self
* Fireball → Area, medium range
* Cone → Direction

---

# 4. DeliverySpec — “How does it reach the target?”

This solves:

* Instant melee
* Projectile
* Delayed AoE
* Ground effect

```go
type DeliveryType int

const (
    DeliveryInstant DeliveryType = iota
    DeliveryProjectile
    DeliveryDelayed
)
```

```go
type DeliverySpec struct {
    Type        DeliveryType
    Speed       float32   // projectile
    Delay       float32   // delayed AoE
    Lifetime    float32
}
```

This is where *projectiles live*, not in skills.

---

# 5. Execution Pipeline (The Heart)

### Step 1 — Input & Intent

```go
func TryCast(caster *Entity, skill *Skill) {
    if !CanCast(caster) || skill.OnCooldown {
        return
    }

    intent := GatherTargetingIntent(skill.Targeting)
    ExecuteSkill(caster, skill, intent)
}
```

---

### Step 2 — Target Resolution

```go
func ResolveTargets(
    caster *Entity,
    intent TargetIntent,
    spec TargetingSpec,
) []*Entity {
    // spatial queries
    // distance checks
    // filters
}
```

This code **never changes** when adding skills.

---

### Step 3 — Delivery Handling

```go
func ExecuteSkill(
    caster *Entity,
    skill *Skill,
    intent TargetIntent,
) {
    switch skill.Delivery.Type {
    case DeliveryInstant:
        ApplySkill(caster, skill, ResolveTargets(...))

    case DeliveryProjectile:
        SpawnProjectile(caster, skill, intent)

    case DeliveryDelayed:
        SpawnDelayedEffect(caster, skill, intent)
    }
}
```

This is your only switch.

---

## 6. Projectiles Are First-Class Entities

A projectile is **not a skill**.

```go
type Projectile struct {
    Pos        rl.Vector2
    Vel        rl.Vector2
    Skill      *Skill
    Caster     *Entity
    Lifetime   float32
}
```

When it hits:

```go
func OnProjectileHit(p *Projectile, target *Entity) {
    ApplySkill(p.Caster, p.Skill, []*Entity{target})
}
```

No duplication.
No custom projectile code per skill.

---

## 7. ApplySkill = Final Common Step

This is where damage and effects happen.

```go
func ApplySkill(
    caster *Entity,
    skill *Skill,
    targets []*Entity,
) {
    for _, t := range targets {
        if skill.Damage != nil {
            ApplyDamage(caster, t, *skill.Damage)
        }

        for _, e := range skill.Effects {
            ApplyEffect(t, e.ToEffect(caster))
        }
    }
}
```

Every skill ends here.

---

## 8. Area-of-Effect Is Just Targeting

AoE skills:

* TargetingSpec.Type = TargetArea
* Radius > 0

Projectile AoE:

* Delivery = Projectile
* On hit → ResolveTargets with radius

No new systems needed.

---

## 9. Self-Target Skills Are Trivial

```go
TargetingSpec{
    Type: TargetSelf,
}
```

Target resolution returns `[caster]`.

---

## 10. Melee Skills Are Just Short-Range Instants

```go
TargetingSpec{
    Type:  TargetEnemy,
    Range: 1.5,
}
Delivery: Instant
```

No special logic.

---

## 11. Why This Doesn’t Explode in Complexity

Let’s add a new skill:

> “Delayed AoE that fires a projectile and silences enemies”

You don’t add new systems.

You define:

* Targeting: Area
* Delivery: Projectile
* Effect: Silence

The engine already supports it.

---

## 12. MVP Simplifications (Strongly Recommended)

For your MVP:

* ❌ No cones
* ❌ No chains
* ❌ No bouncing projectiles
* ❌ No homing

You can add them later by extending **TargetingSpec**, not skills.

---

## 13. Debugging & Visualization (Future-Proof)

Because everything is declarative:

* Draw targeting radius
* Draw projectile paths
* Print resolved targets

This is nearly impossible in ad-hoc designs.

---

## 14. The Golden Rule #3

> **Skills describe
> Casting systems execute
> Effects decide outcome**

If you ever feel tempted to write:

```go
if skill.Name == "Fireball"
```

Stop. Refactor.
