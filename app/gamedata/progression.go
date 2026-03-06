package gamedata

const (
	BasePlayerHP               = 100
	BasePlayerMana             = 50
	BasePlayerMoveSpeed        = float32(200)
	BaseMeleeAutoAttackDamage  = 10
	BaseRangedAutoAttackDamage = 10
	BaseCasterAutoAttackDamage = 15
	LevelUpStatPoints          = 3
	LevelUpGrowthStatPoints    = 1
	XPPerLevel                 = 100
)

func XPToNextLevel(level int) int {
	if level < 1 {
		level = 1
	}
	return level * XPPerLevel
}
