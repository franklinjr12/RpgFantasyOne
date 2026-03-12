package gameobjects

import (
	"math"
	"singlefantasy/app/core"
	"singlefantasy/app/gamedata"
)

type PlayerAttackState int

const (
	PlayerAttackStateIdle PlayerAttackState = iota
	PlayerAttackStateWindup
	PlayerAttackStateRecover
)

type Player struct {
	core.Entity
	Mana                  int
	MaxMana               int
	MoveSpeed             float32
	AttackDamage          int
	EffectiveStats        gamedata.Stats
	DerivedStats          gamedata.DerivedStats
	AttackRange           float32
	HitFlashTimer         float32
	Class                 *gamedata.Class
	Level                 int
	XP                    int
	XPToNext              int
	StatPoints            int
	Skills                []*gamedata.Skill
	ManaShieldActive      bool
	ManaShieldAmount      int
	ManaShieldTimeLeft    float32
	Equipment             map[gamedata.ItemSlot]*gamedata.Item
	AttackCooldown        float32
	CurrentAttackCooldown float32
	FacingRight           bool
	MoveVelocityX         float32
	MoveVelocityY         float32
	AttackState           PlayerAttackState
	AttackStateTimer      float32
	HurtIFrameTimer       float32
	KnockbackVelX         float32
	KnockbackVelY         float32
	ManaRegenRemainder    float32
}

func NewPlayer(x, y float32, classType gamedata.ClassType) *Player {
	class := gamedata.GetClassData(classType)
	stats := gamedata.GetClassBaseStats(classType)

	player := &Player{
		Entity: core.Entity{
			PosX:    x,
			PosY:    y,
			HP:      gamedata.BasePlayerHP,
			MaxHP:   gamedata.BasePlayerHP,
			Stats:   stats,
			Hitbox:  core.Hitbox{Width: 40, Height: 40},
			Faction: core.FactionPlayer,
			Alive:   true,
		},
		Mana:                  gamedata.BasePlayerMana,
		MaxMana:               gamedata.BasePlayerMana,
		MoveSpeed:             gamedata.BasePlayerMoveSpeed,
		AttackDamage:          gamedata.BaseMeleeAutoAttackDamage,
		AttackRange:           class.AttackRange,
		HitFlashTimer:         0,
		Class:                 class,
		Level:                 1,
		XP:                    0,
		XPToNext:              gamedata.XPToNextLevel(1),
		StatPoints:            0,
		Equipment:             make(map[gamedata.ItemSlot]*gamedata.Item),
		AttackCooldown:        1.0,
		CurrentAttackCooldown: 0,
		FacingRight:           true,
		MoveVelocityX:         0,
		MoveVelocityY:         0,
		AttackState:           PlayerAttackStateIdle,
		AttackStateTimer:      0,
		HurtIFrameTimer:       0,
		KnockbackVelX:         0,
		KnockbackVelY:         0,
		ManaRegenRemainder:    0,
	}

	player.ApplyStats()
	player.Skills = gamedata.GetClassSkillData(classType)
	return player
}

func (p *Player) ApplyStats() {
	if p.Entity.Stats == nil {
		p.Entity.Stats = gamedata.NewStats()
	}

	classType := gamedata.ClassTypeMelee
	if p.Class != nil {
		classType = p.Class.Type
	}

	p.EffectiveStats = gamedata.ComputeEffectiveStats(p.Entity.Stats, p.Equipment)
	p.DerivedStats = gamedata.ComputeDerivedStats(classType, p.EffectiveStats)
	p.MaxHP = p.DerivedStats.MaxHP
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}

	p.MaxMana = p.DerivedStats.MaxMana
	if p.Mana > p.MaxMana {
		p.Mana = p.MaxMana
	}

	p.MoveSpeed = p.DerivedStats.MoveSpeed
	p.AttackDamage = p.DerivedStats.AutoAttackDamage
}

func (p *Player) EquipItem(item *gamedata.Item) {
	p.Equipment[item.Slot] = item
	p.ApplyStats()
}

