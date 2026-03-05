package gamedata

type EnemyTemplateType int

const (
	EnemyTemplateNormal EnemyTemplateType = iota
	EnemyTemplateElite
	EnemyTemplateBoss
)

type EnemyTemplate struct {
	MaxHP          int
	Damage         int
	MoveSpeed      float32
	AttackCooldown float32
	AttackRange    float32
	AggroRange     float32
	Width          float32
	Height         float32
}

var enemyTemplates = map[EnemyTemplateType]EnemyTemplate{
	EnemyTemplateNormal: {
		MaxHP:          50,
		Damage:         5,
		MoveSpeed:      100,
		AttackCooldown: 1.0,
		AttackRange:    60,
		AggroRange:     150,
		Width:          30,
		Height:         30,
	},
	EnemyTemplateElite: {
		MaxHP:          100,
		Damage:         10,
		MoveSpeed:      120,
		AttackCooldown: 1.0,
		AttackRange:    60,
		AggroRange:     150,
		Width:          30,
		Height:         30,
	},
	EnemyTemplateBoss: {
		MaxHP:          500,
		Damage:         15,
		MoveSpeed:      80,
		AttackCooldown: 1.0,
		AttackRange:    80,
		AggroRange:     1000,
		Width:          60,
		Height:         60,
	},
}

func GetEnemyTemplate(templateType EnemyTemplateType) EnemyTemplate {
	template, ok := enemyTemplates[templateType]
	if !ok {
		return enemyTemplates[EnemyTemplateNormal]
	}
	return template
}

func GetEnemyTemplateByTier(isElite bool) EnemyTemplate {
	if isElite {
		return GetEnemyTemplate(EnemyTemplateElite)
	}
	return GetEnemyTemplate(EnemyTemplateNormal)
}
