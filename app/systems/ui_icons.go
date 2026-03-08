package systems

import (
	"strings"

	"singlefantasy/app/assets"
	"singlefantasy/app/gamedata"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	IconSpriteSheetAssetKey = assets.TextureRavenIconSpriteSheet
	IconTileSize            = 32
	IconSheetWidth          = 512
	IconSheetHeight         = 4384
	IconSheetColumns        = IconSheetWidth / IconTileSize
	IconSheetRows           = IconSheetHeight / IconTileSize
)

type IconCell struct {
	Col int
	Row int
}

var defaultIconCell = IconCell{Col: 0, Row: 0}

var skillIconCells = map[gamedata.SkillType]IconCell{
	gamedata.SkillTypePowerStrike:   {Col: 5, Row: 79},
	gamedata.SkillTypeGuardStance:   {Col: 6, Row: 84},
	gamedata.SkillTypeBloodOath:     {Col: 3, Row: 75},
	gamedata.SkillTypeShockwaveSlam: {Col: 14, Row: 75},
	gamedata.SkillTypeQuickShot:     {Col: 1, Row: 83},
	gamedata.SkillTypeRetreatRoll:   {Col: 7, Row: 78},
	gamedata.SkillTypeFocusedAim:    {Col: 12, Row: 79},
	gamedata.SkillTypePoisonTip:     {Col: 10, Row: 85},
	gamedata.SkillTypeArcaneBolt:    {Col: 9, Row: 72},
	gamedata.SkillTypeManaShield:    {Col: 11, Row: 76},
	gamedata.SkillTypeFrostField:    {Col: 13, Row: 76},
	gamedata.SkillTypeArcaneDrain:   {Col: 8, Row: 71},
}

var effectIconCells = map[gamedata.EffectType]IconCell{
	gamedata.EffectSlow:               {Col: 10, Row: 73},
	gamedata.EffectStun:               {Col: 11, Row: 73},
	gamedata.EffectFreeze:             {Col: 12, Row: 73},
	gamedata.EffectSilence:            {Col: 13, Row: 73},
	gamedata.EffectBurn:               {Col: 0, Row: 80},
	gamedata.EffectPoison:             {Col: 10, Row: 85},
	gamedata.EffectDamageReduction:    {Col: 6, Row: 84},
	gamedata.EffectMoveSpeedReduction: {Col: 7, Row: 84},
	gamedata.EffectLifesteal:          {Col: 3, Row: 75},
	gamedata.EffectDamageBoost:        {Col: 5, Row: 79},
	gamedata.EffectMoveSpeedBoost:     {Col: 9, Row: 79},
}

var itemSlotIconCells = map[gamedata.ItemSlot]IconCell{
	gamedata.ItemSlotWeapon: {Col: 0, Row: 88},
	gamedata.ItemSlotHead:   {Col: 3, Row: 89},
	gamedata.ItemSlotChest:  {Col: 6, Row: 89},
	gamedata.ItemSlotLower:  {Col: 9, Row: 89},
}

var itemNameIconOverrides = []struct {
	Keyword string
	Cell    IconCell
}{
	{Keyword: "sword", Cell: IconCell{Col: 0, Row: 88}},
	{Keyword: "blade", Cell: IconCell{Col: 0, Row: 88}},
	{Keyword: "axe", Cell: IconCell{Col: 2, Row: 88}},
	{Keyword: "bow", Cell: IconCell{Col: 4, Row: 88}},
	{Keyword: "crossbow", Cell: IconCell{Col: 5, Row: 88}},
	{Keyword: "staff", Cell: IconCell{Col: 11, Row: 72}},
	{Keyword: "rod", Cell: IconCell{Col: 11, Row: 72}},
	{Keyword: "helmet", Cell: IconCell{Col: 3, Row: 89}},
	{Keyword: "cap", Cell: IconCell{Col: 3, Row: 89}},
	{Keyword: "hood", Cell: IconCell{Col: 4, Row: 89}},
	{Keyword: "armor", Cell: IconCell{Col: 6, Row: 89}},
	{Keyword: "mail", Cell: IconCell{Col: 7, Row: 89}},
	{Keyword: "robe", Cell: IconCell{Col: 8, Row: 89}},
	{Keyword: "pants", Cell: IconCell{Col: 9, Row: 89}},
	{Keyword: "greaves", Cell: IconCell{Col: 10, Row: 89}},
	{Keyword: "leggings", Cell: IconCell{Col: 11, Row: 89}},
}

func (cell IconCell) IsValid() bool {
	return cell.Col >= 0 && cell.Col < IconSheetColumns && cell.Row >= 0 && cell.Row < IconSheetRows
}

func GetIconSpriteSheet() rl.Texture2D {
	return assets.Get().GetTexture(IconSpriteSheetAssetKey)
}

func IconCellRect(cell IconCell) rl.Rectangle {
	if !cell.IsValid() {
		cell = defaultIconCell
	}

	x := float32(cell.Col * IconTileSize)
	y := float32(cell.Row * IconTileSize)
	return rl.NewRectangle(x, y, IconTileSize, IconTileSize)
}

func DrawIconCell(cell IconCell, destRect rl.Rectangle, tint, fallbackColor rl.Color) {
	drawTextureOrRect(GetIconSpriteSheet(), IconCellRect(cell), destRect, tint, fallbackColor)
}

func GetSkillIconCell(skillType gamedata.SkillType) IconCell {
	cell, ok := skillIconCells[skillType]
	if !ok {
		return defaultIconCell
	}
	return cell
}

func GetEffectIconCell(effectType gamedata.EffectType) IconCell {
	cell, ok := effectIconCells[effectType]
	if !ok {
		return defaultIconCell
	}
	return cell
}

func GetItemSlotIconCell(slot gamedata.ItemSlot) IconCell {
	cell, ok := itemSlotIconCells[slot]
	if !ok {
		return defaultIconCell
	}
	return cell
}

func GetItemIconCell(item *gamedata.Item) IconCell {
	if item == nil {
		return defaultIconCell
	}

	name := strings.ToLower(item.Name)
	for _, entry := range itemNameIconOverrides {
		if strings.Contains(name, entry.Keyword) {
			return entry.Cell
		}
	}

	return GetItemSlotIconCell(item.Slot)
}