func (p *Player) Update(deltaTime float32) {
	if p.HitFlashTimer > 0 {
		p.HitFlashTimer -= deltaTime
		if p.HitFlashTimer < 0 {
			p.HitFlashTimer = 0
		}
	}

	if p.CurrentAttackCooldown > 0 {
		p.CurrentAttackCooldown -= deltaTime
		if p.CurrentAttackCooldown < 0 {
			p.CurrentAttackCooldown = 0
		}
	}

	if p.HurtIFrameTimer > 0 {
		p.HurtIFrameTimer -= deltaTime
		if p.HurtIFrameTimer < 0 {
			p.HurtIFrameTimer = 0
		}
	}

	if p.ManaShieldActive && p.ManaShieldTimeLeft > 0 {
		p.ManaShieldTimeLeft -= deltaTime
		if p.ManaShieldTimeLeft <= 0 {
			p.ManaShieldTimeLeft = 0
			p.ManaShieldAmount = 0
			p.ManaShieldActive = false
		}
	}

	for _, skill := range p.Skills {
		skill.Update(deltaTime)
	}

	gamedata.UpdateEffects(&p.Entity.Effects, deltaTime, p.TakeDamage)

	p.regenerateMana(deltaTime)
}

func (p *Player) TakeDamage(damage int) {
	p.takeDamageInternal(damage, gamedata.DamagePhysical, true)
}

func (p *Player) TakeDamageWithoutFlash(damage int) {
	p.takeDamageInternal(damage, gamedata.DamagePhysical, false)
}

func (p *Player) TakeTypedDamage(damage int, damageType gamedata.DamageType) {
	p.takeDamageInternal(damage, damageType, true)
}

func (p *Player) TakeTypedDamageWithoutFlash(damage int, damageType gamedata.DamageType) {
	p.takeDamageInternal(damage, damageType, false)
}

func (p *Player) ApplyTypedDamage(damage int, damageType gamedata.DamageType, flash bool) int {
	return p.takeDamageInternal(damage, damageType, flash)
}

func (p *Player) takeDamageInternal(damage int, damageType gamedata.DamageType, flash bool) int {
	if damage <= 0 {
		return 0
	}

	if gamedata.HasEffect(&p.Entity.Effects, gamedata.EffectDamageReduction) {
		magnitude := gamedata.GetEffectMagnitude(&p.Entity.Effects, gamedata.EffectDamageReduction)
		damage = int(float32(damage) * (1.0 - magnitude))
	}

	switch damageType {
	case gamedata.DamagePhysical:
		damage = applyResistance(damage, p.DerivedStats.PhysicalResist)
	case gamedata.DamageMagical:
		damage = applyResistance(damage, p.DerivedStats.MagicalResist)
	}

	if p.ManaShieldActive && p.ManaShieldAmount > 0 {
		if damage <= p.ManaShieldAmount {
			p.ManaShieldAmount -= damage
			damage = 0
		} else {
			damage -= p.ManaShieldAmount
			p.ManaShieldAmount = 0
			p.ManaShieldActive = false
			p.ManaShieldTimeLeft = 0
		}
	}

	applied := p.Entity.ApplyDamage(damage)
	if flash {
		p.HitFlashTimer = EntityHitFlashDuration
	}
	return applied
}

func (p *Player) CanTakeDirectHit() bool {
	return p.HurtIFrameTimer <= 0
}

func (p *Player) StartHurtIFrames(duration float32) {
	if duration <= 0 {
		return
	}
	p.HurtIFrameTimer = duration
}

func (p *Player) ApplyKnockbackFrom(sourceX, sourceY, impulse float32) {
	if impulse <= 0 {
		return
	}

	centerX, centerY := p.Center()
	dx := centerX - sourceX
	dy := centerY - sourceY
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if distance <= 0 {
		return
	}

	p.KnockbackVelX += (dx / distance) * impulse
	p.KnockbackVelY += (dy / distance) * impulse
}

func (p *Player) Heal(amount int) {
	p.Entity.Heal(amount)
}

