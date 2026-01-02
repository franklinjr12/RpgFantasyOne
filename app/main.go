package main

import (
	"math"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type EnemyState int

const (
	EnemyStateIdle EnemyState = iota
	EnemyStateChasing
	EnemyStateAttacking
)

type Player struct {
	X            float32
	Y            float32
	Width        float32
	Height       float32
	Health       int
	MaxHealth    int
	MoveSpeed    float32
	AttackDamage int
	AttackRange  float32
}

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
}

func main() {
	rl.InitWindow(800, 450, "raylib [core] example - basic window")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	playerTexture := rl.LoadTexture("resources/sprites/soldier_character.png")
	defer rl.UnloadTexture(playerTexture)

	enemyTexture := rl.LoadTexture("resources/sprites/orc_chacacter.png")
	defer rl.UnloadTexture(enemyTexture)

	windowWidth := float32(800)
	windowHeight := float32(450)

	player := Player{
		X:            windowWidth / 2,
		Y:            windowHeight / 2,
		Width:        40,
		Height:       40,
		Health:       100,
		MaxHealth:    100,
		MoveSpeed:    200,
		AttackDamage: 10,
		AttackRange:  50,
	}

	numEnemies := 5
	enemies := make([]Enemy, numEnemies)
	for i := 0; i < numEnemies; i++ {
		var x, y float32
		minDistance := float32(100)
		for {
			x = float32(rand.Intn(int(windowWidth - 40)))
			y = float32(rand.Intn(int(windowHeight - 40)))
			dx := x - player.X
			dy := y - player.Y
			distance := dx*dx + dy*dy
			if distance >= minDistance*minDistance {
				break
			}
		}
		enemies[i] = Enemy{
			X:                x,
			Y:                y,
			Width:            30,
			Height:           30,
			Health:           50,
			MaxHealth:        50,
			Damage:           5,
			MoveSpeed:        100,
			AttackCooldown:   1.0,
			CurrentCooldown:  0,
			AttackRange:      60,
			AggroRange:       150,
			HitFlashTimer:    0,
			AttackFlashTimer: 0,
			FacingRight:      true,
			State:            EnemyStateIdle,
			Alive:            true,
		}
	}

	hitFlashTimer := float32(0)

	for !rl.WindowShouldClose() {
		deltaTime := rl.GetFrameTime()

		if hitFlashTimer > 0 {
			hitFlashTimer -= deltaTime
			if hitFlashTimer < 0 {
				hitFlashTimer = 0
			}
		}

		if rl.IsKeyDown(rl.KeyW) || rl.IsKeyDown(rl.KeyUp) {
			player.Y -= player.MoveSpeed * deltaTime
		}
		if rl.IsKeyDown(rl.KeyS) || rl.IsKeyDown(rl.KeyDown) {
			player.Y += player.MoveSpeed * deltaTime
		}
		if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
			player.X -= player.MoveSpeed * deltaTime
		}
		if rl.IsKeyDown(rl.KeyD) || rl.IsKeyDown(rl.KeyRight) {
			player.X += player.MoveSpeed * deltaTime
		}

		if player.X < 0 {
			player.X = 0
		}
		if player.X+player.Width > windowWidth {
			player.X = windowWidth - player.Width
		}
		if player.Y < 0 {
			player.Y = 0
		}
		if player.Y+player.Height > windowHeight {
			player.Y = windowHeight - player.Height
		}

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			mousePos := rl.GetMousePosition()

			for i := range enemies {
				if !enemies[i].Alive {
					continue
				}

				if mousePos.X >= enemies[i].X && mousePos.X <= enemies[i].X+enemies[i].Width &&
					mousePos.Y >= enemies[i].Y && mousePos.Y <= enemies[i].Y+enemies[i].Height {

					playerCenterX := player.X + player.Width/2
					playerCenterY := player.Y + player.Height/2
					enemyCenterX := enemies[i].X + enemies[i].Width/2
					enemyCenterY := enemies[i].Y + enemies[i].Height/2

					dx := enemyCenterX - playerCenterX
					dy := enemyCenterY - playerCenterY
					distance := dx*dx + dy*dy

					if distance <= player.AttackRange*player.AttackRange {
						enemies[i].Health -= player.AttackDamage
						if enemies[i].Health < 0 {
							enemies[i].Health = 0
						}
						if enemies[i].Health <= 0 {
							enemies[i].Alive = false
						}
						enemies[i].HitFlashTimer = 0.2
					}
					break
				}
			}
		}

		for i := range enemies {
			if !enemies[i].Alive {
				continue
			}

			playerCenterX := player.X + player.Width/2
			playerCenterY := player.Y + player.Height/2
			enemyCenterX := enemies[i].X + enemies[i].Width/2
			enemyCenterY := enemies[i].Y + enemies[i].Height/2

			dx := playerCenterX - enemyCenterX
			dy := playerCenterY - enemyCenterY
			distance := dx*dx + dy*dy

			if distance <= enemies[i].AggroRange*enemies[i].AggroRange {
				if enemies[i].State == EnemyStateIdle {
					enemies[i].State = EnemyStateChasing
				}
			} else {
				enemies[i].State = EnemyStateIdle
			}

			if distance <= enemies[i].AttackRange*enemies[i].AttackRange {
				enemies[i].State = EnemyStateAttacking
				if enemies[i].CurrentCooldown <= 0 {
					player.Health -= enemies[i].Damage
					if player.Health < 0 {
						player.Health = 0
					}
					hitFlashTimer = 0.2
					enemies[i].AttackFlashTimer = 0.15
					enemies[i].CurrentCooldown = enemies[i].AttackCooldown
				}
			} else if enemies[i].State == EnemyStateChasing {
				distanceSqrt := float32(math.Sqrt(float64(distance)))
				if distanceSqrt > 0 {
					moveX := (dx / distanceSqrt) * enemies[i].MoveSpeed * deltaTime
					moveY := (dy / distanceSqrt) * enemies[i].MoveSpeed * deltaTime
					enemies[i].X += moveX
					enemies[i].Y += moveY

					if moveX > 0 {
						enemies[i].FacingRight = true
					} else if moveX < 0 {
						enemies[i].FacingRight = false
					}
				}
			} else if enemies[i].State == EnemyStateAttacking {
				enemies[i].State = EnemyStateChasing
			}

			if enemies[i].CurrentCooldown > 0 {
				enemies[i].CurrentCooldown -= deltaTime
				if enemies[i].CurrentCooldown < 0 {
					enemies[i].CurrentCooldown = 0
				}
			}

			if enemies[i].HitFlashTimer > 0 {
				enemies[i].HitFlashTimer -= deltaTime
				if enemies[i].HitFlashTimer < 0 {
					enemies[i].HitFlashTimer = 0
				}
			}

			if enemies[i].AttackFlashTimer > 0 {
				enemies[i].AttackFlashTimer -= deltaTime
				if enemies[i].AttackFlashTimer < 0 {
					enemies[i].AttackFlashTimer = 0
				}
			}
		}

		if player.Health <= 0 {
			player.X = windowWidth / 2
			player.Y = windowHeight / 2
			player.Health = player.MaxHealth

			for i := range enemies {
				if enemies[i].Alive {
					var x, y float32
					minDistance := float32(100)
					for {
						x = float32(rand.Intn(int(windowWidth - 40)))
						y = float32(rand.Intn(int(windowHeight - 40)))
						dx := x - player.X
						dy := y - player.Y
						distance := dx*dx + dy*dy
						if distance >= minDistance*minDistance {
							break
						}
					}
					enemies[i].X = x
					enemies[i].Y = y
					enemies[i].State = EnemyStateIdle
					enemies[i].CurrentCooldown = 0
				}
			}
		}

		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		spriteWidth := float32(playerTexture.Width)
		spriteHeight := float32(playerTexture.Height)
		spriteX := player.X + (player.Width-spriteWidth)/2
		spriteY := player.Y + (player.Height-spriteHeight)/2

		tint := rl.White
		if hitFlashTimer > 0 {
			tint = rl.Red
		}
		rl.DrawTexture(playerTexture, int32(spriteX), int32(spriteY), tint)

		healthBarWidth := player.Width
		healthBarHeight := float32(5)
		healthBarX := player.X
		healthBarY := player.Y - healthBarHeight - 3
		healthPercent := float32(player.Health) / float32(player.MaxHealth)
		if healthPercent < 0 {
			healthPercent = 0
		}
		if healthPercent > 1 {
			healthPercent = 1
		}

		rl.DrawRectangleRec(rl.NewRectangle(healthBarX, healthBarY, healthBarWidth, healthBarHeight), rl.Red)
		rl.DrawRectangleRec(rl.NewRectangle(healthBarX, healthBarY, healthBarWidth*healthPercent, healthBarHeight), rl.Green)

		for i := range enemies {
			if enemies[i].Alive {
				enemySpriteWidth := float32(enemyTexture.Width)
				enemySpriteHeight := float32(enemyTexture.Height)
				enemySpriteX := enemies[i].X + (enemies[i].Width-enemySpriteWidth)/2
				enemySpriteY := enemies[i].Y + (enemies[i].Height-enemySpriteHeight)/2

				enemyTint := rl.White
				if enemies[i].HitFlashTimer > 0 {
					enemyTint = rl.Orange
				} else if enemies[i].AttackFlashTimer > 0 {
					enemyTint = rl.Yellow
				}

				sourceRect := rl.NewRectangle(0, 0, float32(enemyTexture.Width), float32(enemyTexture.Height))
				destRect := rl.NewRectangle(enemySpriteX, enemySpriteY, float32(enemyTexture.Width), float32(enemyTexture.Height))
				origin := rl.NewVector2(0, 0)

				if enemies[i].FacingRight {
					rl.DrawTexturePro(enemyTexture, sourceRect, destRect, origin, 0, enemyTint)
				} else {
					destRectFlipped := rl.NewRectangle(enemySpriteX+float32(enemyTexture.Width), enemySpriteY, -float32(enemyTexture.Width), float32(enemyTexture.Height))
					rl.DrawTexturePro(enemyTexture, sourceRect, destRectFlipped, origin, 0, enemyTint)
				}

				enemyHealthBarWidth := enemies[i].Width
				enemyHealthBarHeight := float32(5)
				enemyHealthBarX := enemies[i].X
				enemyHealthBarY := enemies[i].Y - enemyHealthBarHeight - 3
				enemyHealthPercent := float32(enemies[i].Health) / float32(enemies[i].MaxHealth)
				if enemyHealthPercent < 0 {
					enemyHealthPercent = 0
				}
				if enemyHealthPercent > 1 {
					enemyHealthPercent = 1
				}

				rl.DrawRectangleRec(rl.NewRectangle(enemyHealthBarX, enemyHealthBarY, enemyHealthBarWidth, enemyHealthBarHeight), rl.Red)
				rl.DrawRectangleRec(rl.NewRectangle(enemyHealthBarX, enemyHealthBarY, enemyHealthBarWidth*enemyHealthPercent, enemyHealthBarHeight), rl.Green)
			}
		}

		rl.EndDrawing()
	}
}
