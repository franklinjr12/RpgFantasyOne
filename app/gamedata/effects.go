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

var effectDefinitions = map[EffectType]Effect{
	EffectSlow:               {Type: EffectSlow, Duration: 2.0, Magnitude: 0.3},
	EffectStun:               {Type: EffectStun, Duration: 1.0, Magnitude: 0},
	EffectFreeze:             {Type: EffectFreeze, Duration: 1.0, Magnitude: 0},
	EffectSilence:            {Type: EffectSilence, Duration: 2.0, Magnitude: 0},
	EffectBurn:               {Type: EffectBurn, Duration: 4.0, Magnitude: 3.0, TickRate: 1.0},
	EffectPoison:             {Type: EffectPoison, Duration: 5.0, Magnitude: 3.0, TickRate: 1.0},
	EffectDamageReduction:    {Type: EffectDamageReduction, Duration: 4.0, Magnitude: 0.4},
	EffectMoveSpeedReduction: {Type: EffectMoveSpeedReduction, Duration: 4.0, Magnitude: 0.3},
	EffectLifesteal:          {Type: EffectLifesteal, Duration: 5.0, Magnitude: 0.3},
	EffectDamageBoost:        {Type: EffectDamageBoost, Duration: 5.0, Magnitude: 0.5},
	EffectMoveSpeedBoost:     {Type: EffectMoveSpeedBoost, Duration: 2.0, Magnitude: 0.5},
}

func GetEffectDefinition(effectType EffectType) (Effect, bool) {
	effect, ok := effectDefinitions[effectType]
	return effect, ok
}
