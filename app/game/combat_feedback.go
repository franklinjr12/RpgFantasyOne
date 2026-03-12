package game

import (
	"fmt"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/systems"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type CombatTextKind int

const (
	CombatTextDamage CombatTextKind = iota
	CombatTextHeal
	CombatTextStatus
)

type CombatTextEvent struct {
	X         float32
	Y         float32
	Text      string
	Kind      CombatTextKind
	Duration  float32
	TimeLeft  float32
	RiseSpeed float32
	Color     rl.Color
	Scale     float32
	IsCrit    bool
}

type DirectionalTelegraphEvent struct {
	StartX   float32
	StartY   float32
	EndX     float32
	EndY     float32
	Width    float32
	Duration float32
	TimeLeft float32
	Skill    *gamedata.Skill
}

var (
	combatDamageEnemyColor  = rl.NewColor(255, 232, 142, 255)
	combatDamagePlayerColor = rl.NewColor(255, 106, 106, 255)
	combatDamageCritColor   = rl.NewColor(255, 198, 96, 255)
	combatHealColor         = rl.NewColor(96, 230, 128, 255)
	combatStatusColor       = rl.NewColor(182, 224, 255, 255)
)

func (g *Game) applySkillWithFeedback(caster *gameobjects.Player, skill *gamedata.Skill, targets []interface{}) int {
	if g == nil || caster == nil || skill == nil {
		return 0
	}

	hits := 0
	for _, target := range targets {
		if target == nil {
			continue
		}
		g.applyCombatHitWithFeedback(systems.CombatHitRequest{
			Caster:             caster,
			Target:             target,
			Skill:              skill,
			ApplyOnHitHooks:    true,
			UseSourceModifiers: true,
		})
		hits++
	}
	return hits
}

func (g *Game) applyCombatHitWithFeedback(request systems.CombatHitRequest) systems.CombatHitResult {
	if g == nil || request.Target == nil {
		return systems.CombatHitResult{}
	}

	casterHPBefore := 0
	if request.Caster != nil {
		casterHPBefore = request.Caster.HP
	}

	result := systems.ApplyCombatHit(request)
	if result.Damage.AppliedDamage > 0 {
		g.spawnDamageCombatText(request.Target, result.Damage.AppliedDamage, result.Damage.IsCrit, isPlayerTarget(request.Target))
	}
	if result.EffectsApplied > 0 {
		g.spawnStatusPopupsForTarget(request.Target, feedbackEffectsFromRequest(request))
	}

	if request.Caster != nil {
		healed := request.Caster.HP - casterHPBefore
		if healed > 0 {
			x, y := request.Caster.Center()
			g.spawnHealCombatText(x, y, healed)
		}
	}

	return result
}

func feedbackEffectsFromRequest(request systems.CombatHitRequest) []gamedata.EffectSpec {
	if len(request.Effects) > 0 {
		return request.Effects
	}
	if request.Skill != nil {
		return request.Skill.Effects
	}
	return nil
}

func isPlayerTarget(target interface{}) bool {
	_, ok := target.(*gameobjects.Player)
	return ok
}

func (g *Game) healPlayerWithFeedback(amount int) int {
	if g == nil || g.Player == nil || amount <= 0 {
		return 0
	}

	before := g.Player.HP
	g.Player.Heal(amount)
	healed := g.Player.HP - before
	if healed > 0 {
		x, y := g.Player.Center()
		g.spawnHealCombatText(x, y, healed)
	}
	return healed
}

func (g *Game) spawnDamageCombatText(target interface{}, amount int, isCrit bool, targetIsPlayer bool) {
	if g == nil || amount <= 0 {
		return
	}

	x, y, ok := combatFeedbackTargetAnchor(target)
	if !ok {
		return
	}

	color := combatDamageEnemyColor
	if targetIsPlayer {
		color = combatDamagePlayerColor
	}
	scale := CombatFeedbackBaseScale
	text := fmt.Sprintf("%d", amount)
	if isCrit {
		color = combatDamageCritColor
		scale = CombatFeedbackCritScale
		text = fmt.Sprintf("!%d", amount)
	}
	g.addCombatTextEvent(x, y, text, CombatTextDamage, color, CombatFeedbackTextDuration, scale, isCrit)
}

func (g *Game) spawnHealCombatText(x, y float32, amount int) {
	if g == nil || amount <= 0 {
		return
	}
	g.addCombatTextEvent(x, y, fmt.Sprintf("+%d", amount), CombatTextHeal, combatHealColor, CombatFeedbackTextDuration, CombatFeedbackBaseScale, false)
}

func (g *Game) spawnStatusPopupsForTarget(target interface{}, effects []gamedata.EffectSpec) {
	if g == nil || len(effects) == 0 {
		return
	}
	x, y, ok := combatFeedbackTargetAnchor(target)
	if !ok {
		return
	}

	seen := map[string]struct{}{}
	for _, effect := range effects {
		label := gamedata.EffectPopupLabel(effect.Type)
		if label == "" {
			continue
		}
		if _, exists := seen[label]; exists {
			continue
		}
		seen[label] = struct{}{}
		g.addCombatTextEvent(x, y, label, CombatTextStatus, combatStatusColor, CombatFeedbackStatusDuration, CombatFeedbackBaseScale, false)
	}
}

func combatFeedbackTargetAnchor(target interface{}) (float32, float32, bool) {
	switch t := target.(type) {
	case *gameobjects.Player:
		if t == nil {
			return 0, 0, false
		}
		x, y := t.Center()
		return x, y - t.Hitbox.Height*0.55, true
	case *gameobjects.Enemy:
		if t == nil {
			return 0, 0, false
		}
		x, y := t.Center()
		return x, y - t.Hitbox.Height*0.55, true
	case *gameobjects.Boss:
		if t == nil {
			return 0, 0, false
		}
		x, y := t.Center()
		return x, y - t.Hitbox.Height*0.55, true
	default:
		return 0, 0, false
	}
}

func (g *Game) addCombatTextEvent(x, y float32, text string, kind CombatTextKind, color rl.Color, duration, scale float32, isCrit bool) {
	if g == nil || duration <= 0 || text == "" {
		return
	}
	jitterX, stackOffset := g.nextCombatTextOffset()
	g.CombatTextEvents = append(g.CombatTextEvents, &CombatTextEvent{
		X:         x + jitterX,
		Y:         y - CombatFeedbackTextBaseLift - stackOffset,
		Text:      text,
		Kind:      kind,
		Duration:  duration,
		TimeLeft:  duration,
		RiseSpeed: CombatFeedbackTextRiseSpeed,
		Color:     color,
		Scale:     scale,
		IsCrit:    isCrit,
	})
}

func (g *Game) nextCombatTextOffset() (float32, float32) {
	if g == nil {
		return 0, 0
	}
	jitterPattern := [...]float32{-10, -5, 0, 5, 10}
	stackPattern := [...]float32{0, CombatFeedbackTextStackSpacing, CombatFeedbackTextStackSpacing * 2}
	index := g.CombatFeedbackSequence
	g.CombatFeedbackSequence++
	return jitterPattern[index%len(jitterPattern)], stackPattern[(index/len(jitterPattern))%len(stackPattern)]
}

func (g *Game) spawnDirectionalTelegraph(skill *gamedata.Skill, intent systems.CastIntent) {
	if g == nil || g.Player == nil || skill == nil {
		return
	}
	if skill.Targeting.Type != gamedata.TargetDirection {
		return
	}

	startX, startY := g.Player.Center()
	rangeValue := skill.Targeting.Range
	if rangeValue <= 0 {
		rangeValue = CombatFeedbackDirectionalDefaultRange
	}

	dirX := intent.DirectionX
	dirY := intent.DirectionY
	if dirX == 0 && dirY == 0 {
		if g.Player.FacingRight {
			dirX = 1
		} else {
			dirX = -1
		}
	}

	width := skill.Targeting.DirectionalLineWidth
	if width <= 0 {
		width = CombatFeedbackDirectionalDefaultWidth
	}

	g.DirectionalTelegraphs = append(g.DirectionalTelegraphs, &DirectionalTelegraphEvent{
		StartX:   startX,
		StartY:   startY,
		EndX:     startX + dirX*rangeValue,
		EndY:     startY + dirY*rangeValue,
		Width:    width,
		Duration: CombatFeedbackDirectionalDuration,
		TimeLeft: CombatFeedbackDirectionalDuration,
		Skill:    skill,
	})
}

func (g *Game) updateCombatFeedback(dt float32) {
	if g == nil {
		return
	}

	for i := len(g.CombatTextEvents) - 1; i >= 0; i-- {
		event := g.CombatTextEvents[i]
		if event == nil || event.TimeLeft <= 0 {
			g.CombatTextEvents = append(g.CombatTextEvents[:i], g.CombatTextEvents[i+1:]...)
			continue
		}
		event.TimeLeft -= dt
		event.Y -= event.RiseSpeed * dt
		if event.TimeLeft <= 0 {
			g.CombatTextEvents = append(g.CombatTextEvents[:i], g.CombatTextEvents[i+1:]...)
		}
	}

	for i := len(g.DirectionalTelegraphs) - 1; i >= 0; i-- {
		telegraph := g.DirectionalTelegraphs[i]
		if telegraph == nil || telegraph.TimeLeft <= 0 {
			g.DirectionalTelegraphs = append(g.DirectionalTelegraphs[:i], g.DirectionalTelegraphs[i+1:]...)
			continue
		}
		telegraph.TimeLeft -= dt
		if telegraph.TimeLeft <= 0 {
			g.DirectionalTelegraphs = append(g.DirectionalTelegraphs[:i], g.DirectionalTelegraphs[i+1:]...)
		}
	}
}

func (g *Game) drawCombatFeedback() {
	if g == nil || g.Camera == nil {
		return
	}

	for _, telegraph := range g.DirectionalTelegraphs {
		if telegraph == nil || telegraph.TimeLeft <= 0 || telegraph.Duration <= 0 {
			continue
		}
		remainingRatio := telegraph.TimeLeft / telegraph.Duration
		systems.DrawDirectionalTelegraph(
			telegraph.StartX,
			telegraph.StartY,
			telegraph.EndX,
			telegraph.EndY,
			telegraph.Width,
			remainingRatio,
			telegraph.Skill,
			g.Camera,
		)
	}

	for _, event := range g.CombatTextEvents {
		if event == nil || event.TimeLeft <= 0 || event.Duration <= 0 {
			continue
		}
		remainingRatio := clampRatio(event.TimeLeft / event.Duration)
		alpha := uint8(255 * remainingRatio)
		color := rl.NewColor(event.Color.R, event.Color.G, event.Color.B, alpha)
		shadow := rl.NewColor(0, 0, 0, uint8(float32(alpha)*0.65))

		scale := event.Scale
		if scale <= 0 {
			scale = CombatFeedbackBaseScale
		}
		fontSize := int32(float32(CombatFeedbackBaseFontSize) * scale)
		if fontSize < 12 {
			fontSize = 12
		}

		screenX, screenY := systems.WorldToScreenIso(event.X, event.Y, g.Camera)
		textWidth := rl.MeasureText(event.Text, fontSize)
		textX := int32(screenX) - textWidth/2
		textY := int32(screenY) - fontSize/2
		rl.DrawText(event.Text, textX+1, textY+1, fontSize, shadow)
		rl.DrawText(event.Text, textX, textY, fontSize, color)
	}
}

func clampRatio(value float32) float32 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
