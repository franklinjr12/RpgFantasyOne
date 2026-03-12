package game

import (
	"singlefantasy/app/assets"
	"singlefantasy/app/gamedata"
	"singlefantasy/app/systems"
)

func (g *Game) spawnSkillCastVisual(skill *gamedata.Skill, intent systems.CastIntent) {
	if g == nil || g.Player == nil || skill == nil {
		return
	}
	if skill.Targeting.Type == gamedata.TargetDirection {
		g.spawnDirectionalTelegraph(skill, intent)
	}

	px, py := g.Player.Center()
	x := px
	y := py
	filled := false

	switch skill.Targeting.Type {
	case gamedata.TargetEnemy:
		x = intent.CursorX
		y = intent.CursorY
	case gamedata.TargetArea, gamedata.TargetDirection:
		x = intent.CursorX
		y = intent.CursorY
	}

	if skill.Delivery.Type == gamedata.DeliveryProjectile {
		x = px
		y = py
		filled = true
	}

	g.SkillVisualEffects = append(g.SkillVisualEffects, &SkillVisualEffect{
		X:        x,
		Y:        y,
		Radius:   castVisualRadius(skill),
		Duration: 0.28,
		TimeLeft: 0.28,
		Skill:    skill,
		Filled:   filled,
	})
}

func (g *Game) spawnSkillImpactVisual(skill *gamedata.Skill, x, y float32) {
	if g == nil || skill == nil {
		return
	}

	g.SkillVisualEffects = append(g.SkillVisualEffects, &SkillVisualEffect{
		X:        x,
		Y:        y,
		Radius:   impactVisualRadius(skill),
		Duration: 0.2,
		TimeLeft: 0.2,
		Skill:    skill,
		Filled:   true,
	})
}

func (g *Game) updateSkillVisualEffects(dt float32) {
	if g == nil {
		return
	}

	for i := len(g.SkillVisualEffects) - 1; i >= 0; i-- {
		effect := g.SkillVisualEffects[i]
		if effect == nil || effect.TimeLeft <= 0 {
			g.SkillVisualEffects = append(g.SkillVisualEffects[:i], g.SkillVisualEffects[i+1:]...)
			continue
		}
		effect.TimeLeft -= dt
		if effect.TimeLeft <= 0 {
			g.SkillVisualEffects = append(g.SkillVisualEffects[:i], g.SkillVisualEffects[i+1:]...)
		}
	}
}

func (g *Game) playSkillCastSFX(skill *gamedata.Skill) {
	if skill == nil {
		return
	}
	assets.Get().PlaySound(skillCastSFXKey(skill))
}

func (g *Game) playSkillImpactSFX(skill *gamedata.Skill) {
	if skill == nil {
		return
	}
	assets.Get().PlaySound(skillImpactSFXKey(skill))
}

func skillCastSFXKey(skill *gamedata.Skill) string {
	switch skill.Type {
	case gamedata.SkillTypePowerStrike, gamedata.SkillTypeGuardStance, gamedata.SkillTypeBloodOath, gamedata.SkillTypeShockwaveSlam:
		return "sfx.skill.cast.melee"
	case gamedata.SkillTypeQuickShot, gamedata.SkillTypeRetreatRoll, gamedata.SkillTypeFocusedAim, gamedata.SkillTypePoisonTip:
		return "sfx.skill.cast.ranged"
	default:
		return "sfx.skill.cast.caster"
	}
}

func skillImpactSFXKey(skill *gamedata.Skill) string {
	if skill.DamageSpec != nil && skill.DamageSpec.DamageType == gamedata.DamageMagical {
		return "sfx.skill.impact.magic"
	}
	return "sfx.skill.impact.physical"
}

func castVisualRadius(skill *gamedata.Skill) float32 {
	switch skill.Type {
	case gamedata.SkillTypePowerStrike:
		return 32
	case gamedata.SkillTypeGuardStance:
		return 48
	case gamedata.SkillTypeBloodOath:
		return 44
	case gamedata.SkillTypeShockwaveSlam:
		return 74
	case gamedata.SkillTypeQuickShot:
		return 24
	case gamedata.SkillTypeRetreatRoll:
		return 40
	case gamedata.SkillTypeFocusedAim:
		return 40
	case gamedata.SkillTypePoisonTip:
		return 32
	case gamedata.SkillTypeArcaneBolt:
		return 28
	case gamedata.SkillTypeManaShield:
		return 52
	case gamedata.SkillTypeFrostField:
		return 80
	case gamedata.SkillTypeArcaneDrain:
		return 68
	default:
		return 36
	}
}

func impactVisualRadius(skill *gamedata.Skill) float32 {
	switch skill.Type {
	case gamedata.SkillTypePowerStrike:
		return 26
	case gamedata.SkillTypeShockwaveSlam:
		return 60
	case gamedata.SkillTypeQuickShot:
		return 18
	case gamedata.SkillTypePoisonTip:
		return 30
	case gamedata.SkillTypeArcaneBolt:
		return 34
	case gamedata.SkillTypeFrostField:
		return 54
	case gamedata.SkillTypeArcaneDrain:
		return 48
	default:
		return 26
	}
}
