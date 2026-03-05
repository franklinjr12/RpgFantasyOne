package pure

const (
	IsoXScale       = 0.5
	IsoYScale       = 0.25
	IsoVerticalBias = 0.18
)

func WorldToScreenIso(worldX, worldY, cameraX, cameraY, cameraWidth, cameraHeight float32) (float32, float32) {
	localX := worldX - cameraX
	localY := worldY - cameraY
	screenX := (localX-localY)*IsoXScale + cameraWidth*0.5
	screenY := (localX+localY)*IsoYScale + cameraHeight*IsoVerticalBias
	return screenX, screenY
}

func ScreenToWorldIso(screenX, screenY, cameraX, cameraY, cameraWidth, cameraHeight float32) (float32, float32) {
	localScreenX := screenX - cameraWidth*0.5
	localScreenY := screenY - cameraHeight*IsoVerticalBias
	localX := localScreenY*2 + localScreenX
	localY := localScreenY*2 - localScreenX
	worldX := localX + cameraX
	worldY := localY + cameraY
	return worldX, worldY
}

func IsometricToScreen(isoX, isoY float32) (float32, float32) {
	screenX := (isoX - isoY) * IsoXScale
	screenY := (isoX + isoY) * IsoYScale
	return screenX, screenY
}

func ScreenToIsometric(screenX, screenY float32) (float32, float32) {
	isoX := screenY*2 + screenX
	isoY := screenY*2 - screenX
	return isoX, isoY
}
