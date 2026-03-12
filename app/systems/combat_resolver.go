package systems

import (
	"math/rand"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
)

type CombatHitRequest struct {
	Caster             *gameobjects.Player
	Target             interface{}
	Skill              *gamedata.Skill
	BaseDamage         int
	DamageType         gamedata.DamageType
	CritChance         float32
	CritMultiplier     float32
	Effects            []gamedata.EffectSpec
	ApplyOnHitHooks    bool
	UseSourceModifiers bool
	SuppressFlash      bool
	CritRoll           *float32
	OnHitProcRoll      *float32
}

type CombatHitResult struct {
	Damage         DamageResult
	EffectsApplied int
	TargetKilled   bool
}

func ApplyCombatHit(request CombatHitRequest) CombatHitResult {
	if request.Target == nil {
		return CombatHitResult{}
	}

	beforeAlive := isTargetAlive(request.Target)
	result := CombatHitResult{}

	damageRequest, hasDamage := buildDamageRequest(request)
	if hasDamage {
		result.Damage = ResolveAndApplyDamage(damageRequest)
	}

	for _, effectSpec := range resolveEffects(request) {
		if applyEffectSpecToTarget(request.Target, effectSpec) {
			result.EffectsApplied++
		}
	}

	if request.ApplyOnHitHooks && request.Caster != nil && result.Damage.AppliedDamage > 0 {
		damageType := resolveDamageType(request)
		applyOnHitHooks(request.Caster, request.Target, result.Damage.AppliedDamage, damageType, request.OnHitProcRoll)
	}

	result.TargetKilled = beforeAlive && !isTargetAlive(request.Target)
	return result
}

func ApplySkill(caster *gameobjects.Player, skill *gamedata.Skill, targets []interface{}) {
	if caster == nil || skill == nil {
		return
	}
	for _, target := range targets {
		ApplyCombatHit(CombatHitRequest{
			Caster:             caster,
			Target:             target,
			Skill:              skill,
			ApplyOnHitHooks:    true,
			UseSourceModifiers: true,
		})
	}
}

func buildDamageRequest(request CombatHitRequest) (DamageRequest, bool) {
	itemCritBonus := itemCritChanceBonus(request.Caster, request.Target)

	if request.Skill != nil && request.Skill.DamageSpec != nil && request.Caster != nil {
		baseDamage := ComputeDamage(request.Skill.DamageSpec, request.Caster.GetEffectiveStats())
		return DamageRequest{
			Source:             request.Caster,
			Target:             request.Target,
			BaseDamage:         baseDamage,
			DamageType:         request.Skill.DamageSpec.DamageType,
			CritChance:         request.CritChance + request.Caster.DerivedStats.CritChance + request.Skill.DamageSpec.CritChance + itemCritBonus,
			CritMultiplier:     request.Skill.DamageSpec.CritMult,
			UseSourceModifiers: request.UseSourceModifiers,
			SuppressFlash:      request.SuppressFlash,
			CritRoll:           request.CritRoll,
		}, true
	}

	if request.BaseDamage <= 0 {
		return DamageRequest{}, false
	}

	totalCritChance := request.CritChance
	if request.Caster != nil {
		totalCritChance += request.Caster.DerivedStats.CritChance
		totalCritChance += itemCritBonus
	}
	return DamageRequest{
		Source:             request.Caster,
		Target:             request.Target,
		BaseDamage:         float32(request.BaseDamage),
		DamageType:         request.DamageType,
		CritChance:         totalCritChance,
		CritMultiplier:     request.CritMultiplier,
		UseSourceModifiers: request.UseSourceModifiers,
		SuppressFlash:      request.SuppressFlash,
		CritRoll:           request.CritRoll,
	}, true
}

func resolveEffects(request CombatHitRequest) []gamedata.EffectSpec {
	if len(request.Effects) > 0 {
		return request.Effects
	}
	if request.Skill != nil {
		return request.Skill.Effects
	}
	return nil
}

func resolveDamageType(request CombatHitRequest) gamedata.DamageType {
	if request.Skill != nil && request.Skill.DamageSpec != nil {
		return request.Skill.DamageSpec.DamageType
	}
	return request.DamageType
}

func applyEffectSpecToTarget(target interface{}, effectSpec gamedata.EffectSpec) bool {
	magnitude := resolveEffectMagnitude(target, effectSpec)
	effect := gamedata.Effect{
		Type:      effectSpec.Type,
		Duration:  effectSpec.Duration,
		Magnitude: magnitude,
		TickRate:  effectSpec.TickRate,
	}

	switch t := target.(type) {
	case *gameobjects.Player:
		gamedata.ApplyEffect(&t.Effects, effect)
		return true
	case *gameobjects.Enemy:
		gamedata.ApplyEffect(&t.Effects, effect)
		return true
	case *gameobjects.Boss:
		gamedata.ApplyEffect(&t.Effects, effect)
		return true
	default:
		return false
	}
}

