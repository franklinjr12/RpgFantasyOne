package gameobjects

import "testing"

func TestBossEnrageTriggersOnceAtThreshold(t *testing.T) {
	boss := NewBoss(0, 0, "forest")
	threshold := int(float32(boss.MaxHP) * boss.Config.Enrage.ThresholdHPPercent)
	if threshold < 1 {
		threshold = 1
	}

	boss.HP = threshold + 10
	boss.Update(0.016, 100, 0)
	if boss.EnrageTriggered {
		t.Fatalf("expected boss to stay normal above threshold")
	}

	boss.HP = threshold
	boss.Update(0.016, 100, 0)
	if !boss.EnrageTriggered {
		t.Fatalf("expected enrage at threshold")
	}
	if boss.Phase != BossPhaseEnraged {
		t.Fatalf("expected enraged phase, got %s", boss.Phase.String())
	}

	damageAfterEnrage := boss.Damage
	moveAfterEnrage := boss.MoveSpeed
	boss.Update(0.016, 100, 0)
	if boss.Damage != damageAfterEnrage || boss.MoveSpeed != moveAfterEnrage {
		t.Fatalf("expected enrage multipliers to apply only once")
	}
}

func TestBossHeavyTelegraphResolvesAtSnapshottedLocation(t *testing.T) {
	boss := NewBoss(0, 0, "forest")
	boss.HeavyCooldownRemaining = 0
	boss.AreaCooldownRemaining = boss.Config.AreaDenial.Cooldown

	initialTargetX := float32(120)
	initialTargetY := float32(90)
	boss.Update(0.016, initialTargetX, initialTargetY)
	telegraph, ok := boss.ActiveHeavyTelegraph()
	if !ok {
		t.Fatalf("expected heavy telegraph to start")
	}
	if telegraph.X != initialTargetX || telegraph.Y != initialTargetY {
		t.Fatalf("expected telegraph to capture initial target position")
	}

	boss.Update(boss.Config.HeavyAttack.TelegraphDuration+0.02, 420, 420)
	events := boss.ConsumeDamageEvents()
	if len(events) == 0 {
		t.Fatalf("expected heavy resolve event")
	}

	foundHeavy := false
	for _, event := range events {
		if event.Type != BossDamageEventHeavy {
			continue
		}
		foundHeavy = true
		if event.X != initialTargetX || event.Y != initialTargetY {
			t.Fatalf("expected heavy resolve at snapshotted location, got (%.2f,%.2f)", event.X, event.Y)
		}
	}
	if !foundHeavy {
		t.Fatalf("expected heavy event in pending damage events")
	}
}

func TestBossAreaDenialTicksAndCleansUp(t *testing.T) {
	boss := NewBoss(0, 0, "forest")
	boss.HeavyCooldownRemaining = 999
	boss.AreaCooldownRemaining = 0

	boss.Update(0.016, 140, 110)
	if boss.ActiveZoneCount() < boss.Config.AreaDenial.ZoneCount {
		t.Fatalf("expected spawned area denial zones")
	}

	totalAreaEvents := 0
	simDuration := boss.Config.AreaDenial.WarningDuration + boss.Config.AreaDenial.ActiveDuration + 0.5
	step := boss.Config.AreaDenial.TickRate / 2
	if step <= 0 {
		step = 0.1
	}

	for elapsed := float32(0); elapsed < simDuration; elapsed += step {
		boss.Update(step, 140, 110)
		for _, event := range boss.ConsumeDamageEvents() {
			if event.Type == BossDamageEventArea {
				totalAreaEvents++
			}
		}
	}

	if totalAreaEvents == 0 {
		t.Fatalf("expected area denial damage tick events")
	}
	if boss.ActiveZoneCount() != 0 {
		t.Fatalf("expected zones to expire, got %d active", boss.ActiveZoneCount())
	}
}

func TestBossEnrageIncreasesAreaZoneCount(t *testing.T) {
	boss := NewBoss(0, 0, "forest")
	threshold := int(float32(boss.MaxHP) * boss.Config.Enrage.ThresholdHPPercent)
	if threshold < 1 {
		threshold = 1
	}
	boss.HP = threshold
	boss.Update(0.016, 100, 0)
	if !boss.EnrageTriggered {
		t.Fatalf("expected enrage state")
	}

	boss.HeavyCooldownRemaining = 999
	boss.AreaCooldownRemaining = 0
	boss.Update(0.016, 100, 0)

	expected := boss.Config.AreaDenial.ZoneCount + boss.Config.Enrage.ZoneCountBonus
	if expected < 1 {
		expected = 1
	}
	if boss.ActiveZoneCount() != expected {
		t.Fatalf("expected %d zones while enraged, got %d", expected, boss.ActiveZoneCount())
	}
}

func TestBossProvokedOutsideAggroRangeStillChases(t *testing.T) {
	boss := NewBoss(0, 0, "forest")
	boss.AggroRange = 20

	playerX := float32(500)
	playerY := float32(0)

	boss.Update(0.016, playerX, playerY)
	if boss.State != EnemyStateIdle {
		t.Fatalf("expected idle while outside aggro and not provoked, got %d", boss.State)
	}

	boss.TakeDamage(1)
	if !boss.Provoked {
		t.Fatalf("expected boss to be provoked after taking damage")
	}

	boss.Update(0.016, playerX, playerY)
	if boss.State != EnemyStateChasing {
		t.Fatalf("expected provoked boss to chase outside aggro range, got %d", boss.State)
	}
}
