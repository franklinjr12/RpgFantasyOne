package gameobjects

import "math"

type BossPhase int

const (
	BossPhase1 BossPhase = iota
	BossPhase2
	BossPhase3
)

type Boss struct {
	*Enemy
	Phase            BossPhase
	TelegraphTimer   float32
	AddSpawnTimer    float32
	AddSpawnCooldown float32
	Projectiles      []*BossProjectile
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

func NewBoss(x, y float32) *Boss {
	enemy := NewEnemy(x, y, false)
	enemy.Health = 500
	enemy.MaxHealth = 500
	enemy.Damage = 15
	enemy.MoveSpeed = 80
	enemy.AttackRange = 80
	enemy.AggroRange = 1000
	enemy.Width = 60
	enemy.Height = 60

	return &Boss{
		Enemy:            enemy,
		Phase:            BossPhase1,
		TelegraphTimer:   0,
		AddSpawnTimer:    0,
		AddSpawnCooldown: 5.0,
		Projectiles:      []*BossProjectile{},
	}
}

func (b *Boss) Update(deltaTime float32, playerX, playerY float32) {
	if !b.Alive {
		return
	}

	healthPercent := float32(b.Health) / float32(b.MaxHealth)
	if healthPercent < 0.33 {
		b.Phase = BossPhase3
	} else if healthPercent < 0.66 {
		b.Phase = BossPhase2
	} else {
		b.Phase = BossPhase1
	}

	b.Enemy.Update(deltaTime, playerX, playerY)

	b.TelegraphTimer -= deltaTime
	b.AddSpawnTimer -= deltaTime

	if b.AddSpawnTimer <= 0 && b.Phase >= BossPhase2 {
		b.AddSpawnTimer = b.AddSpawnCooldown
	}

	for i := len(b.Projectiles) - 1; i >= 0; i-- {
		proj := b.Projectiles[i]
		if !proj.Alive {
			b.Projectiles = append(b.Projectiles[:i], b.Projectiles[i+1:]...)
			continue
		}

		proj.X += proj.VX * deltaTime
		proj.Y += proj.VY * deltaTime
	}

	if b.State == EnemyStateAttacking && b.CurrentCooldown <= 0 {
		if b.Phase >= BossPhase2 {
			bossCenterX := b.X + b.Width/2
			bossCenterY := b.Y + b.Height/2
			dx := playerX - bossCenterX
			dy := playerY - bossCenterY
			distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if distance > 0 {
				speed := float32(200)
				proj := &BossProjectile{
					X:      bossCenterX,
					Y:      bossCenterY,
					VX:     (dx / distance) * speed,
					VY:     (dy / distance) * speed,
					Speed:  speed,
					Damage: 10,
					Radius: 8,
					Alive:  true,
				}
				b.Projectiles = append(b.Projectiles, proj)
			}
		}
	}
}

func (b *Boss) ShouldSpawnAdds() bool {
	return b.AddSpawnTimer <= 0 && b.Phase >= BossPhase2
}

func (b *Boss) ResetAddSpawnTimer() {
	b.AddSpawnTimer = b.AddSpawnCooldown
}
