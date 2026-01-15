package game

import (
	"singlefantasy/app/gamedata"
	"singlefantasy/app/systems"
)

func (g *Game) TryCastSkill(skill *gamedata.Skill, input *systems.Input) {
	if !systems.CanCast(g.Player, skill) {
		return
	}

	mouseX, mouseY := systems.GetMousePosition()
	worldX, worldY := systems.ScreenToWorld(mouseX, mouseY, g.Camera)
	playerCenterX := g.Player.X + g.Player.Width/2
	playerCenterY := g.Player.Y + g.Player.Height/2

	if skill.ManaCost > 0 {
		if !g.Player.CanUseMana(skill.ManaCost) {
			return
		}
		g.Player.UseMana(skill.ManaCost)
	}

	switch skill.Delivery.Type {
	case gamedata.DeliveryInstant:
		targets := systems.ResolveTargets(g.Player, worldX, worldY, skill.Targeting, g.Enemies, g.Boss)

		if skill.Type == gamedata.SkillTypeRetreatRoll {
			dx := playerCenterX - worldX
			dy := playerCenterY - worldY
			distance := systems.GetDistance(0, 0, dx, dy)
			if distance > 0 {
				rollDistance := float32(80)
				g.Player.X += (dx / distance) * rollDistance
				g.Player.Y += (dy / distance) * rollDistance
			}
		}

		if skill.Type == gamedata.SkillTypeManaShield {
			g.Player.ManaShieldActive = true
			g.Player.ManaShieldAmount = g.Player.Mana / 2
		}

		systems.ApplySkill(g.Player, skill, targets)

		if skill.Type == gamedata.SkillTypeArcaneDrain {
			manaRestore := len(targets) * 10
			g.Player.Mana += manaRestore
			if g.Player.Mana > g.Player.MaxMana {
				g.Player.Mana = g.Player.MaxMana
			}
		}

		skill.Use()

	case gamedata.DeliveryProjectile:
		dx := worldX - playerCenterX
		dy := worldY - playerCenterY
		distance := systems.GetDistance(0, 0, dx, dy)
		if distance > 0 {
			proj := &Projectile{
				X:         playerCenterX,
				Y:         playerCenterY,
				VX:        (dx / distance) * skill.Delivery.Speed,
				VY:        (dy / distance) * skill.Delivery.Speed,
				Speed:     skill.Delivery.Speed,
				Damage:    int(systems.ComputeDamage(skill.DamageSpec, g.Player.Stats)),
				Radius:    5,
				Alive:     true,
				Skill:     skill,
				Caster:    g.Player,
			}
			g.Projectiles = append(g.Projectiles, proj)
		}
		skill.Use()
	}
}

