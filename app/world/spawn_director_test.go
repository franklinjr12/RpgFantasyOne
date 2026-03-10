package world

import (
	"testing"

	"singlefantasy/app/gamedata"
)

func TestSpawnDirectorUsesTypedEnemyRefs(t *testing.T) {
	dungeon := NewDungeon()
	for roomIndex, room := range dungeon.Rooms {
		if room == nil || room.IsBoss() {
			continue
		}
		if room.ProgressionIndex != roomIndex {
			t.Fatalf("expected progression index %d, got %d", roomIndex, room.ProgressionIndex)
		}

		for _, enemy := range room.Enemies {
			if enemy == nil {
				t.Fatalf("room %d has nil enemy ref", roomIndex)
			}
			spec := gamedata.GetEnemyArchetype(enemy.Type)
			if spec.Name == "" {
				t.Fatalf("room %d enemy has invalid archetype %d", roomIndex, enemy.Type)
			}
			if enemy.IsElite {
				mod := gamedata.GetEliteModifier(enemy.EliteModifier)
				if mod.Name == "" {
					t.Fatalf("room %d elite has invalid modifier %d", roomIndex, enemy.EliteModifier)
				}
			}
		}
	}
}

func TestSpawnDirectorIsDeterministicPerDungeonBuild(t *testing.T) {
	left := NewDungeon()
	right := NewDungeon()

	if len(left.Rooms) != len(right.Rooms) {
		t.Fatalf("room count mismatch: %d vs %d", len(left.Rooms), len(right.Rooms))
	}

	for roomIndex := range left.Rooms {
		lRoom := left.Rooms[roomIndex]
		rRoom := right.Rooms[roomIndex]
		if len(lRoom.Enemies) != len(rRoom.Enemies) {
			t.Fatalf("room %d enemy count mismatch: %d vs %d", roomIndex, len(lRoom.Enemies), len(rRoom.Enemies))
		}

		for enemyIndex := range lRoom.Enemies {
			l := lRoom.Enemies[enemyIndex]
			r := rRoom.Enemies[enemyIndex]
			if l.Type != r.Type || l.IsElite != r.IsElite || l.EliteModifier != r.EliteModifier {
				t.Fatalf("room %d enemy %d mismatch", roomIndex, enemyIndex)
			}
			if l.X != r.X || l.Y != r.Y {
				t.Fatalf("room %d enemy %d spawn mismatch", roomIndex, enemyIndex)
			}
		}
	}
}

func TestSpawnDirectorIntroducesRoleMixFromRoomTwoOnward(t *testing.T) {
	dungeon := NewDungeon()
	for roomIndex, room := range dungeon.Rooms {
		if room == nil || room.IsBoss() {
			continue
		}

		unique := map[gamedata.EnemyArchetypeType]struct{}{}
		for _, enemy := range room.Enemies {
			if enemy == nil {
				continue
			}
			unique[enemy.Type] = struct{}{}
		}

		if roomIndex >= 2 && len(room.Enemies) >= 2 && len(unique) < 2 {
			t.Fatalf("room %d expected at least two archetypes, got %d", roomIndex, len(unique))
		}
	}
}
