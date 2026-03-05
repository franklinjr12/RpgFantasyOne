package game

import (
	"math"
	"testing"

	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"singlefantasy/app/settings"
)

func nearlyEqual(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.001
}

func TestSmoothAxisVelocity(t *testing.T) {
	v := smoothAxisVelocity(0, 100, 200, 300, 0.1)
	if !nearlyEqual(v, 20) {
		t.Fatalf("expected accelerated velocity 20, got %.2f", v)
	}

	v = smoothAxisVelocity(v, 0, 200, 300, 0.1)
	if !nearlyEqual(v, 0) {
		t.Fatalf("expected decelerated velocity 0, got %.2f", v)
	}
}

func TestDecayAxisVelocity(t *testing.T) {
	v := decayAxisVelocity(100, 240, 0.25)
	if !nearlyEqual(v, 40) {
		t.Fatalf("expected decayed velocity 40, got %.2f", v)
	}

	v = decayAxisVelocity(v, 240, 0.25)
	if !nearlyEqual(v, 0) {
		t.Fatalf("expected velocity snap to 0 after second decay, got %.2f", v)
	}
}

func TestDampBlockedAxis(t *testing.T) {
	moveVelocity, knockbackVelocity := dampBlockedAxis(120, 80, 10, 2)
	if !nearlyEqual(moveVelocity, 0) {
		t.Fatalf("expected blocked move velocity to reset, got %.2f", moveVelocity)
	}
	if !nearlyEqual(knockbackVelocity, 16) {
		t.Fatalf("expected damped knockback velocity 16, got %.2f", knockbackVelocity)
	}

	moveVelocity, knockbackVelocity = dampBlockedAxis(120, 80, 10, 9)
	if !nearlyEqual(moveVelocity, 120) || !nearlyEqual(knockbackVelocity, 80) {
		t.Fatalf("expected unblocked axis to keep velocities, got move=%.2f knockback=%.2f", moveVelocity, knockbackVelocity)
	}
}

func TestApplyPlayerDirectHitIFrames(t *testing.T) {
	g := NewGame(settings.Default())
	g.Player = gameobjects.NewPlayer(100, 100, gamedata.ClassTypeMelee)

	startHP := g.Player.HP
	if !g.ApplyPlayerDirectHit(10, 80, 100) {
		t.Fatalf("expected first hit to apply")
	}
	if g.Player.HP != startHP-10 {
		t.Fatalf("expected hp %d after hit, got %d", startHP-10, g.Player.HP)
	}
	if g.Player.HurtIFrameTimer <= 0 {
		t.Fatalf("expected hurt iframe timer to be active")
	}

	hpAfterFirstHit := g.Player.HP
	if g.ApplyPlayerDirectHit(10, 80, 100) {
		t.Fatalf("expected second hit during iframe to be ignored")
	}
	if g.Player.HP != hpAfterFirstHit {
		t.Fatalf("expected hp unchanged during iframe, got %d", g.Player.HP)
	}

	g.Player.HurtIFrameTimer = 0
	if !g.ApplyPlayerDirectHit(7, 80, 100) {
		t.Fatalf("expected hit to apply after iframe expires")
	}
	if g.Player.HP != hpAfterFirstHit-7 {
		t.Fatalf("expected hp %d after iframe expiry hit, got %d", hpAfterFirstHit-7, g.Player.HP)
	}
}
