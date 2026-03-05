package pure

func DepthLess(depthAY, depthAX float32, stableIDA int, depthBY, depthBX float32, stableIDB int) bool {
	if depthAY != depthBY {
		return depthAY < depthBY
	}
	if depthAX != depthBX {
		return depthAX < depthBX
	}
	return stableIDA < stableIDB
}

func DepthSortKey(worldX, worldY float32) (float32, float32) {
	return worldY, worldX
}
