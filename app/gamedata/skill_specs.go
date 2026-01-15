package gamedata

type TargetType int

const (
	TargetSelf TargetType = iota
	TargetEnemy
	TargetArea
)

type TargetingSpec struct {
	Type       TargetType
	Range      float32
	Radius     float32
	MaxTargets int
}

type DeliveryType int

const (
	DeliveryInstant DeliveryType = iota
	DeliveryProjectile
	DeliveryDelayed
)

type DeliverySpec struct {
	Type     DeliveryType
	Speed    float32
	Delay    float32
	Lifetime float32
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
	Type      int
	Duration  float32
	Magnitude float32
	TickRate  float32
}

