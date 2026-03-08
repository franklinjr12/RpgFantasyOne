package systems

import (
	"singlefantasy/app/gamedata"
	"testing"
)

func TestIconCellRectUsesExpectedAtlasCoordinates(t *testing.T) {
	origin := IconCellRect(IconCell{Col: 0, Row: 0})
	if origin.X != 0 || origin.Y != 0 || origin.Width != IconTileSize || origin.Height != IconTileSize {
		t.Fatalf("unexpected origin rect: %+v", origin)
	}

	middle := IconCellRect(IconCell{Col: 7, Row: 44})
	if middle.X != 224 || middle.Y != 1408 || middle.Width != IconTileSize || middle.Height != IconTileSize {
		t.Fatalf("unexpected middle rect: %+v", middle)
	}

	edge := IconCellRect(IconCell{Col: IconSheetColumns - 1, Row: IconSheetRows - 1})
	if edge.X != 480 || edge.Y != 4352 || edge.Width != IconTileSize || edge.Height != IconTileSize {
		t.Fatalf("unexpected edge rect: %+v", edge)
	}
}

func TestIconCellRectFallsBackForInvalidCell(t *testing.T) {
	got := IconCellRect(IconCell{Col: -1, Row: IconSheetRows + 5})
	want := IconCellRect(defaultIconCell)
	if got != want {
		t.Fatalf("expected fallback rect %+v, got %+v", want, got)
	}
}

func TestEveryClassSkillHasValidIconCell(t *testing.T) {
	classTypes := []gamedata.ClassType{
		gamedata.ClassTypeMelee,
		gamedata.ClassTypeRanged,
		gamedata.ClassTypeCaster,
	}

	for _, classType := range classTypes {
		skills := gamedata.GetClassSkillData(classType)
		for _, skill := range skills {
			if skill == nil {
				t.Fatalf("nil skill found for class %v", classType)
			}
			cell := GetSkillIconCell(skill.Type)
			if !cell.IsValid() {
				t.Fatalf("invalid icon cell for skill %s (%v): %+v", skill.Name, skill.Type, cell)
			}
		}
	}
}

func TestGetItemIconCellFallbackAndOverrides(t *testing.T) {
	if cell := GetItemIconCell(nil); cell != defaultIconCell {
		t.Fatalf("expected nil item to fallback to default cell, got %+v", cell)
	}

	invalidSlot := gamedata.NewItem("Unknown Relic", "", gamedata.ItemSlot(99), map[gamedata.StatType]int{}, gamedata.ClassTypeMelee)
	if cell := GetItemIconCell(invalidSlot); cell != defaultIconCell {
		t.Fatalf("expected invalid slot item to fallback to default cell, got %+v", cell)
	}

	head := gamedata.NewItem("Plain Headpiece", "", gamedata.ItemSlotHead, map[gamedata.StatType]int{}, gamedata.ClassTypeMelee)
	if cell := GetItemIconCell(head); cell != GetItemSlotIconCell(gamedata.ItemSlotHead) {
		t.Fatalf("expected head item to use head slot icon, got %+v", cell)
	}

	crossbow := gamedata.NewItem("Crossbow", "", gamedata.ItemSlotWeapon, map[gamedata.StatType]int{}, gamedata.ClassTypeRanged)
	if cell := GetItemIconCell(crossbow); cell == GetItemSlotIconCell(gamedata.ItemSlotWeapon) {
		t.Fatalf("expected crossbow to use name override cell, got slot default %+v", cell)
	}
}
