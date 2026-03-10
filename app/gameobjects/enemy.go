package gameobjects

import (
	"singlefantasy/app/core"
	"singlefantasy/app/gamedata"
)

type EnemyState int

const (
	EnemyStateIdle EnemyState = iota
	EnemyStateChasing
	EnemyStateRetreating
	EnemyStateAttacking
)

type EnemyAttackPayload struct {
	Damage             int
	DamageType         gamedata.DamageType
	SourceX            float32
	SourceY            float32
	TargetX            float32
	TargetY            float32
	AttackMode         gamedata.EnemyAttackMode
	ProjectileSpeed    float32
	ProjectileRadius   float32
	ProjectileLifetime float32
	OnHitEffects       []gamedata.EffectSpec
}

type Enemy struct {
	core.Entity
	Name               string
	Role               string
	Archetype          gamedata.EnemyArchetypeType
	AttackMode         gamedata.EnemyAttackMode
	DamageType         gamedata.DamageType
	Damage             int
	MoveSpeed          float32
	AttackCooldown     float32
	CurrentCooldown    float32
	AttackRange        float32
	AggroRange         float32
	PreferredRange     float32
	RetreatRange       float32
	ProjectileSpeed    float32
	ProjectileRadius   float32
	ProjectileLifetime float32
	OnHitEffects       []gamedata.EffectSpec
	XPReward           int
	ThreatValue        int
	HitFlashTimer      float32
	AttackFlashTimer   float32
	FacingRight        bool
	State              EnemyState
	IsElite            bool
	EliteModifierType  gamedata.EliteModifierType
	EliteModifierName  string
	IntentMoveX        float32
	IntentMoveY        float32
	WantsAttack        bool
}

func NewEnemy(x, y float32, isElite bool) *Enemy {
	return NewEnemyFromArchetype(x, y, gamedata.EnemyArchetypeRaider, isElite, gamedata.EliteModifierScorching)
}

func NewEnemyFromArchetype(x, y float32, archetypeType gamedata.EnemyArchetypeType, isElite bool, eliteModifierType gamedata.EliteModifierType) *Enemy {
	archetype := gamedata.GetEnemyArchetype(archetypeType)
	maxHP := archetype.MaxHP
	damage := archetype.Damage
	modifierName := ""
	combinedEffects := make([]gamedata.EffectSpec, 0, len(archetype.OnHitEffects)+2)
	if len(archetype.OnHitEffects) > 0 {
		combinedEffects = append(combinedEffects, archetype.OnHitEffects...)
	}

	if isElite {
		modifier := gamedata.GetEliteModifier(eliteModifierType)
		if modifier.HPMultiplier > 0 {
			maxHP = int(float32(maxHP) * modifier.HPMultiplier)
		}
		if modifier.DmgMultiplier > 0 {
			damage = int(float32(damage) * modifier.DmgMultiplier)
		}
		modifierName = modifier.Name
		combinedEffects = append(combinedEffects, modifier.OnHitEffects...)
	}

	if maxHP <= 1 {
		maxHP = 1
	}
	if damage <= 1 {
		damage = 1
	}
	return &Enemy{
		Entity: core.Entity{
			PosX:    x,
			PosY:    y,
			HP:      maxHP,
			MaxHP:   maxHP,
			Stats:   nil,
			Hitbox:  core.Hitbox{Width: archetype.Width, Height: archetype.Height},
			Faction: core.FactionEnemy,
			Alive:   true,
		},
		Name:               archetype.Name,
		Role:               archetype.Role,
		Archetype:          archetype.Type,
		AttackMode:         archetype.AttackMode,
		DamageType:         archetype.DamageType,
		Damage:             damage,
		MoveSpeed:          archetype.MoveSpeed,
		AttackCooldown:     archetype.AttackCooldown,
		CurrentCooldown:    0,
		AttackRange:        archetype.AttackRange,
		AggroRange:         archetype.AggroRange,
		PreferredRange:     archetype.PreferredRange,
		RetreatRange:       archetype.RetreatRange,
		ProjectileSpeed:    archetype.ProjectileSpeed,
		ProjectileRadius:   archetype.ProjectileRadius,
		ProjectileLifetime: archetype.ProjectileLifetime,
		OnHitEffects:       combinedEffects,
		XPReward:           archetype.XPReward,
		ThreatValue:        archetype.ThreatValue,
		HitFlashTimer:      0,
		AttackFlashTimer:   0,
		FacingRight:        true,
		State:              EnemyStateIdle,
		IsElite:            isElite,
		EliteModifierType:  eliteModifierType,
		EliteModifierName:  modifierName,
		IntentMoveX:        0,
		IntentMoveY:        0,
		WantsAttack:        false,
	}
}

func (e *Enemy) Update(deltaTime float32) {
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
	e.IntentMoveX = 0
	e.IntentMoveY = 0
	e.WantsAttack = false
	if !gamedata.CanAct(&e.Entity.Effects) {
		e.State = EnemyStateIdle
		return
	}
}

func (e *Enemy) Attack(playerX, playerY float32) (bool, EnemyAttackPayload) {
	zero := EnemyAttackPayload{}
	if !gamedata.CanAct(&e.Entity.Effects) {
		return false, zero
	}
	if e.CurrentCooldown > 0 || e.State != EnemyStateAttacking || !e.WantsAttack {
		return false, zero
	}
	e.AttackFlashTimer = 0.15
	e.CurrentCooldown = e.AttackCooldown
	sourceX, sourceY := e.Center()
	onHit := make([]gamedata.EffectSpec, len(e.OnHitEffects))
	copy(onHit, e.OnHitEffects)
	return true, EnemyAttackPayload{
		Damage:             e.Damage,
		DamageType:         e.DamageType,
		SourceX:            sourceX,
		SourceY:            sourceY,
		TargetX:            playerX,
		TargetY:            playerY,
		AttackMode:         e.AttackMode,
		ProjectileSpeed:    e.ProjectileSpeed,
		ProjectileRadius:   e.ProjectileRadius,
		ProjectileLifetime: e.ProjectileLifetime,
		OnHitEffects:       onHit,
	}
}

func (e *Enemy) TakeDamage(damage int) {
	e.Entity.ApplyDamage(damage)
	e.HitFlashTimer = 0.2
}

func (e *Enemy) DisplayName() string {
	if e == nil {
		return "Enemy"
	}
	if e.IsElite && e.EliteModifierName != "" {
		return "Elite " + e.EliteModifierName + " " + e.Name
	}
	if e.Name != "" {
		return e.Name
	}
	if e.IsElite {
		return "Elite Enemy"
	}
	return "Enemy"
}
