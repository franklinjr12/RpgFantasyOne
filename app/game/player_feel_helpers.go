package game

import "math"

func moveToward(current, target, maxDelta float32) float32 {
	if maxDelta <= 0 {
		return current
	}
	if current < target {
		next := current + maxDelta
		if next > target {
			return target
		}
		return next
	}
	if current > target {
		next := current - maxDelta
		if next < target {
			return target
		}
		return next
	}
	return current
}

func smoothAxisVelocity(current, desired, acceleration, deceleration, dt float32) float32 {
	if dt <= 0 {
		return current
	}
	if desired == 0 {
		return snapSmallVelocity(moveToward(current, 0, deceleration*dt))
	}
	return moveToward(current, desired, acceleration*dt)
}

func decayAxisVelocity(current, decayPerSecond, dt float32) float32 {
	if dt <= 0 {
		return current
	}
	return snapSmallVelocity(moveToward(current, 0, decayPerSecond*dt))
}

func dampBlockedAxis(moveVelocity, knockbackVelocity, attemptedDelta, appliedDelta float32) (float32, float32) {
	attemptedAbs := float32(math.Abs(float64(attemptedDelta)))
	if attemptedAbs <= PlayerVelocitySnapThreshold {
		return moveVelocity, knockbackVelocity
	}

	appliedAbs := float32(math.Abs(float64(appliedDelta)))
	signMismatch := attemptedDelta*appliedDelta < 0
	if !signMismatch && appliedAbs >= attemptedAbs*PlayerCollisionBlockedRatio {
		return moveVelocity, knockbackVelocity
	}

	moveVelocity = 0
	knockbackVelocity *= PlayerBlockedAxisKnockbackDamping
	return moveVelocity, snapSmallVelocity(knockbackVelocity)
}

func snapSmallVelocity(value float32) float32 {
	if float32(math.Abs(float64(value))) < PlayerVelocitySnapThreshold {
		return 0
	}
	return value
}
