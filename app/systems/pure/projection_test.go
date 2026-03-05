package pure

import (
	"math"
	"testing"
)

func TestProjectionRoundTrip(t *testing.T) {
	points := [][2]float32{
		{0, 0},
		{120, 75},
		{640, 420},
		{1800, 300},
		{2550, 980},
	}

	cameraX := float32(320)
	cameraY := float32(140)
	cameraWidth := float32(1600)
	cameraHeight := float32(900)

	for _, point := range points {
		sx, sy := WorldToScreenIso(point[0], point[1], cameraX, cameraY, cameraWidth, cameraHeight)
		wx, wy := ScreenToWorldIso(sx, sy, cameraX, cameraY, cameraWidth, cameraHeight)
		if math.Abs(float64(wx-point[0])) > 0.001 {
			t.Fatalf("x mismatch: got %.6f want %.6f", wx, point[0])
		}
		if math.Abs(float64(wy-point[1])) > 0.001 {
			t.Fatalf("y mismatch: got %.6f want %.6f", wy, point[1])
		}
	}
}
