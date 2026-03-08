package systems

import (
	"math"
	"singlefantasy/app/gamedata"
	"singlefantasy/app/gameobjects"
	"sort"
)

type CastIntent struct {
	CursorX    float32
	CursorY    float32
	DirectionX float32
	DirectionY float32
}

type targetCandidate struct {
	Target    interface{}
	CenterX   float32
	CenterY   float32
	Distance2 float32
	Order     int
}

func BuildCastIntent(caster *gameobjects.Player, cursorX, cursorY float32) CastIntent {
	intent := CastIntent{
		CursorX: cursorX,
		CursorY: cursorY,
	}
	if caster == nil {
		return intent
	}

	casterX, casterY := caster.Center()
	dx := cursorX - casterX
	dy := cursorY - casterY
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if distance > 0 {
		intent.DirectionX = dx / distance
		intent.DirectionY = dy / distance
		return intent
	}

	if caster.FacingRight {
		intent.DirectionX = 1
		intent.DirectionY = 0
	} else {
		intent.DirectionX = -1
		intent.DirectionY = 0
	}
	return intent
}

func ResolveTargets(caster *gameobjects.Player, intent CastIntent, spec gamedata.TargetingSpec, enemies []*gameobjects.Enemy, boss *gameobjects.Boss) []interface{} {
	if caster == nil {
		return nil
	}

	casterX, casterY := caster.Center()
	candidates := gatherCandidates(casterX, casterY, enemies, boss)

	switch spec.Type {
	case gamedata.TargetSelf:
		return []interface{}{caster}
	case gamedata.TargetEnemy:
		return resolveEnemyTargets(intent, spec, candidates)
	case gamedata.TargetArea:
		return resolveAreaTargets(casterX, casterY, intent, spec, candidates)
	case gamedata.TargetDirection:
		return resolveDirectionalTargets(casterX, casterY, intent, spec, candidates)
	default:
		return nil
	}
}

func resolveEnemyTargets(intent CastIntent, spec gamedata.TargetingSpec, candidates []targetCandidate) []interface{} {
	maxRange := spec.Range
	if maxRange <= 0 {
		maxRange = float32(math.MaxFloat32)
	}
	maxRange2 := maxRange * maxRange

	inRange := make([]targetCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Distance2 > maxRange2 {
			continue
		}
		inRange = append(inRange, candidate)
	}
	if len(inRange) == 0 {
		return nil
	}

	sortCandidates(inRange)
	singleTarget := spec.MaxTargets == 1
	if singleTarget {
		for _, candidate := range inRange {
			if isPointOverTarget(candidate.Target, intent.CursorX, intent.CursorY) {
				return []interface{}{candidate.Target}
			}
		}
		return []interface{}{inRange[0].Target}
	}

	return selectTargets(inRange, spec.MaxTargets)
}

func resolveAreaTargets(casterX, casterY float32, intent CastIntent, spec gamedata.TargetingSpec, candidates []targetCandidate) []interface{} {
	centerX := intent.CursorX
	centerY := intent.CursorY
	if spec.Range == 0 {
		centerX = casterX
		centerY = casterY
	}

	if spec.Range > 0 {
		dx := centerX - casterX
		dy := centerY - casterY
		distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if distance > spec.Range && distance > 0 {
			ratio := spec.Range / distance
			centerX = casterX + dx*ratio
			centerY = casterY + dy*ratio
		}
	}

	radius2 := spec.Radius * spec.Radius
	filtered := make([]targetCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		dx := candidate.CenterX - centerX
		dy := candidate.CenterY - centerY
		distance2 := dx*dx + dy*dy
		if distance2 > radius2 {
			continue
		}
		filtered = append(filtered, targetCandidate{
			Target:    candidate.Target,
			CenterX:   candidate.CenterX,
			CenterY:   candidate.CenterY,
			Distance2: distance2,
			Order:     candidate.Order,
		})
	}

	sortCandidates(filtered)
	return selectTargets(filtered, spec.MaxTargets)
}

