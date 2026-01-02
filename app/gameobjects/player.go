package gameobjects

import "singlefantasy/app/gamedata"

type Player struct {
	X                float32
	Y                float32
	Width            float32
	Height           float32
	Health           int
	MaxHealth        int
	Mana             int
	MaxMana          int
	MoveSpeed        float32
	AttackDamage     int
	AttackRange      float32
	HitFlashTimer    float32
	Class            *gamedata.Class
	Stats            *gamedata.Stats
	Level            int
	XP               int
	XPToNext         int
	StatPoints       int
	Skills           []*gamedata.Skill
	ManaShieldActive bool
	ManaShieldAmount int
	Equipment        map[gamedata.ItemSlot]*gamedata.Item
}

func NewPlayer(x, y float32, classType gamedata.ClassType) *Player {
	class := gamedata.GetClass(classType)
	stats := gamedata.NewStats()

	player := &Player{
		X:             x,
		Y:             y,
		Width:         40,
		Height:        40,
		Health:        100,
		MaxHealth:     100,
		Mana:          50,
		MaxMana:       50,
		MoveSpeed:     200,
		AttackDamage:  10,
		AttackRange:   class.AttackRange,
		HitFlashTimer: 0,
		Class:         class,
		Stats:         stats,
		Level:         1,
		XP:            0,
		XPToNext:      100,
		StatPoints:    0,
		Equipment:     make(map[gamedata.ItemSlot]*gamedata.Item),
	}

	player.ApplyStats()
	player.Skills = gamedata.GetClassSkills(classType)
	return player
}

func (p *Player) ApplyStats() {
	baseStats := *p.Stats

	for _, item := range p.Equipment {
		if item != nil {
			for statType, bonus := range item.StatBonuses {
				baseStats.AddStat(statType, bonus)
			}
		}
	}

	p.MaxHealth = baseStats.CalculateMaxHealth(100)
	if p.Health > p.MaxHealth {
		p.Health = p.MaxHealth
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

	for _, skill := range p.Skills {
		skill.Update(deltaTime)
	}

	if p.Mana < p.MaxMana {
		p.Mana += int(deltaTime * 5)
		if p.Mana > p.MaxMana {
			p.Mana = p.MaxMana
		}
	}
}

func (p *Player) TakeDamage(damage int) {
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

	p.Health -= damage
	if p.Health < 0 {
		p.Health = 0
	}
	p.HitFlashTimer = 0.2
}

func (p *Player) Heal(amount int) {
	p.Health += amount
	if p.Health > p.MaxHealth {
		p.Health = p.MaxHealth
	}
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
	return p.Health > 0
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
