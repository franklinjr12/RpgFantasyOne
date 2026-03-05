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
)

const (
	FixedDeltaTime float32 = 1.0 / 60.0
	MaxFrameTime   float32 = 0.25
	MaxUpdateSteps int     = 8
)

const (
	DebugToggleKey int32 = rl.KeyF3
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

const (
	PlayerMoveTargetStopDistance      float32 = 6
	PlayerMoveTargetSlowRadius        float32 = 48
	PlayerMoveAcceleration            float32 = 1600
	PlayerMoveDeceleration            float32 = 1900
	PlayerFacingDeadzone              float32 = 8
	PlayerVelocitySnapThreshold       float32 = 0.5
	PlayerCollisionBlockedRatio       float32 = 0.5
	PlayerBlockedAxisKnockbackDamping float32 = 0.2
	PlayerKnockbackImpulse            float32 = 110
	PlayerKnockbackDecayPerSecond     float32 = 520
	PlayerHitFlashDuration            float32 = 0.12
	PlayerHurtIFrameDuration          float32 = 0.22
	MeleeAttackWindup                 float32 = 0.12
	MeleeAttackRecover                float32 = 0.1
	RangedAttackWindup                float32 = 0.08
	RangedAttackRecover               float32 = 0.06
	CasterAttackWindup                float32 = 0.1
	CasterAttackRecover               float32 = 0.08
	MeleeAttackHitRangeBuffer         float32 = 8
	AutoAttackProjectileSpeed         float32 = 400
)

var TerrainColorNormalRGBA = rl.NewColor(128, 128, 128, 255)
var TerrainColorBossRGBA = rl.NewColor(64, 64, 64, 255)
var PlayerColorRGBA = rl.NewColor(0, 0, 255, 255)
var EnemyColorRGBA = rl.NewColor(255, 0, 0, 255)
var EliteColorRGBA = rl.NewColor(255, 136, 0, 255)
var BossColorRGBA = rl.NewColor(136, 0, 255, 255)
var ProjectileColorRGBA = rl.NewColor(255, 255, 0, 255)