func (p *Player) GainXP(amount int) {
	if amount <= 0 {
		return
	}

	if p.Entity.Stats == nil {
		p.Entity.Stats = gamedata.NewStats()
	}

	p.XP += amount
	leveledUp := false
	for p.XP >= p.XPToNext {
		p.XP -= p.XPToNext
		p.Level++
		p.StatPoints += gamedata.LevelUpStatPoints

		if p.Class != nil {
			p.Stats.AddStat(p.Class.GrowthBias, gamedata.LevelUpGrowthStatPoints)
		}

		p.XPToNext = gamedata.XPToNextLevel(p.Level)
		leveledUp = true
	}

	if leveledUp {
		p.ApplyStats()
	}
}

func (p *Player) AddStatPoint(statType gamedata.StatType) {
	if p.StatPoints > 0 {
		p.Stats.AddStat(statType, 1)
		p.StatPoints--
		p.ApplyStats()
	}
}

func (p *Player) IsAlive() bool {
	return p.Entity.IsAlive()
}

func (p *Player) CanUseMana(amount int) bool {
	return p.Mana >= amount
}

func (p *Player) UseMana(amount int) {
	p.Mana -= amount
	if p.Mana < 0 {
		p.Mana = 0
	}
}

func (p *Player) GainMana(amount int) {
	if amount <= 0 {
		return
	}
	p.Mana += amount
	if p.Mana > p.MaxMana {
		p.Mana = p.MaxMana
	}
}

func (p *Player) SetManaShield(amount int, duration float32) {
	if amount <= 0 {
		p.ManaShieldAmount = 0
		p.ManaShieldTimeLeft = 0
		p.ManaShieldActive = false
		return
	}

	p.ManaShieldAmount = amount
	p.ManaShieldActive = true
	if duration > 0 {
		p.ManaShieldTimeLeft = duration
	} else {
		p.ManaShieldTimeLeft = 0
	}
}

func (p *Player) GetAttackCooldown() float32 {
	attackSpeed := p.DerivedStats.AttackSpeedMultiplier
	if attackSpeed <= 0 {
		return p.AttackCooldown
	}
	return p.AttackCooldown / attackSpeed
}

func (p *Player) GetAutoAttackDamage() int {
	return p.DerivedStats.AutoAttackDamage
}

func (p *Player) GetEffectiveStats() *gamedata.Stats {
	return &p.EffectiveStats
}

func (p *Player) GetEquippedItems() []*gamedata.Item {
	if p == nil || len(p.Equipment) == 0 {
		return nil
	}

	items := make([]*gamedata.Item, 0, len(p.Equipment))
	for _, item := range p.Equipment {
		if item == nil {
			continue
		}
		items = append(items, item)
	}
	return items
}

func (p *Player) GetItemEffects() []gamedata.ItemEffect {
	if p == nil {
		return nil
	}

	effects := make([]gamedata.ItemEffect, 0, len(p.Equipment))
	for _, item := range p.Equipment {
		if item == nil || len(item.Effects) == 0 {
			continue
		}
		effects = append(effects, item.Effects...)
	}
	return effects
}

func applyResistance(damage int, resistance float32) int {
	if damage <= 0 {
		return 0
	}

	mitigated := int(float32(damage) * (1.0 - resistance))
	if mitigated < 1 {
		return 1
	}
	return mitigated
}

func (p *Player) regenerateMana(deltaTime float32) {
	if p == nil || deltaTime <= 0 {
		return
	}
	if p.Mana >= p.MaxMana {
		p.ManaRegenRemainder = 0
		return
	}

	regenPerSec := float32(2)
	if p.Class != nil && p.Class.ManaRegenPerSec > 0 {
		regenPerSec = p.Class.ManaRegenPerSec
	}

	p.ManaRegenRemainder += deltaTime * regenPerSec
	wholePoints := int(p.ManaRegenRemainder)
	if wholePoints <= 0 {
		return
	}

	p.ManaRegenRemainder -= float32(wholePoints)
	p.GainMana(wholePoints)
}
