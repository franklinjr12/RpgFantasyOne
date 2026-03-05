package systems

import (
	"math"

	"singlefantasy/app/systems/pure"
)

func WorldToScreenIso(worldX, worldY float32, camera *Camera) (float32, float32) {
	if camera == nil {
		camera = NewCamera()
	}
	return pure.WorldToScreenIso(worldX, worldY, camera.X, camera.Y, camera.Width, camera.Height)
}

func ScreenToWorldIso(screenX, screenY float32, camera *Camera) (float32, float32) {
	if camera == nil {
		camera = NewCamera()
	}
	return pure.ScreenToWorldIso(screenX, screenY, camera.X, camera.Y, camera.Width, camera.Height)
}

func WorldToScreen(worldX, worldY float32, camera *Camera) (float32, float32) {
	return WorldToScreenIso(worldX, worldY, camera)
}

func ScreenToWorld(screenX, screenY float32, camera *Camera) (float32, float32) {
	return ScreenToWorldIso(screenX, screenY, camera)
}

func IsometricToScreen(isoX, isoY float32) (float32, float32) {
	return pure.IsometricToScreen(isoX, isoY)
}

func ScreenToIsometric(screenX, screenY float32) (float32, float32) {
	return pure.ScreenToIsometric(screenX, screenY)
}

func clamp(value, minValue, maxValue float32) float32 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func nearlyEqual(a, b, epsilon float32) bool {
	return float32(math.Abs(float64(a-b))) <= epsilon
}
