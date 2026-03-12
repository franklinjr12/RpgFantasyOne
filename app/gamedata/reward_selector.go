package gamedata

import (
	"math/rand"
	"sort"
	"strings"
)

type RewardContext int

const (
	RewardContextNone RewardContext = iota
	RewardContextBoss
	RewardContextMilestone
)

type RewardOfferHistoryEntry struct {
	Context  RewardContext
	ItemIDs  []string
	OfferKey string
}

type RewardSelectionRequest struct {
	ClassType ClassType
	Biome     string
	Context   RewardContext
	OfferSize int
	Seed      int64
	History   []RewardOfferHistoryEntry
}

func (context RewardContext) DefaultOfferSize() int {
	switch context {
	case RewardContextMilestone:
		return 2
	case RewardContextBoss:
		return 3
	default:
		return 3
	}
}

func BuildRewardHistoryEntry(context RewardContext, items []*Item) RewardOfferHistoryEntry {
	ids := sortedItemIDs(items)
	return RewardOfferHistoryEntry{
		Context:  context,
		ItemIDs:  ids,
		OfferKey: rewardOfferKey(ids),
	}
}

func SelectRewardOptions(request RewardSelectionRequest) []*Item {
	context := request.Context
	if context == RewardContextNone {
		context = RewardContextBoss
	}

	offerSize := request.OfferSize
	if offerSize <= 0 {
		offerSize = context.DefaultOfferSize()
	}

	pool := rewardPoolForRequest(request.ClassType, request.Biome)
	if len(pool) == 0 {
		return nil
	}
	if offerSize > len(pool) {
		offerSize = len(pool)
	}

	historyKeys := make(map[string]struct{}, len(request.History))
	for _, entry := range request.History {
		key := entry.OfferKey
		if key == "" {
			key = rewardOfferKey(entry.ItemIDs)
		}
		if key == "" {
			continue
		}
		historyKeys[key] = struct{}{}
	}

	const maxAttempts = 6
	var fallback []*Item
	for attempt := 0; attempt < maxAttempts; attempt++ {
		rng := rand.New(rand.NewSource(deriveRewardSeed(request.Seed, context, len(request.History), attempt)))
		candidate := weightedSampleWithoutReplacement(pool, offerSize, context, rng)
		if len(candidate) == 0 {
			continue
		}

		key := rewardOfferKeyFromItems(candidate)
		if _, seen := historyKeys[key]; !seen {
			return cloneItems(candidate)
		}
		if len(fallback) == 0 {
			fallback = candidate
		}
	}

	if len(fallback) == 0 {
		rng := rand.New(rand.NewSource(deriveRewardSeed(request.Seed, context, len(request.History), maxAttempts)))
		fallback = weightedSampleWithoutReplacement(pool, offerSize, context, rng)
	}
	if len(fallback) == 0 {
		return nil
	}

	diversified := diversifyOffer(fallback, pool, historyKeys)
	if len(diversified) > 0 {
		return cloneItems(diversified)
	}
	return cloneItems(fallback)
}

func rewardPoolForRequest(classType ClassType, biome string) []*Item {
	source := GetBiomeItemPool(biome)
	filtered := make([]*Item, 0, len(source))
	seen := map[string]struct{}{}
	for _, item := range source {
		if item == nil || !item.IsClassAllowed(classType) {
			continue
		}
		if _, exists := seen[item.ID]; exists {
			continue
		}
		seen[item.ID] = struct{}{}
		filtered = append(filtered, item)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].Weight == filtered[j].Weight {
			return filtered[i].ID < filtered[j].ID
		}
		return filtered[i].Weight > filtered[j].Weight
	})
	return filtered
}

func weightedSampleWithoutReplacement(pool []*Item, desired int, context RewardContext, rng *rand.Rand) []*Item {
	if desired <= 0 || len(pool) == 0 {
		return nil
	}

	available := make([]*Item, len(pool))
	copy(available, pool)
	selected := make([]*Item, 0, desired)

	for len(selected) < desired && len(available) > 0 {
		totalWeight := 0
		for _, item := range available {
			totalWeight += contextualWeight(item, context)
		}
		if totalWeight <= 0 {
			break
		}

		roll := rng.Intn(totalWeight)
		cursor := 0
		pickedIndex := 0
		for i, item := range available {
			cursor += contextualWeight(item, context)
			if roll < cursor {
				pickedIndex = i
				break
			}
		}

		selected = append(selected, available[pickedIndex])
		available = append(available[:pickedIndex], available[pickedIndex+1:]...)
	}

	return selected
}

func contextualWeight(item *Item, context RewardContext) int {
	if item == nil {
		return 1
	}

	weight := item.Weight
	if weight <= 0 {
		weight = 1
	}

	switch context {
	case RewardContextMilestone:
		if item.Slot == ItemSlotWeapon {
			weight /= 2
		}
	case RewardContextBoss:
		if item.Slot == ItemSlotWeapon {
			weight += 2
		}
	}

	if weight <= 0 {
		return 1
	}
	return weight
}

func deriveRewardSeed(seed int64, context RewardContext, historyLen, attempt int) int64 {
	base := seed
	if base == 0 {
		base = DefaultRewardSeed
	}
	return base + int64(context)*1000003 + int64(historyLen)*4096 + int64(attempt)*97531
}

func diversifyOffer(current []*Item, pool []*Item, historyKeys map[string]struct{}) []*Item {
	if len(current) == 0 {
		return nil
	}
	if len(pool) <= len(current) {
		return current
	}

	currentIDs := map[string]struct{}{}
	for _, item := range current {
		if item == nil {
			continue
		}
		currentIDs[item.ID] = struct{}{}
	}

	for _, candidate := range pool {
		if candidate == nil {
			continue
		}
		if _, exists := currentIDs[candidate.ID]; exists {
			continue
		}

		for index := range current {
			trial := make([]*Item, len(current))
			copy(trial, current)
			trial[index] = candidate
			key := rewardOfferKeyFromItems(trial)
			if _, seen := historyKeys[key]; seen {
				continue
			}
			return trial
		}
	}

	return current
}

func rewardOfferKeyFromItems(items []*Item) string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		ids = append(ids, item.ID)
	}
	return rewardOfferKey(ids)
}

func rewardOfferKey(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	normalized := make([]string, 0, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) == 0 {
		return ""
	}
	sort.Strings(normalized)
	return strings.Join(normalized, "|")
}