func resolveEffectMagnitude(target interface{}, effectSpec gamedata.EffectSpec) float32 {
	magnitude := effectSpec.Magnitude
	if effectSpec.PercentMaxHPPerTick <= 0 {
		return magnitude
	}

	maxHP := targetMaxHP(target)
	if maxHP <= 0 {
		return magnitude
	}

	magnitude += float32(maxHP) * effectSpec.PercentMaxHPPerTick

	if effectSpec.MinTickDamage > 0 && magnitude < float32(effectSpec.MinTickDamage) {
		magnitude = float32(effectSpec.MinTickDamage)
	}
	if effectSpec.MaxTickDamage > 0 && magnitude > float32(effectSpec.MaxTickDamage) {
		magnitude = float32(effectSpec.MaxTickDamage)
	}
	return magnitude
}

func targetMaxHP(target interface{}) int {
	switch t := target.(type) {
	case *gameobjects.Player:
		return t.MaxHP
	case *gameobjects.Enemy:
		return t.MaxHP
	case *gameobjects.Boss:
		return t.MaxHP
	default:
		return 0
	}
}

func applyOnHitHooks(caster *gameobjects.Player, target interface{}, appliedDamage int, damageType gamedata.DamageType, procRoll *float32) {
	if caster == nil || appliedDamage <= 0 || target == nil {
		return
	}

	if damageType == gamedata.DamagePhysical && caster.Class != nil && caster.Class.LifestealPercent > 0 {
		caster.Heal(int(float32(appliedDamage) * caster.Class.LifestealPercent))
	}
	if gamedata.HasEffect(&caster.Effects, gamedata.EffectLifesteal) {
		caster.Heal(int(float32(appliedDamage) * gamedata.GetEffectMagnitude(&caster.Effects, gamedata.EffectLifesteal)))
	}

	for _, effect := range caster.GetItemEffects() {
		switch effect.Type {
		case gamedata.ItemEffectLifestealOnHit:
			caster.Heal(int(float32(appliedDamage) * effect.Magnitude))
		case gamedata.ItemEffectManaOnHit:
			if effect.Magnitude > 0 {
				caster.GainMana(int(effect.Magnitude))
			}
		case gamedata.ItemEffectBurnOnHit:
			if !shouldTriggerItemProc(effect.Chance, procRoll) {
				continue
			}

			duration := effect.Duration
			if duration <= 0 {
				duration = 4
			}
			tickRate := effect.TickRate
			if tickRate <= 0 {
				tickRate = 1
			}
			applyEffectSpecToTarget(target, gamedata.EffectSpec{
				Type:      gamedata.EffectBurn,
				Duration:  duration,
				Magnitude: effect.Magnitude,
				TickRate:  tickRate,
			})
		}
	}
}

func shouldTriggerItemProc(chance float32, roll *float32) bool {
	if chance <= 0 {
		chance = 1
	}
	if chance >= 1 {
		return true
	}
	value := rand.Float32()
	if roll != nil {
		value = *roll
	}
	return value < clamp01(chance)
}

func itemCritChanceBonus(caster *gameobjects.Player, target interface{}) float32 {
	if caster == nil || !targetHasSlowedState(target) {
		return 0
	}

	bonus := float32(0)
	for _, effect := range caster.GetItemEffects() {
		if effect.Type == gamedata.ItemEffectCritChanceVsSlowed && effect.Magnitude > 0 {
			bonus += effect.Magnitude
		}
	}
	return bonus
}

func targetHasSlowedState(target interface{}) bool {
	switch t := target.(type) {
	case *gameobjects.Player:
		return gamedata.HasEffect(&t.Effects, gamedata.EffectSlow) ||
			gamedata.HasEffect(&t.Effects, gamedata.EffectFreeze) ||
			gamedata.HasEffect(&t.Effects, gamedata.EffectMoveSpeedReduction)
	case *gameobjects.Enemy:
		return gamedata.HasEffect(&t.Effects, gamedata.EffectSlow) ||
			gamedata.HasEffect(&t.Effects, gamedata.EffectFreeze) ||
			gamedata.HasEffect(&t.Effects, gamedata.EffectMoveSpeedReduction)
	case *gameobjects.Boss:
		return gamedata.HasEffect(&t.Effects, gamedata.EffectSlow) ||
			gamedata.HasEffect(&t.Effects, gamedata.EffectFreeze) ||
			gamedata.HasEffect(&t.Effects, gamedata.EffectMoveSpeedReduction)
	default:
		return false
	}
}

func isTargetAlive(target interface{}) bool {
	switch t := target.(type) {
	case *gameobjects.Player:
		return t.IsAlive()
	case *gameobjects.Enemy:
		return t.IsAlive()
	case *gameobjects.Boss:
		return t.IsAlive()
	default:
		return false
	}
}
