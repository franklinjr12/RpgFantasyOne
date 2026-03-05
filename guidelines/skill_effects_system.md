The key idea you need to internalize:

> **Skills should trigger effects.
> Effects should be dumb, data-driven, and time-bound.
> Entities should not know *why* they are affected.**

If you respect that separation, complexity stays flat instead of exponential.

---

## 1. The Core Mental Model (Lock This First)

### ❌ Bad model (what causes chaos)

* Skill → custom logic → modifies entity directly
* Freeze skill sets `enemy.CanMove = false`
* Burn skill ticks damage inside enemy update
* Every new effect adds `if stunned`, `if frozen`, `if silenced` everywhere

This leads to:

* Spaghetti conditionals
* Hard-to-balance interactions
* Fear of adding new effects

---

### ✅ Correct model (scales cleanly)

```
Skill → Effect(s) → Status System → Entity
```

* Skills **do not implement behavior**
* Skills only **apply effects**
* Effects are **data + timers**
* Systems interpret effects generically

This is how RO, Diablo, Hades, etc. survive.

---

## 2. Entities Must Be Effect-Agnostic

Your `Player` and `Enemy` structs should **not** have:

* `IsFrozen`
* `IsStunned`
* `IsPoisoned`

Instead, they have **one thing**:

```go
type Entity struct {
    Pos     rl.Vector2
    Stats   Stats
    HP      float32

    Effects []EffectInstance
}
```

That’s it.

Entities don’t care *what* effects do.
They only carry them.

---

## 3. Effects Are Data, Not Logic

### Effect Definition (Static)

```go
type EffectType int

const (
    EffectSlow EffectType = iota
    EffectStun
    EffectFreeze
    EffectBurn
    EffectSilence
)
```

```go
type Effect struct {
    Type       EffectType
    Duration   float32
    Magnitude  float32
    TickRate   float32 // 0 if non-periodic
}
```

This is **pure data**.

---

### Effect Instance (Runtime)

```go
type EffectInstance struct {
    Effect
    TimeLeft float32
    TickTimer float32
}
```

Every applied effect becomes an instance with its own timer.

---

## 4. The Status Effect System (The Heart)

This system runs **once per frame**, for *all entities*.

```go
func UpdateEffects(entity *Entity, dt float32) {
    for i := 0; i < len(entity.Effects); i++ {
        e := &entity.Effects[i]

        e.TimeLeft -= dt

        if e.TickRate > 0 {
            e.TickTimer += dt
            if e.TickTimer >= e.TickRate {
                ApplyEffectTick(entity, e)
                e.TickTimer = 0
            }
        }

        if e.TimeLeft <= 0 {
            RemoveEffect(entity, e)
            i--
        }
    }
}
```

This loop never changes when you add new skills.

---

## 5. Effect Interpretation Happens in Systems

Effects do **nothing on their own**.
Systems *query* effects and modify behavior.

---

### Example: Movement System

```go
func GetMoveSpeed(entity *Entity) float32 {
    speed := entity.Stats.MoveSpeed

    for _, e := range entity.Effects {
        if e.Type == EffectSlow {
            speed *= (1.0 - e.Magnitude)
        }
        if e.Type == EffectFreeze {
            return 0
        }
        if e.Type == EffectStun {
            return 0
        }
    }

    return speed
}
```

No flags. No booleans.
Just **queries**.

---

### Example: Skill Casting System

```go
func CanCast(entity *Entity) bool {
    for _, e := range entity.Effects {
        if e.Type == EffectSilence || e.Type == EffectStun {
            return false
        }
    }
    return true
}
```

---

### Example: Damage Over Time (Burn)

```go
func ApplyEffectTick(entity *Entity, e *EffectInstance) {
    switch e.Type {
    case EffectBurn:
        entity.HP -= e.Magnitude
    }
}
```

Only tick-based effects need logic here.

---

## 6. Skills Become Extremely Simple

Skills now only:

* Check conditions
* Apply effects

### Skill Execution Example

```go
func CastFrostField(caster *Entity, targets []*Entity) {
    for _, t := range targets {
        ApplyEffect(t, Effect{
            Type:      EffectSlow,
            Duration:  4.0,
            Magnitude: 0.4,
        })
    }
}
```

No movement logic.
No timers.
No per-skill hacks.

---

## 7. Stacking Rules (Keep MVP Simple)

For MVP:

**Rule: No stacking. Refresh duration only.**

```go
func ApplyEffect(entity *Entity, newEffect Effect) {
    for i := range entity.Effects {
        if entity.Effects[i].Type == newEffect.Type {
            entity.Effects[i].TimeLeft = newEffect.Duration
            return
        }
    }

    entity.Effects = append(entity.Effects, EffectInstance{
        Effect:   newEffect,
        TimeLeft: newEffect.Duration,
    })
}
```

Later you can add:

* Stack count
* Diminishing returns
* Immunities

Not now.

---

## 8. Freeze vs Stun (Important Distinction)

Avoid special cases.

| Effect  | Interpretation                  |
| ------- | ------------------------------- |
| Slow    | Reduces movement speed          |
| Freeze  | Slow with magnitude = 1         |
| Stun    | Blocks movement **and** actions |
| Silence | Blocks skills only              |

Mechanically:

* Freeze is just a *full slow*
* Stun is a *movement + action block*

No new code paths.

---

## 9. Why This Architecture Scales

Let’s say you add **Burning Ice**:

* Slows
* Deals damage over time

You don’t write a new system.

You just apply two effects.

---

## 10. Debugging Becomes Trivial

You can:

* Print all effects on an entity
* Visualize timers
* Toggle effects live

That’s impossible with flag-based designs.

---

## 11. Golden Rule (Tattoo This)

> **Entities hold state
> Effects describe state
> Systems interpret state**
