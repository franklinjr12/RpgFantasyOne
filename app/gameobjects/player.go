package gameobjects

import (
	"singlefantasy/app/core"
	"singlefantasy/app/gamedata"
)

type Player struct {
	core.Entity
	Mana                  int
	MaxMana               int
	MoveSpeed             float32
	AttackDamage          int
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
	Equipment             map[gamedata.ItemSlot]*gamedata.Item
	AttackCooldown        float32
	CurrentAttackCooldown float32
}

func NewPlayer(x, y float32, classType gamedata.ClassType) *Player {
	class := gamedata.GetClassData(classType)
	stats := gamedata.NewStats()

	player := &Player{
		Entity: core.Entity{
			PosX:    x,
			PosY:    y,
			HP:      100,
			MaxHP:   100,
			Stats:   stats,
			Hitbox:  core.Hitbox{Width: 40, Height: 40},
			Faction: core.FactionPlayer,
			Alive:   true,
		},
		Mana:                  50,
		MaxMana:               50,
		MoveSpeed:             200,
		AttackDamage:          10,
		AttackRange:           class.AttackRange,
		HitFlashTimer:         0,
		Class:                 class,
		Level:                 1,
		XP:                    0,
		XPToNext:              100,
		StatPoints:            0,
		Equipment:             make(map[gamedata.ItemSlot]*gamedata.Item),
		AttackCooldown:        1.0,
		CurrentAttackCooldown: 0,
	}

	player.ApplyStats()
	player.Skills = gamedata.GetClassSkillData(classType)
	return player
}

func (p *Player) ApplyStats() {
	baseStats := *p.Entity.Stats

	for _, item := range p.Equipment {
		if item != nil {
			for statType, bonus := range item.StatBonuses {
				baseStats.AddStat(statType, bonus)
			}
		}
	}

	p.MaxHP = baseStats.CalculateMaxHealth(100)
	if p.HP > p.MaxHP {
		p.HP = p.MaxHP
	}

	p.MaxMana = baseStats.CalculateMaxMana(50)
	if p.Mana > p.MaxMana {
		p.Mana = p.MaxMana
	}

	p.MoveSpeed = baseStats.CalculateMoveSpeed(200)

	switch p.Class.Type {
	case gamedata.ClassTypeMelee:
		p.AttackDamage = baseStats.CalculatePhysicalDamage(10)
	case gamedata.ClassTypeRanged:
		p.AttackDamage = baseStats.CalculateRangedDamage(10)
	case gamedata.ClassTypeCaster:
		p.AttackDamage = baseStats.CalculateMagicDamage(15)
	}
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

	for _, skill := range p.Skills {
		skill.Update(deltaTime)
	}

	gamedata.UpdateEffects(&p.Entity.Effects, deltaTime, p.TakeDamage)

	if p.Mana < p.MaxMana {
		p.Mana += int(deltaTime * 5)
		if p.Mana > p.MaxMana {
			p.Mana = p.MaxMana
		}
	}
}

func (p *Player) TakeDamage(damage int) {
	if gamedata.HasEffect(&p.Entity.Effects, gamedata.EffectDamageReduction) {
		magnitude := gamedata.GetEffectMagnitude(&p.Entity.Effects, gamedata.EffectDamageReduction)
		damage = int(float32(damage) * (1.0 - magnitude))
	}

	if p.ManaShieldActive && p.ManaShieldAmount > 0 {
		if damage <= p.ManaShieldAmount {
			p.ManaShieldAmount -= damage
			damage = 0
		} else {
			damage -= p.ManaShieldAmount
			p.ManaShieldAmount = 0
			p.ManaShieldActive = false
		}
	}

	p.Entity.ApplyDamage(damage)
	p.HitFlashTimer = 0.2
}

func (p *Player) Heal(amount int) {
	p.Entity.Heal(amount)
}

func (p *Player) GainXP(amount int) {
	p.XP += amount
	for p.XP >= p.XPToNext {
		p.XP -= p.XPToNext
		p.Level++
		p.StatPoints += 3
		p.XPToNext = p.Level * 100
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

func (p *Player) GetAttackCooldown() float32 {
	baseStats := *p.Entity.Stats
	for _, item := range p.Equipment {
		if item != nil {
			for statType, bonus := range item.StatBonuses {
				baseStats.AddStat(statType, bonus)
			}
		}
	}
	attackSpeed := baseStats.CalculateAttackSpeed(1.0)
	return p.AttackCooldown / attackSpeed
}

func (p *Player) GetAutoAttackDamage() int {
	baseStats := *p.Entity.Stats
	for _, item := range p.Equipment {
		if item != nil {
			for statType, bonus := range item.StatBonuses {
				baseStats.AddStat(statType, bonus)
			}
		}
	}
	return baseStats.CalculateAutoAttackDamage(p.AttackDamage)
}
