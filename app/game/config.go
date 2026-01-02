package game

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	BiomeCount    = 1
	DungeonLength = 5
	BossRoomCount = 1
	TotalRooms    = DungeonLength + BossRoomCount
	ClassCount    = 3
	BossTypeCount = 1
)

const (
	WindowWidth  = 1600
	WindowHeight = 900
	TargetFPS    = 60
)

const (
	TerrainColorNormal = 0x808080FF
	TerrainColorBoss   = 0x404040FF
)

const (
	PlayerColor     = 0x0000FFFF
	EnemyColor      = 0xFF0000FF
	EliteColor      = 0xFF8800FF
	BossColor       = 0x8800FFFF
	ProjectileColor = 0xFFFF00FF
)

const (
	RoomMinWidth   = 400
	RoomMaxWidth   = 600
	RoomMinHeight  = 300
	RoomMaxHeight  = 450
	BossRoomWidth  = 800
	BossRoomHeight = 600
)

const (
	EnemiesPerRoomMin = 3
	EnemiesPerRoomMax = 5
)

var TerrainColorNormalRGBA = rl.NewColor(128, 128, 128, 255)
var TerrainColorBossRGBA = rl.NewColor(64, 64, 64, 255)
var PlayerColorRGBA = rl.NewColor(0, 0, 255, 255)
var EnemyColorRGBA = rl.NewColor(255, 0, 0, 255)
var EliteColorRGBA = rl.NewColor(255, 136, 0, 255)
var BossColorRGBA = rl.NewColor(136, 0, 255, 255)
var ProjectileColorRGBA = rl.NewColor(255, 255, 0, 255)
