package pure

func WorldToScreenIso(worldX, worldY, cameraX, cameraY, cameraWidth, cameraHeight float32) (float32, float32) {
	return worldX - cameraX, worldY - cameraY
}

func ScreenToWorldIso(screenX, screenY, cameraX, cameraY, cameraWidth, cameraHeight float32) (float32, float32) {
	return screenX + cameraX, screenY + cameraY
}

func IsometricToScreen(isoX, isoY float32) (float32, float32) {
	return isoX, isoY
}

func ScreenToIsometric(screenX, screenY float32) (float32, float32) {
	return screenX, screenY
}
