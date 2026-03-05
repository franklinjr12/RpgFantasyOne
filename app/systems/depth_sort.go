package systems

import (
	"sort"

	"singlefantasy/app/systems/pure"
)

type RenderQueueItem struct {
	DepthY   float32
	DepthX   float32
	StableID int
	Draw     func()
}

func SortRenderQueue(items []RenderQueueItem) {
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		return pure.DepthLess(left.DepthY, left.DepthX, left.StableID, right.DepthY, right.DepthX, right.StableID)
	})
}

func DepthSortKey(worldX, worldY float32) (float32, float32) {
	return pure.DepthSortKey(worldX, worldY)
}
