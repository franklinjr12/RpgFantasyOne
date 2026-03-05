package pure

import (
	"testing"

	"singlefantasy/app/world"
)

func TestResolvePlayerMovementBoundsClamp(t *testing.T) {
	room := &world.Room{X: 0, Y: 0, Width: 300, Height: 200}
	x, y := ResolvePlayerMovement(10, 20, 30, 30, -50, -80, room)
	if x != 0 || y != 0 {
		t.Fatalf("expected clamp to bounds, got (%.2f, %.2f)", x, y)
	}
}

func TestResolvePlayerMovementObstacleBlock(t *testing.T) {
	room := &world.Room{
		X: 0, Y: 0, Width: 300, Height: 300,
		Obstacles: []world.AABB{{X: 120, Y: 120, Width: 60, Height: 60}},
	}

	x, y := ResolvePlayerMovement(70, 130, 30, 30, 80, 0, room)
	if x != 90 {
		t.Fatalf("expected obstacle stop at x=90, got %.2f", x)
	}
	if y != 130 {
		t.Fatalf("expected unchanged y=130, got %.2f", y)
	}
}

func TestResolvePlayerMovementSliding(t *testing.T) {
	room := &world.Room{
		X: 0, Y: 0, Width: 300, Height: 300,
		Obstacles: []world.AABB{{X: 120, Y: 120, Width: 60, Height: 60}},
	}

	x, y := ResolvePlayerMovement(80, 80, 30, 30, 50, 50, room)
	if x != 130 {
		t.Fatalf("expected x to advance to 130 while sliding, got %.2f", x)
	}
	if y != 90 {
		t.Fatalf("expected y to resolve against obstacle top at 90, got %.2f", y)
	}
}
