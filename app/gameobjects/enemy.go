package gameobjects

import (
	"math"
	"singlefantasy/app/gamedata"
)

type EnemyState int

const (
	EnemyStateIdle EnemyState = iota
	EnemyStateChasing
	EnemyStateAttacking
)

type Enemy struct {
	X                float32
	Y                float32
	Width            float32
	Height           float32
	Health           int
	MaxHealth        int
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
	Alive            bool
	IsElite          bool
	Effects          []gamedata.EffectInstance
}

func NewEnemy(x, y float32, isElite bool) *Enemy {
	health := 50
	damage := 5
	moveSpeed := float32(100)

	if isElite {
		health = 100
		damage = 10
		moveSpeed = 120
	}

	return &Enemy{
		X:                x,
		Y:                y,
		Width:            30,
		Height:           30,
		Health:           health,
		MaxHealth:        health,
		Damage:           damage,
		MoveSpeed:        moveSpeed,
		AttackCooldown:   1.0,
		CurrentCooldown:  0,
		AttackRange:      60,
		AggroRange:       150,
		HitFlashTimer:    0,
		AttackFlashTimer: 0,
		FacingRight:      true,
		State:            EnemyStateIdle,
		Alive:            true,
		IsElite:          isElite,
	}
}

func (e *Enemy) Update(deltaTime float32, playerX, playerY float32) {
	if !e.Alive {
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

	gamedata.UpdateEffects(&e.Effects, deltaTime, e.TakeDamage)

	playerCenterX := playerX
	playerCenterY := playerY
	enemyCenterX := e.X + e.Width/2
	enemyCenterY := e.Y + e.Height/2

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
			e.X += moveX
			e.Y += moveY

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
	e.Health -= damage
	if e.Health < 0 {
		e.Health = 0
	}
	if e.Health <= 0 {
		e.Alive = false
	}
	e.HitFlashTimer = 0.2
}