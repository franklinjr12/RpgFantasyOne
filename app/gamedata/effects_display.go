package gamedata

func EffectPopupLabel(effectType EffectType) string {
	switch effectType {
	case EffectStun:
		return "STUNNED"
	case EffectSilence:
		return "SILENCED"
	case EffectFreeze:
		return "FROZEN"
	case EffectSlow:
		return "SLOWED"
	case EffectBurn:
		return "BURNING"
	case EffectPoison:
		return "POISONED"
	default:
		return ""
	}
}
