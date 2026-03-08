package gameobjects

import (
	"testing"

	"singlefantasy/app/gamedata"
)

func TestMageManaRegenerationWorksWithFixedStep(t *testing.T) {
	player := NewPlayer(0, 0, gamedata.ClassTypeCaster)
	player.Mana = 0

	for i := 0; i < 60; i++ {
		player.Update(1.0 / 60.0)
	}

	if player.Mana <= 0 {
		t.Fatalf("expected mage mana to regenerate above 0 after one second, got %d", player.Mana)
	}
}

func TestMageRegeneratesMoreManaThanMelee(t *testing.T) {
	mage := NewPlayer(0, 0, gamedata.ClassTypeCaster)
	melee := NewPlayer(0, 0, gamedata.ClassTypeMelee)
	mage.Mana = 0
	melee.Mana = 0

	for i := 0; i < 60; i++ {
		mage.Update(1.0 / 60.0)
		melee.Update(1.0 / 60.0)
	}

	if mage.Mana <= melee.Mana {
		t.Fatalf("expected mage mana regen (%d) to exceed melee (%d)", mage.Mana, melee.Mana)
	}
}
