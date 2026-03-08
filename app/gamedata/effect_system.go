package gamedata

type EffectStackPolicy int

const (
	EffectStackPolicyRefreshDuration EffectStackPolicy = iota
	EffectStackPolicyRefreshDurationKeepStrongestMagnitude
)

var DefaultEffectStackPolicy = EffectStackPolicyRefreshDurationKeepStrongestMagnitude

func UpdateEffects(effects *[]EffectInstance, dt float32, takeDamage func(int)) {
	if effects == nil {
		return
	}

	for i := 0; i < len(*effects); i++ {
		eff := &(*effects)[i]

		eff.TimeLeft -= dt

		if eff.TickRate > 0 && IsDamageOverTimeEffect(eff.Type) {
			eff.TickTimer += dt
			for eff.TickTimer >= eff.TickRate {
				if takeDamage != nil {
					takeDamage(int(eff.Magnitude))
				}
				eff.TickTimer -= eff.TickRate
			}
		}

		if eff.TimeLeft <= 0 {
			RemoveEffect(effects, i)
			i--
		}
	}
}

func ApplyEffect(effects *[]EffectInstance, newEffect Effect) {
	ApplyEffectWithPolicy(effects, newEffect, DefaultEffectStackPolicy)
}

func ApplyEffectWithPolicy(effects *[]EffectInstance, newEffect Effect, policy EffectStackPolicy) {
	if effects == nil {
		return
	}

	for i := range *effects {
		if (*effects)[i].Type == newEffect.Type {
			(*effects)[i].TimeLeft = newEffect.Duration
			if policy == EffectStackPolicyRefreshDurationKeepStrongestMagnitude && newEffect.Magnitude > (*effects)[i].Magnitude {
				(*effects)[i].Magnitude = newEffect.Magnitude
			}
			return
		}
	}

	*effects = append(*effects, EffectInstance{
		Effect:   newEffect,
		TimeLeft: newEffect.Duration,
	})
}

func RemoveEffect(effects *[]EffectInstance, index int) {
	if index < 0 || index >= len(*effects) {
		return
	}
	*effects = append((*effects)[:index], (*effects)[index+1:]...)
}

func HasEffect(effects *[]EffectInstance, effectType EffectType) bool {
	if effects == nil {
		return false
	}

	for _, e := range *effects {
		if e.Type == effectType {
			return true
		}
	}
	return false
}

func GetEffectMagnitude(effects *[]EffectInstance, effectType EffectType) float32 {
	if effects == nil {
		return 0
	}

	for _, e := range *effects {
		if e.Type == effectType {
			return e.Magnitude
		}
	}
	return 0
}

func IsDamageOverTimeEffect(effectType EffectType) bool {
	return effectType == EffectBurn || effectType == EffectPoison
}

func CanAct(effects *[]EffectInstance) bool {
	return !HasEffect(effects, EffectStun)
}

func CanCast(effects *[]EffectInstance) bool {
	if !CanAct(effects) {
		return false
	}
	return !HasEffect(effects, EffectSilence)
}

func HasCrowdControl(effects *[]EffectInstance) bool {
	return HasEffect(effects, EffectStun) || HasEffect(effects, EffectFreeze) || HasEffect(effects, EffectSlow) || HasEffect(effects, EffectSilence)
}

func MoveSpeedMultiplier(effects *[]EffectInstance) float32 {
	if effects == nil {
		return 1
	}
	if !CanAct(effects) {
		return 0
	}
	if HasEffect(effects, EffectFreeze) {
		return 0
	}

	multiplier := float32(1)
	if HasEffect(effects, EffectSlow) {
		multiplier *= (1 - GetEffectMagnitude(effects, EffectSlow))
	}
	if HasEffect(effects, EffectMoveSpeedReduction) {
		multiplier *= (1 - GetEffectMagnitude(effects, EffectMoveSpeedReduction))
	}
	if HasEffect(effects, EffectMoveSpeedBoost) {
		multiplier *= (1 + GetEffectMagnitude(effects, EffectMoveSpeedBoost))
	}
	if multiplier < 0 {
		return 0
	}
	return multiplier
}
