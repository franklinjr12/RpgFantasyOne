package gameobjects

import (
	"math"

	"singlefantasy/app/gamedata"
)

type BossPhase int

const (
	BossPhaseNormal BossPhase = iota
	BossPhaseEnraged
)

func (p BossPhase) String() string {
	switch p {
	case BossPhaseEnraged:
		return "enraged"
	default:
		return "normal"
	}
}

type BossHeavyAttackState int

const (
	BossHeavyAttackIdle BossHeavyAttackState = iota
	BossHeavyAttackTelegraph
	BossHeavyAttackCooldown
)

func (s BossHeavyAttackState) String() string {
	switch s {
	case BossHeavyAttackTelegraph:
		return "telegraph"
	case BossHeavyAttackCooldown:
		return "cooldown"
	default:
		return "idle"
	}
}

type BossDamageEventType int

const (
	BossDamageEventHeavy BossDamageEventType = iota
	BossDamageEventArea
)

type BossDamageEvent struct {
	Type       BossDamageEventType
	X          float32
	Y          float32
	Radius     float32
	Damage     int
	DamageType gamedata.DamageType
	Effects    []gamedata.EffectSpec
}

type BossTelegraph struct {
	X        float32
	Y        float32
	Radius   float32
	Duration float32
	TimeLeft float32
}

type BossAreaDenialZone struct {
	X               float32
	Y               float32
	Radius          float32
	WarningDuration float32
	WarningTimeLeft float32
	ActiveDuration  float32
	ActiveTimeLeft  float32
	TickRate        float32
	TickTimer       float32
	Damage          int
	DamageType      gamedata.DamageType
	Effects         []gamedata.EffectSpec
	Active          bool
}

type Boss struct {
	*Enemy
	Config                 gamedata.BossEncounterConfig
	Phase                  BossPhase
	HeavyState             BossHeavyAttackState
	HeavyTelegraph         BossTelegraph
	HeavyCooldownRemaining float32
	AreaCooldownRemaining  float32
	AreaZones              []*BossAreaDenialZone
	pendingDamageEvents    []BossDamageEvent
	EnrageTriggered        bool
	BaseDamage             int
	BaseMoveSpeed          float32
	Projectiles            []*BossProjectile
}

type BossProjectile struct {
	X      float32
	Y      float32
	VX     float32
	VY     float32
	Speed  float32
	Damage int
	Radius float32
	Alive  bool
}

func NewBoss(x, y float32, biome string) *Boss {
	cfg := gamedata.GetBossEncounterConfig(biome)
	enemy := NewEnemy(x, y, false)
	enemy.Name = "Dungeon Boss"
	enemy.Role = "Boss"
	enemy.Archetype = gamedata.EnemyArchetypeBrute
	enemy.AttackMode = gamedata.EnemyAttackMelee
	enemy.DamageType = gamedata.DamagePhysical
	enemy.HP = cfg.MaxHP
	enemy.MaxHP = cfg.MaxHP
	enemy.Damage = cfg.Damage
	enemy.MoveSpeed = cfg.MoveSpeed
	enemy.AttackRange = cfg.AttackRange
	enemy.AggroRange = cfg.AggroRange
	enemy.AttackCooldown = cfg.AttackCooldown
	enemy.Hitbox.Width = cfg.Width
	enemy.Hitbox.Height = cfg.Height
	enemy.PreferredRange = cfg.AttackRange * 0.8
	enemy.RetreatRange = 0
	enemy.XPReward = 120
	enemy.Stats = nil

	boss := &Boss{
		Enemy:                  enemy,
		Config:                 cfg,
		Phase:                  BossPhaseNormal,
		HeavyState:             BossHeavyAttackIdle,
		HeavyTelegraph:         BossTelegraph{},
		HeavyCooldownRemaining: cfg.HeavyAttack.Cooldown * 0.55,
		AreaCooldownRemaining:  cfg.AreaDenial.Cooldown * 0.55,
		AreaZones:              []*BossAreaDenialZone{},
		pendingDamageEvents:    []BossDamageEvent{},
		EnrageTriggered:        false,
		BaseDamage:             cfg.Damage,
		BaseMoveSpeed:          cfg.MoveSpeed,
		Projectiles:            []*BossProjectile{},
	}
	if boss.HeavyCooldownRemaining < 0 {
		boss.HeavyCooldownRemaining = 0
	}
	if boss.AreaCooldownRemaining < 0 {
		boss.AreaCooldownRemaining = 0
	}
	return boss
}

func (b *Boss) Update(deltaTime float32, playerX, playerY float32) {
	if b == nil || !b.Entity.IsAlive() {
		return
	}

	if len(b.pendingDamageEvents) > 0 {
		b.pendingDamageEvents = b.pendingDamageEvents[:0]
	}

	b.Enemy.Update(deltaTime)
	ResolveEnemyIntent(b.Enemy, playerX, playerY)
	b.updateEnrageState()
	b.updateHeavyAttack(deltaTime, playerX, playerY)
	b.updateAreaDenial(deltaTime, playerX, playerY)
	b.updateAreaZones(deltaTime)
}

