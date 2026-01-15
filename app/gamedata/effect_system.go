package gamedata

func UpdateEffects(effects *[]EffectInstance, dt float32, takeDamage func(int)) {
	if effects == nil {
		return
	}

	for i := 0; i < len(*effects); i++ {
		eff := &(*effects)[i]

		eff.TimeLeft -= dt

		if eff.TickRate > 0 {
			eff.TickTimer += dt
			if eff.TickTimer >= eff.TickRate {
				if takeDamage != nil {
					takeDamage(int(eff.Magnitude))
				}
				eff.TickTimer = 0
			}
		}

		if eff.TimeLeft <= 0 {
			RemoveEffect(effects, i)
			i--
		}
	}
}

func ApplyEffect(effects *[]EffectInstance, newEffect Effect) {
	if effects == nil {
		return
	}

	for i := range *effects {
		if (*effects)[i].Type == newEffect.Type {
			(*effects)[i].TimeLeft = newEffect.Duration
			if newEffect.Magnitude > (*effects)[i].Magnitude {
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

