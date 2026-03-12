package game

import "singlefantasy/app/assets"

const (
	sfxPlayerHit     = "sfx.player.hit"
	sfxEnemyHit      = "sfx.enemy.hit"
	sfxPlayerCast    = "sfx.player.cast"
	sfxEnemyCast     = "sfx.enemy.cast"
	sfxPlayerHealing = "sfx.player.healing"
	sfxPlayerLevelUp = "sfx.player.level_up"
	sfxDoorOpen      = "sfx.door.open"
)

const (
	hitSoundCooldownSeconds  float32 = 0.04
	maxLevelUpSoundsPerGrant int     = 3
)

type healingSoundSource int

const (
	healingSoundPassiveOnHit healingSoundSource = iota
	healingSoundKillReward
	healingSoundActiveSkill
)

func (g *Game) playSound(key string) {
	if g == nil || key == "" {
		return
	}

	cooldown := soundCooldownForKey(key)
	if cooldown > 0 {
		if g.soundCooldowns == nil {
			g.soundCooldowns = map[string]float32{}
		}
		remaining := g.soundCooldowns[key]
		if remaining > 0 {
			return
		}
		g.soundCooldowns[key] = cooldown
	}

	if g.soundPlayer != nil {
		g.soundPlayer(key)
		return
	}
	assets.Get().PlaySound(key)
}

func (g *Game) updateSoundCooldowns(dt float32) {
	if g == nil || dt <= 0 || len(g.soundCooldowns) == 0 {
		return
	}
	for key, remaining := range g.soundCooldowns {
		remaining -= dt
		if remaining <= 0 {
			delete(g.soundCooldowns, key)
			continue
		}
		g.soundCooldowns[key] = remaining
	}
}

func soundCooldownForKey(key string) float32 {
	switch key {
	case sfxPlayerHit, sfxEnemyHit:
		return hitSoundCooldownSeconds
	default:
		return 0
	}
}

func sfxKeyForDamageTarget(target interface{}) string {
	if isPlayerTarget(target) {
		return sfxPlayerHit
	}
	return sfxEnemyHit
}

func (g *Game) playDamageSFX(target interface{}) {
	g.playSound(sfxKeyForDamageTarget(target))
}

func (g *Game) playHealingSFX(source healingSoundSource) {
	if source != healingSoundActiveSkill {
		return
	}
	g.playSound(sfxPlayerHealing)
}

func (g *Game) healPlayerWithFeedbackSource(amount int, source healingSoundSource) int {
	if g == nil || g.Player == nil || amount <= 0 {
		return 0
	}

	before := g.Player.HP
	g.Player.Heal(amount)
	healed := g.Player.HP - before
	if healed > 0 {
		x, y := g.Player.Center()
		g.spawnHealCombatText(x, y, healed)
		g.playHealingSFX(source)
	}
	return healed
}

func (g *Game) healPlayerFromActiveSkillWithFeedback(amount int) int {
	return g.healPlayerWithFeedbackSource(amount, healingSoundActiveSkill)
}

func (g *Game) grantPlayerXP(amount int) {
	if g == nil || g.Player == nil || amount <= 0 {
		return
	}

	beforeLevel := g.Player.Level
	g.Player.GainXP(amount)
	levelsGained := g.Player.Level - beforeLevel
	if levelsGained <= 0 {
		return
	}

	if levelsGained > maxLevelUpSoundsPerGrant {
		levelsGained = maxLevelUpSoundsPerGrant
	}
	for i := 0; i < levelsGained; i++ {
		g.playSound(sfxPlayerLevelUp)
	}
}
