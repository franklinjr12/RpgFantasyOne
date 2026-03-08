package gameobjects

import (
	"testing"

	"singlefantasy/app/gamedata"
)

func TestManaShieldAbsorbsDamageAndExpires(t *testing.T) {
	player := NewPlayer(0, 0, gamedata.ClassTypeCaster)
	startHP := player.HP

	player.SetManaShield(20, 1.0)
	applied := player.ApplyTypedDamage(10, gamedata.DamagePhysical, false)
	if applied != 0 {
		t.Fatalf("expected shield to absorb all damage, applied=%d", applied)
	}
	if player.HP != startHP {
		t.Fatalf("expected hp unchanged while shield absorbs damage")
	}
	if player.ManaShieldAmount != 11 {
		t.Fatalf("expected shield amount to reduce to 11 after mitigation, got %d", player.ManaShieldAmount)
	}

	player.Update(1.1)
	if player.ManaShieldActive {
		t.Fatalf("expected shield to expire after duration")
	}
	if player.ManaShieldAmount != 0 {
		t.Fatalf("expected shield amount cleared on expiry")
	}
}

func TestManaShieldWithoutDurationPersistsUntilDepleted(t *testing.T) {
	player := NewPlayer(0, 0, gamedata.ClassTypeCaster)
	player.SetManaShield(5, 0)
	player.Update(10)
	if !player.ManaShieldActive {
		t.Fatalf("expected no-duration shield to persist")
	}

	player.ApplyTypedDamage(10, gamedata.DamagePhysical, false)
	if player.ManaShieldActive {
		t.Fatalf("expected shield to deactivate when depleted")
	}
}
