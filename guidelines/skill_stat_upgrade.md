## 1. First Principle (Very Important)

> **Skills do not calculate damage.
> Skills describe how damage should be calculated.**

This single rule keeps your system sane.

---

## 2. Separate “Damage Description” from “Damage Resolution”

Think in two layers:

```
Skill → DamageSpec → DamageSystem → Entity
```

* **Skill**: declares *what kind* of damage
* **DamageSpec**: data describing scaling
* **DamageSystem**: does math using stats
* **Entity**: receives final number

---

## 3. DamageSpec (The Key Abstraction)

This is a **pure data struct**.

```go
type DamageType int

const (
    DamagePhysical DamageType = iota
    DamageMagical
    DamageTrue
)
```

```go
type DamageSpec struct {
    Base        float32
    Scaling     map[StatType]float32
    DamageType  DamageType
    CritChance  float32
    CritMult    float32
}
```

Example interpretation:

* Base = flat damage
* Scaling = how much each stat contributes
* DamageType = armor / resist logic
* Crit is optional

---

## 4. Stat Scaling Without Conditionals

Your stats are likely something like:

```go
type StatType int

const (
    STR StatType = iota
    AGI
    INT
    DEX
    VIT
    LUK
)
```

```go
type Stats struct {
    Values map[StatType]float32
}
```

Now scaling is trivial:

```go
func ComputeDamage(spec DamageSpec, stats Stats) float32 {
    dmg := spec.Base

    for stat, factor := range spec.Scaling {
        dmg += stats.Values[stat] * factor
    }

    return dmg
}
```

No switch statements.
No class-specific logic.

---

## 5. Applying Damage to a Target

This happens **after** scaling.

```go
func ApplyDamage(
    attacker *Entity,
    target *Entity,
    spec DamageSpec,
) {
    raw := ComputeDamage(spec, attacker.Stats)

    final := ResolveMitigation(raw, spec.DamageType, target)

    target.HP -= final
}
```

---

## 6. Defense & Mitigation Stay Centralized

```go
func ResolveMitigation(dmg float32, dmgType DamageType, target *Entity) float32 {
    switch dmgType {
    case DamagePhysical:
        return dmg * (1.0 - target.Stats.PhysicalResist)
    case DamageMagical:
        return dmg * (1.0 - target.Stats.MagicalResist)
    case DamageTrue:
        return dmg
    }
    return dmg
}
```

Again:

* Skills don’t care
* Effects don’t care
* Enemies don’t care

---

## 7. Skills Become Declarative (And Beautiful)

### Example: Melee Power Strike

```go
var PowerStrikeDamage = DamageSpec{
    Base: 20,
    Scaling: map[StatType]float32{
        STR: 1.2,
    },
    DamageType: DamagePhysical,
    CritChance: 0.1,
    CritMult:   1.5,
}
```

```go
func CastPowerStrike(caster *Entity, target *Entity) {
    ApplyDamage(caster, target, PowerStrikeDamage)
}
```

That’s it.

---

### Example: Caster Arcane Bolt

```go
var ArcaneBoltDamage = DamageSpec{
    Base: 15,
    Scaling: map[StatType]float32{
        INT: 1.5,
    },
    DamageType: DamageMagical,
}
```

---

### Example: Hybrid Scaling (RO-style)

```go
var PoisonStrike = DamageSpec{
    Base: 10,
    Scaling: map[StatType]float32{
        STR: 0.6,
        DEX: 0.4,
    },
    DamageType: DamagePhysical,
}
```

Exactly how Ragnarok does it internally.

---

## 8. Damage Over Time That Scales

Do **not** recalc damage every tick.

Calculate once, store magnitude.

```go
func ApplyBurn(attacker *Entity, target *Entity) {
    dmg := ComputeDamage(BurnDamageSpec, attacker.Stats)

    ApplyEffect(target, Effect{
        Type:      EffectBurn,
        Duration:  5,
        Magnitude: dmg,
        TickRate:  1,
    })
}
```

Each tick simply applies `Magnitude`.

This avoids stat snapshot bugs.

---

## 9. Skill + Effect Combo Example

### Frost Explosion

* Immediate damage
* Applies slow

```go
func CastFrostExplosion(caster *Entity, targets []*Entity) {
    for _, t := range targets {
        ApplyDamage(caster, t, FrostExplosionDamage)
        ApplyEffect(t, SlowEffect)
    }
}
```

Zero complexity explosion.

---

## 10. Why This Architecture Survives Growth

Let’s say you add:

* Damage that scales with missing HP
* Damage that scales with number of effects
* Damage that converts damage types

You **extend DamageSpec**, not skills.

Example:

```go
type DamageSpec struct {
    Base       float32
    Scaling    map[StatType]float32
    OnCompute  func(attacker, target *Entity) float32
}
```

Still centralized. Still safe.

---

## 11. Performance & Go Considerations

* `map[StatType]float32` is fine at MVP scale
* Later you can replace with fixed arrays
* Do **not** prematurely optimize

Clarity > speed until proven otherwise.

---

## 12. Golden Rule #2 (Just as Important)

> **Skills declare intent
> DamageSpec declares math
> DamageSystem executes math**
