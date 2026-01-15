package gamedata

type EffectType int

const (
	EffectSlow EffectType = iota
	EffectStun
	EffectFreeze
	EffectSilence
	EffectBurn
	EffectPoison
	EffectDamageReduction
	EffectMoveSpeedReduction
	EffectLifesteal
	EffectDamageBoost
	EffectMoveSpeedBoost
)

type Effect struct {
	Type      EffectType
	Duration  float32
	Magnitude float32
	TickRate  float32
}

type EffectInstance struct {
	Effect
	TimeLeft  float32
	TickTimer float32
}