func resolveDirectionalTargets(casterX, casterY float32, intent CastIntent, spec gamedata.TargetingSpec, candidates []targetCandidate) []interface{} {
	rangeLimit := spec.Range
	if rangeLimit <= 0 {
		return nil
	}

	dirX := intent.DirectionX
	dirY := intent.DirectionY
	dirLen := float32(math.Sqrt(float64(dirX*dirX + dirY*dirY)))
	if dirLen <= 0 {
		return nil
	}
	dirX /= dirLen
	dirY /= dirLen

	arc := spec.DirectionalArcDegrees
	if arc <= 0 {
		arc = 60
	}
	halfArc := arc * 0.5
	lineWidth := spec.DirectionalLineWidth
	rangeLimit2 := rangeLimit * rangeLimit

	filtered := make([]targetCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		dx := candidate.CenterX - casterX
		dy := candidate.CenterY - casterY
		distance2 := dx*dx + dy*dy
		if distance2 > rangeLimit2 {
			continue
		}

		projection := dx*dirX + dy*dirY
		if projection < 0 {
			continue
		}

		if lineWidth > 0 {
			if projection > rangeLimit {
				continue
			}
			perp := float32(math.Abs(float64(dx*dirY - dy*dirX)))
			if perp > lineWidth*0.5 {
				continue
			}
		} else {
			distance := float32(math.Sqrt(float64(distance2)))
			if distance > 0 {
				cosTheta := projection / distance
				if cosTheta > 1 {
					cosTheta = 1
				}
				if cosTheta < -1 {
					cosTheta = -1
				}
				angle := float32(math.Acos(float64(cosTheta)) * 180 / math.Pi)
				if angle > halfArc {
					continue
				}
			}
		}

		filtered = append(filtered, targetCandidate{
			Target:    candidate.Target,
			CenterX:   candidate.CenterX,
			CenterY:   candidate.CenterY,
			Distance2: distance2,
			Order:     candidate.Order,
		})
	}

	sortCandidates(filtered)
	return selectTargets(filtered, spec.MaxTargets)
}

func gatherCandidates(casterX, casterY float32, enemies []*gameobjects.Enemy, boss *gameobjects.Boss) []targetCandidate {
	candidates := make([]targetCandidate, 0, len(enemies)+1)
	order := 0
	for _, enemy := range enemies {
		if enemy == nil || !enemy.IsAlive() {
			order++
			continue
		}
		enemyX, enemyY := enemy.Center()
		dx := enemyX - casterX
		dy := enemyY - casterY
		candidates = append(candidates, targetCandidate{
			Target:    enemy,
			CenterX:   enemyX,
			CenterY:   enemyY,
			Distance2: dx*dx + dy*dy,
			Order:     order,
		})
		order++
	}

	if boss != nil && boss.IsAlive() {
		bossX, bossY := boss.Center()
		dx := bossX - casterX
		dy := bossY - casterY
		candidates = append(candidates, targetCandidate{
			Target:    boss,
			CenterX:   bossX,
			CenterY:   bossY,
			Distance2: dx*dx + dy*dy,
			Order:     order,
		})
	}

	return candidates
}

func sortCandidates(candidates []targetCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Distance2 == candidates[j].Distance2 {
			return candidates[i].Order < candidates[j].Order
		}
		return candidates[i].Distance2 < candidates[j].Distance2
	})
}

func selectTargets(candidates []targetCandidate, maxTargets int) []interface{} {
	if len(candidates) == 0 {
		return nil
	}

	limit := len(candidates)
	if maxTargets > 0 && maxTargets < limit {
		limit = maxTargets
	}

	targets := make([]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		targets = append(targets, candidates[i].Target)
	}
	return targets
}

func isPointOverTarget(target interface{}, x, y float32) bool {
	switch t := target.(type) {
	case *gameobjects.Enemy:
		return pointInAABB(x, y, t.PosX, t.PosY, t.Hitbox.Width, t.Hitbox.Height)
	case *gameobjects.Boss:
		return pointInAABB(x, y, t.PosX, t.PosY, t.Hitbox.Width, t.Hitbox.Height)
	default:
		return false
	}
}

func pointInAABB(x, y, minX, minY, width, height float32) bool {
	return x >= minX && x <= minX+width && y >= minY && y <= minY+height
}

func CanCast(caster *gameobjects.Player, skill *gamedata.Skill) bool {
	if caster == nil || skill == nil {
		return false
	}

	if !skill.CanUse() {
		return false
	}

	if skill.ManaCost > 0 && !caster.CanUseMana(skill.ManaCost) {
		return false
	}

	if !gamedata.CanCast(&caster.Effects) {
		return false
	}

	return true
}
