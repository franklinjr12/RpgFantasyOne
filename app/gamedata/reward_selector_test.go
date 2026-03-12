package gamedata

import (
	"math/rand"
	"testing"
)

func TestForestBiomeItemPoolMeetsCurationTargets(t *testing.T) {
	if total := CountBiomeItems("forest"); total < 30 {
		t.Fatalf("expected at least 30 curated forest items, got %d", total)
	}

	classes := []ClassType{ClassTypeMelee, ClassTypeRanged, ClassTypeCaster}
	for _, classType := range classes {
		if count := CountBiomeFlavorItems("forest", classType); count < 10 {
			t.Fatalf("expected class %d to have at least 10 flavor items, got %d", classType, count)
		}
	}

	pool := GetBiomeItemPool("forest")
	seenIDs := map[string]struct{}{}
	hasBurn := false
	hasCritVsSlow := false
	for _, item := range pool {
		if item == nil {
			continue
		}
		if item.ID == "" {
			t.Fatalf("expected item id to be set for %q", item.Name)
		}
		if _, exists := seenIDs[item.ID]; exists {
			t.Fatalf("duplicate item id found: %s", item.ID)
		}
		seenIDs[item.ID] = struct{}{}

		for _, effect := range item.Effects {
			if effect.Type == ItemEffectBurnOnHit {
				hasBurn = true
			}
			if effect.Type == ItemEffectCritChanceVsSlowed {
				hasCritVsSlow = true
			}
		}
	}

	if !hasBurn {
		t.Fatalf("expected at least one burn-on-hit item")
	}
	if !hasCritVsSlow {
		t.Fatalf("expected at least one crit-vs-slowed item")
	}
}

func TestSelectRewardOptionsIsDeterministicForSameRequest(t *testing.T) {
	request := RewardSelectionRequest{
		ClassType: ClassTypeMelee,
		Biome:     "forest",
		Context:   RewardContextBoss,
		OfferSize: 3,
		Seed:      901,
	}

	left := SelectRewardOptions(request)
	right := SelectRewardOptions(request)

	if len(left) != 3 || len(right) != 3 {
		t.Fatalf("expected both selections to have size 3, got %d and %d", len(left), len(right))
	}

	for i := range left {
		if left[i] == nil || right[i] == nil {
			t.Fatalf("expected non-nil options at index %d", i)
		}
		if left[i].ID != right[i].ID {
			t.Fatalf("determinism mismatch at index %d: %s vs %s", i, left[i].ID, right[i].ID)
		}
	}
}

func TestSelectRewardOptionsProducesUniqueOfferEntries(t *testing.T) {
	request := RewardSelectionRequest{
		ClassType: ClassTypeCaster,
		Biome:     "forest",
		Context:   RewardContextBoss,
		OfferSize: 3,
		Seed:      77,
	}

	options := SelectRewardOptions(request)
	if len(options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(options))
	}

	seen := map[string]struct{}{}
	for i, item := range options {
		if item == nil {
			t.Fatalf("expected non-nil option at index %d", i)
		}
		if _, exists := seen[item.ID]; exists {
			t.Fatalf("duplicate item in offer: %s", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
}

func TestSelectRewardOptionsAppliesAntiRepeatHistory(t *testing.T) {
	request := RewardSelectionRequest{
		ClassType: ClassTypeRanged,
		Biome:     "forest",
		Context:   RewardContextBoss,
		OfferSize: 3,
		Seed:      404,
	}

	first := SelectRewardOptions(request)
	if len(first) != 3 {
		t.Fatalf("expected first offer size 3, got %d", len(first))
	}
	firstKey := rewardOfferKeyFromItems(first)

	request.History = append(request.History, BuildRewardHistoryEntry(RewardContextBoss, first))
	second := SelectRewardOptions(request)
	if len(second) != 3 {
		t.Fatalf("expected second offer size 3, got %d", len(second))
	}
	secondKey := rewardOfferKeyFromItems(second)

	if firstKey == secondKey {
		t.Fatalf("expected anti-repeat to avoid identical offer sets, got %s", secondKey)
	}
}

func TestWeightedSampleWithoutReplacementRespectsWeightBias(t *testing.T) {
	highWeight := NewCuratedItem(
		"high_weight",
		"High Weight",
		"",
		ItemSlotHead,
		map[StatType]int{},
		ClassTypeAny,
		ItemMetadata{Weight: 100},
	)
	lowWeight := NewCuratedItem(
		"low_weight",
		"Low Weight",
		"",
		ItemSlotHead,
		map[StatType]int{},
		ClassTypeAny,
		ItemMetadata{Weight: 1},
	)
	pool := []*Item{highWeight, lowWeight}

	highPicks := 0
	lowPicks := 0
	for seed := int64(1); seed <= 500; seed++ {
		rng := rand.New(rand.NewSource(seed))
		picks := weightedSampleWithoutReplacement(pool, 1, RewardContextBoss, rng)
		if len(picks) != 1 || picks[0] == nil {
			t.Fatalf("expected one pick for seed %d", seed)
		}
		switch picks[0].ID {
		case "high_weight":
			highPicks++
		case "low_weight":
			lowPicks++
		default:
			t.Fatalf("unexpected pick %q", picks[0].ID)
		}
	}

	if highPicks <= lowPicks {
		t.Fatalf("expected weighted bias toward high weight item, high=%d low=%d", highPicks, lowPicks)
	}
}
