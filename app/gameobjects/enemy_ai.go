package gameobjects

import (
	"math"

	"singlefantasy/app/gamedata"
)

func ResolveEnemyIntent(enemy *Enemy, playerX, playerY float32) {
	if enemy == nil || !enemy.IsAlive() {
		return
	}

	enemy.IntentMoveX = 0
	enemy.IntentMoveY = 0
	enemy.WantsAttack = false

	enemyX, enemyY := enemy.Center()
	dx := playerX - enemyX
	dy := playerY - enemyY
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if distance > enemy.AggroRange {
		enemy.State = EnemyStateIdle
		return
	}

	if enemy.AttackMode == gamedata.EnemyAttackProjectile || enemy.AttackMode == gamedata.EnemyAttackCasterAOE {
		resolveRangedOrCasterIntent(enemy, dx, dy, distance)
		return
	}
	resolveMeleeIntent(enemy, dx, dy, distance)
}

func resolveMeleeIntent(enemy *Enemy, dx, dy, distance float32) {
	if distance <= enemy.AttackRange {
		enemy.State = EnemyStateAttacking
		enemy.WantsAttack = true
		return
	}

	enemy.State = EnemyStateChasing
	setEnemyMoveIntent(enemy, dx, dy, distance)
}

func resolveRangedOrCasterIntent(enemy *Enemy, dx, dy, distance float32) {
	preferred := enemy.PreferredRange
	if preferred <= 0 {
		preferred = enemy.AttackRange * 0.75
	}

	retreat := enemy.RetreatRange
	if retreat <= 0 {
		retreat = preferred * 0.6
	}

	if distance < retreat {
		enemy.State = EnemyStateRetreating
		setEnemyMoveIntent(enemy, -dx, -dy, distance)
		return
	}

	if distance > preferred {
		enemy.State = EnemyStateChasing
		setEnemyMoveIntent(enemy, dx, dy, distance)
		return
	}

	if distance <= enemy.AttackRange {
		enemy.State = EnemyStateAttacking
		enemy.WantsAttack = true
		return
	}

	enemy.State = EnemyStateChasing
	setEnemyMoveIntent(enemy, dx, dy, distance)
}

func setEnemyMoveIntent(enemy *Enemy, dx, dy, distance float32) {
	if distance <= 0.0001 {
		enemy.IntentMoveX = 0
		enemy.IntentMoveY = 0
		return
	}

	enemy.IntentMoveX = dx / distance
	enemy.IntentMoveY = dy / distance
	if enemy.IntentMoveX > 0 {
		enemy.FacingRight = true
	}
	if enemy.IntentMoveX < 0 {
		enemy.FacingRight = false
	}
}
