package pure

import "singlefantasy/app/world"

func AABBOverlap(a, b world.AABB) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}

func ResolvePlayerMovement(
	posX, posY,
	width, height,
	deltaX, deltaY float32,
	room *world.Room,
) (float32, float32) {
	if room == nil {
		return posX + deltaX, posY + deltaY
	}

	roomBounds := world.AABB{
		X:      room.X,
		Y:      room.Y,
		Width:  room.Width,
		Height: room.Height,
	}

	newX := resolveX(posX, posY, width, height, deltaX, roomBounds, room.Obstacles)
	newY := resolveY(newX, posY, width, height, deltaY, roomBounds, room.Obstacles)
	return newX, newY
}

func resolveX(posX, posY, width, height, deltaX float32, bounds world.AABB, obstacles []world.AABB) float32 {
	newX := posX + deltaX
	minX := bounds.X
	maxX := bounds.X + bounds.Width - width
	if maxX < minX {
		maxX = minX
	}
	newX = clamp(newX, minX, maxX)

	playerRect := world.AABB{X: newX, Y: posY, Width: width, Height: height}
	for _, obstacle := range obstacles {
		if !AABBOverlap(playerRect, obstacle) {
			continue
		}
		if deltaX > 0 {
			newX = obstacle.X - width
		} else if deltaX < 0 {
			newX = obstacle.X + obstacle.Width
		}
		newX = clamp(newX, minX, maxX)
		playerRect.X = newX
	}

	return newX
}

func resolveY(posX, posY, width, height, deltaY float32, bounds world.AABB, obstacles []world.AABB) float32 {
	newY := posY + deltaY
	minY := bounds.Y
	maxY := bounds.Y + bounds.Height - height
	if maxY < minY {
		maxY = minY
	}
	newY = clamp(newY, minY, maxY)

	playerRect := world.AABB{X: posX, Y: newY, Width: width, Height: height}
	for _, obstacle := range obstacles {
		if !AABBOverlap(playerRect, obstacle) {
			continue
		}
		if deltaY > 0 {
			newY = obstacle.Y - height
		} else if deltaY < 0 {
			newY = obstacle.Y + obstacle.Height
		}
		newY = clamp(newY, minY, maxY)
		playerRect.Y = newY
	}

	return newY
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
