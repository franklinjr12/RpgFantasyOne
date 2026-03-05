package gameobjects

import (
	"math"
	"singlefantasy/app/core"
	"singlefantasy/app/gamedata"
)

type EnemyState int

const (
	EnemyStateIdle EnemyState = iota
	EnemyStateChasing
	EnemyStateAttacking
)

type Enemy struct {
	core.Entity
	Damage           int
	MoveSpeed        float32
	AttackCooldown   float32
	CurrentCooldown  float32
	AttackRange      float32
	AggroRange       float32
	HitFlashTimer    float32
	AttackFlashTimer float32
	FacingRight      bool
	State            EnemyState
	IsElite          bool
}

func NewEnemy(x, y float32, isElite bool) *Enemy {
	template := gamedata.GetEnemyTemplateByTier(isElite)

	return &Enemy{
		Entity: core.Entity{
			PosX:    x,
			PosY:    y,
			HP:      template.MaxHP,
			MaxHP:   template.MaxHP,
			Stats:   nil,
			Hitbox:  core.Hitbox{Width: template.Width, Height: template.Height},
			Faction: core.FactionEnemy,
			Alive:   true,
		},
		Damage:           template.Damage,
		MoveSpeed:        template.MoveSpeed,
		AttackCooldown:   template.AttackCooldown,
		CurrentCooldown:  0,
		AttackRange:      template.AttackRange,
		AggroRange:       template.AggroRange,
		HitFlashTimer:    0,
		AttackFlashTimer: 0,
		FacingRight:      true,
		State:            EnemyStateIdle,
		IsElite:          isElite,
	}
}

func (e *Enemy) Update(deltaTime float32, playerX, playerY float32) {
	if !e.Entity.IsAlive() {
		return
	}

	if e.CurrentCooldown > 0 {
		e.CurrentCooldown -= deltaTime
		if e.CurrentCooldown < 0 {
			e.CurrentCooldown = 0
		}
	}

	if e.HitFlashTimer > 0 {
		e.HitFlashTimer -= deltaTime
		if e.HitFlashTimer < 0 {
			e.HitFlashTimer = 0
		}
	}

	if e.AttackFlashTimer > 0 {
		e.AttackFlashTimer -= deltaTime
		if e.AttackFlashTimer < 0 {
			e.AttackFlashTimer = 0
		}
	}

	gamedata.UpdateEffects(&e.Entity.Effects, deltaTime, e.TakeDamage)

	playerCenterX := playerX
	playerCenterY := playerY
	enemyCenterX := e.PosX + e.Hitbox.Width/2
	enemyCenterY := e.PosY + e.Hitbox.Height/2

	dx := playerCenterX - enemyCenterX
	dy := playerCenterY - enemyCenterY
	distance := dx*dx + dy*dy

	if distance <= e.AggroRange*e.AggroRange {
		if e.State == EnemyStateIdle {
			e.State = EnemyStateChasing
		}
	} else {
		e.State = EnemyStateIdle
	}

	if distance <= e.AttackRange*e.AttackRange {
		e.State = EnemyStateAttacking
	} else if e.State == EnemyStateChasing {
		distanceSqrt := float32(math.Sqrt(float64(distance)))
		if distanceSqrt > 0 {
			moveX := (dx / distanceSqrt) * e.MoveSpeed * deltaTime
			moveY := (dy / distanceSqrt) * e.MoveSpeed * deltaTime
			e.PosX += moveX
			e.PosY += moveY

			if moveX > 0 {
				e.FacingRight = true
			} else if moveX < 0 {
				e.FacingRight = false
			}
		}
	} else if e.State == EnemyStateAttacking {
		e.State = EnemyStateChasing
	}
}

func (e *Enemy) Attack(player *Player) bool {
	if e.CurrentCooldown <= 0 && e.State == EnemyStateAttacking {
		player.TakeDamage(e.Damage)
		e.AttackFlashTimer = 0.15
		e.CurrentCooldown = e.AttackCooldown
		return true
	}
	return false
}

func (e *Enemy) TakeDamage(damage int) {
	e.Entity.ApplyDamage(damage)
	e.HitFlashTimer = 0.2
}
