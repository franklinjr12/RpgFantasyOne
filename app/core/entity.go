package core

import "singlefantasy/app/gamedata"

type Faction int

const (
	FactionPlayer Faction = iota
	FactionEnemy
	FactionNeutral
)

type Hitbox struct {
	Width  float32
	Height float32
}

type Entity struct {
	PosX    float32
	PosY    float32
	VelX    float32
	VelY    float32
	HP      int
	MaxHP   int
	Stats   *gamedata.Stats
	Hitbox  Hitbox
	Effects []gamedata.EffectInstance
	Faction Faction
	Alive   bool
}

func (e *Entity) Center() (float32, float32) {
	return e.PosX + e.Hitbox.Width/2, e.PosY + e.Hitbox.Height/2
}

func (e *Entity) IsAlive() bool {
	return e != nil && e.Alive && e.HP > 0
}

func (e *Entity) ApplyDamage(amount int) int {
	if e == nil || amount <= 0 || !e.Alive {
		return 0
	}

	if amount > e.HP {
		amount = e.HP
	}
	e.HP -= amount
	if e.HP <= 0 {
		e.HP = 0
		e.Alive = false
	}

	return amount
}

func (e *Entity) Heal(amount int) int {
	if e == nil || amount <= 0 || !e.Alive {
		return 0
	}

	previous := e.HP
	e.HP += amount
	if e.HP > e.MaxHP {
		e.HP = e.MaxHP
	}
	return e.HP - previous
}