func (b *Boss) updateEnrageState() {
	if b.EnrageTriggered {
		return
	}
	threshold := int(float32(b.MaxHP) * b.Config.Enrage.ThresholdHPPercent)
	if threshold < 1 {
		threshold = 1
	}
	if b.HP > threshold {
		return
	}

	b.EnrageTriggered = true
	b.Phase = BossPhaseEnraged

	damage := int(float32(b.BaseDamage) * b.Config.Enrage.DamageMultiplier)
	if damage < 1 {
		damage = b.BaseDamage
	}
	b.Damage = damage

	moveSpeed := b.BaseMoveSpeed * b.Config.Enrage.MoveSpeedMultiplier
	if moveSpeed <= 0 {
		moveSpeed = b.BaseMoveSpeed
	}
	b.MoveSpeed = moveSpeed

	if b.HeavyState == BossHeavyAttackCooldown {
		maxHeavyCooldown := b.currentHeavyCooldown()
		if b.HeavyCooldownRemaining > maxHeavyCooldown {
			b.HeavyCooldownRemaining = maxHeavyCooldown
		}
	}

	maxAreaCooldown := b.currentAreaCooldown()
	if b.AreaCooldownRemaining > maxAreaCooldown {
		b.AreaCooldownRemaining = maxAreaCooldown
	}
}

func (b *Boss) updateHeavyAttack(deltaTime float32, playerX, playerY float32) {
	switch b.HeavyState {
	case BossHeavyAttackTelegraph:
		b.HeavyTelegraph.TimeLeft -= deltaTime
		if b.HeavyTelegraph.TimeLeft <= 0 {
			b.HeavyTelegraph.TimeLeft = 0
			b.pendingDamageEvents = append(b.pendingDamageEvents, BossDamageEvent{
				Type:       BossDamageEventHeavy,
				X:          b.HeavyTelegraph.X,
				Y:          b.HeavyTelegraph.Y,
				Radius:     b.HeavyTelegraph.Radius,
				Damage:     b.Config.HeavyAttack.Damage,
				DamageType: b.Config.HeavyAttack.DamageType,
			})
			b.HeavyState = BossHeavyAttackCooldown
			b.HeavyCooldownRemaining = b.currentHeavyCooldown()
		}
		return
	case BossHeavyAttackCooldown:
		b.HeavyCooldownRemaining -= deltaTime
		if b.HeavyCooldownRemaining <= 0 {
			b.HeavyCooldownRemaining = 0
			b.HeavyState = BossHeavyAttackIdle
		}
	}

	if b.HeavyState != BossHeavyAttackIdle || b.HeavyCooldownRemaining > 0 {
		return
	}
	if !gamedata.CanAct(&b.Effects) {
		return
	}

	centerX, centerY := b.Center()
	dx := playerX - centerX
	dy := playerY - centerY
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if distance > b.AggroRange {
		return
	}

	duration := b.Config.HeavyAttack.TelegraphDuration
	if duration <= 0 {
		duration = 0.2
	}
	radius := b.Config.HeavyAttack.Radius
	if radius <= 0 {
		radius = 32
	}

	b.HeavyState = BossHeavyAttackTelegraph
	b.HeavyTelegraph = BossTelegraph{
		X:        playerX,
		Y:        playerY,
		Radius:   radius,
		Duration: duration,
		TimeLeft: duration,
	}
	b.AttackFlashTimer = 0.22
}

func (b *Boss) updateAreaDenial(deltaTime float32, playerX, playerY float32) {
	if b.AreaCooldownRemaining > 0 {
		b.AreaCooldownRemaining -= deltaTime
	}
	if b.AreaCooldownRemaining > 0 {
		return
	}
	if !gamedata.CanAct(&b.Effects) {
		return
	}

	centerX, centerY := b.Center()
	dx := playerX - centerX
	dy := playerY - centerY
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if distance > b.AggroRange {
		return
	}

	count := b.Config.AreaDenial.ZoneCount
	if b.EnrageTriggered {
		count += b.Config.Enrage.ZoneCountBonus
	}
	if count < 1 {
		count = 1
	}
	spawnDistance := b.Config.AreaDenial.SpawnDistance
	if spawnDistance < 0 {
		spawnDistance = 0
	}

	for i := 0; i < count; i++ {
		zoneX := playerX
		zoneY := playerY
		if i > 0 && spawnDistance > 0 {
			angle := float32(i) * 2 * math.Pi / float32(count)
			zoneX = playerX + float32(math.Cos(float64(angle)))*spawnDistance
			zoneY = playerY + float32(math.Sin(float64(angle)))*spawnDistance
		}
		b.spawnAreaZone(zoneX, zoneY)
	}

	b.AreaCooldownRemaining = b.currentAreaCooldown()
}

