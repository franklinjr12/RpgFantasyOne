package gamedata

type TargetType int

const (
	TargetSelf TargetType = iota
	TargetEnemy
	TargetArea
	TargetDirection
)

type TargetingSpec struct {
	Type                  TargetType
	Range                 float32
	Radius                float32
	MaxTargets            int
	DirectionalArcDegrees float32
	DirectionalLineWidth  float32
}

type DeliveryType int

const (
	DeliveryInstant DeliveryType = iota
	DeliveryProjectile
	DeliveryDelayed
)

type DeliverySpec struct {
	Type         DeliveryType
	Speed        float32
	Delay        float32
	Lifetime     float32
	Pierce       int
	ProjectileRadius float32
	ZoneDuration float32
	ZoneTickRate float32
}

type DamageType int

const (
	DamagePhysical DamageType = iota
	DamageMagical
	DamageTrue
)

type DamageSpec struct {
	Base       float32
	Scaling    map[StatType]float32
	DamageType DamageType
	CritChance float32
	CritMult   float32
}

type EffectSpec struct {
	Type                EffectType
	Duration            float32
	Magnitude           float32
	TickRate            float32
	PercentMaxHPPerTick float32
	MinTickDamage       int
	MaxTickDamage       int
}

type SelfMovementMode int

const (
	SelfMovementNone SelfMovementMode = iota
	SelfMovementBackwardFromCursor
)

type SelfMovementSpec struct {
	Mode     SelfMovementMode
	Distance float32
}

type ManaShieldSpec struct {
	AbsorbFromCurrentManaRatio float32
	Duration                   float32
}

type ResourceGainSpec struct {
	ManaPerTarget int
}
