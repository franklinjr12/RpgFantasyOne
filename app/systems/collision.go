package systems

import (
	"singlefantasy/app/systems/pure"
	"singlefantasy/app/world"
)

func AABBOverlap(a, b world.AABB) bool {
	return pure.AABBOverlap(a, b)
}

func ResolvePlayerMovement(
	posX, posY,
	width, height,
	deltaX, deltaY float32,
	room *world.Room,
) (float32, float32) {
	return pure.ResolvePlayerMovement(posX, posY, width, height, deltaX, deltaY, room)
}