func (b *Boss) spawnAreaZone(x, y float32) {
	warning := b.Config.AreaDenial.WarningDuration
	if warning <= 0 {
		warning = 0.2
	}
	active := b.Config.AreaDenial.ActiveDuration
	if active <= 0 {
		active = 0.5
	}
	tickRate := b.Config.AreaDenial.TickRate
	if tickRate <= 0 {
		tickRate = 0.5
	}
	radius := b.Config.AreaDenial.Radius
	if radius <= 0 {
		radius = 48
	}
	damage := b.Config.AreaDenial.Damage
	if damage <= 0 {
		damage = 1
	}

	effects := make([]gamedata.EffectSpec, len(b.Config.AreaDenial.Effects))
	copy(effects, b.Config.AreaDenial.Effects)

	b.AreaZones = append(b.AreaZones, &BossAreaDenialZone{
		X:               x,
		Y:               y,
		Radius:          radius,
		WarningDuration: warning,
		WarningTimeLeft: warning,
		ActiveDuration:  active,
		ActiveTimeLeft:  active,
		TickRate:        tickRate,
		TickTimer:       0,
		Damage:          damage,
		DamageType:      b.Config.AreaDenial.DamageType,
		Effects:         effects,
		Active:          false,
	})
}

func (b *Boss) updateAreaZones(deltaTime float32) {
	for i := len(b.AreaZones) - 1; i >= 0; i-- {
		zone := b.AreaZones[i]
		if zone == nil {
			b.AreaZones = append(b.AreaZones[:i], b.AreaZones[i+1:]...)
			continue
		}

		if !zone.Active {
			zone.WarningTimeLeft -= deltaTime
			if zone.WarningTimeLeft <= 0 {
				zone.WarningTimeLeft = 0
				zone.Active = true
				zone.TickTimer = 0
			}
			continue
		}

		zone.ActiveTimeLeft -= deltaTime
		zone.TickTimer += deltaTime
		for zone.TickTimer >= zone.TickRate && zone.ActiveTimeLeft > 0 {
			zone.TickTimer -= zone.TickRate
			effects := make([]gamedata.EffectSpec, len(zone.Effects))
			copy(effects, zone.Effects)
			b.pendingDamageEvents = append(b.pendingDamageEvents, BossDamageEvent{
				Type:       BossDamageEventArea,
				X:          zone.X,
				Y:          zone.Y,
				Radius:     zone.Radius,
				Damage:     zone.Damage,
				DamageType: zone.DamageType,
				Effects:    effects,
			})
		}

		if zone.ActiveTimeLeft <= 0 {
			b.AreaZones = append(b.AreaZones[:i], b.AreaZones[i+1:]...)
		}
	}
}

func (b *Boss) currentHeavyCooldown() float32 {
	cooldown := b.Config.HeavyAttack.Cooldown
	if b.EnrageTriggered {
		cooldown *= b.Config.Enrage.HeavyCooldownMultiplier
	}
	if cooldown < 0.2 {
		cooldown = 0.2
	}
	return cooldown
}

func (b *Boss) currentAreaCooldown() float32 {
	cooldown := b.Config.AreaDenial.Cooldown
	if b.EnrageTriggered {
		cooldown *= b.Config.Enrage.AreaCooldownMultiplier
	}
	if cooldown < 0.4 {
		cooldown = 0.4
	}
	return cooldown
}

func (b *Boss) Attack(playerX, playerY float32) (bool, int, float32, float32) {
	hit, payload := b.Enemy.Attack(playerX, playerY)
	if !hit {
		return false, 0, 0, 0
	}
	return true, payload.Damage, payload.SourceX, payload.SourceY
}

func (b *Boss) ConsumeDamageEvents() []BossDamageEvent {
	if len(b.pendingDamageEvents) == 0 {
		return nil
	}
	out := make([]BossDamageEvent, len(b.pendingDamageEvents))
	copy(out, b.pendingDamageEvents)
	b.pendingDamageEvents = b.pendingDamageEvents[:0]
	return out
}

func (b *Boss) ActiveHeavyTelegraph() (BossTelegraph, bool) {
	if b.HeavyState != BossHeavyAttackTelegraph || b.HeavyTelegraph.TimeLeft <= 0 {
		return BossTelegraph{}, false
	}
	return b.HeavyTelegraph, true
}

func (b *Boss) ActiveAreaZones() []BossAreaDenialZone {
	if len(b.AreaZones) == 0 {
		return nil
	}
	out := make([]BossAreaDenialZone, 0, len(b.AreaZones))
	for _, zone := range b.AreaZones {
		if zone == nil {
			continue
		}
		copied := *zone
		copied.Effects = append([]gamedata.EffectSpec(nil), zone.Effects...)
		out = append(out, copied)
	}
	return out
}

func (b *Boss) ActiveZoneCount() int {
	return len(b.AreaZones)
}

func (b *Boss) HeavyTimeRemaining() float32 {
	if b.HeavyState == BossHeavyAttackTelegraph {
		return b.HeavyTelegraph.TimeLeft
	}
	if b.HeavyState == BossHeavyAttackCooldown {
		return b.HeavyCooldownRemaining
	}
	return 0
}
