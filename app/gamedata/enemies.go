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

type EnemyAttackMode int

const (
	EnemyAttackMelee EnemyAttackMode = iota
	EnemyAttackProjectile
	EnemyAttackCasterAOE
)

type EnemyArchetypeType int

const (
	EnemyArchetypeRaider EnemyArchetypeType = iota
	EnemyArchetypePikeman
	EnemyArchetypeArcher
	EnemyArchetypeHexCaller
	EnemyArchetypeBrute
	EnemyArchetypeSwarmling
)

type EnemyArchetype struct {
	Type               EnemyArchetypeType
	Name               string
	Role               string
	MaxHP              int
	Damage             int
	MoveSpeed          float32
	AttackCooldown     float32
	AttackRange        float32
	AggroRange         float32
	PreferredRange     float32
	RetreatRange       float32
	Width              float32
	Height             float32
	AttackMode         EnemyAttackMode
	ProjectileSpeed    float32
	ProjectileRadius   float32
	ProjectileLifetime float32
	DamageType         DamageType
	OnHitEffects       []EffectSpec
	XPReward           int
	ThreatValue        int
}

type EliteModifierType int

const (
	EliteModifierScorching EliteModifierType = iota
	EliteModifierCrippling
)

type EliteModifier struct {
	Type          EliteModifierType
	Name          string
	HPMultiplier  float32
	DmgMultiplier float32
	OnHitEffects  []EffectSpec
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

var enemyArchetypeOrder = []EnemyArchetypeType{
	EnemyArchetypeRaider,
	EnemyArchetypePikeman,
	EnemyArchetypeArcher,
	EnemyArchetypeHexCaller,
	EnemyArchetypeBrute,
	EnemyArchetypeSwarmling,
}

var enemyArchetypes = map[EnemyArchetypeType]EnemyArchetype{
	EnemyArchetypeRaider: {
		Type:           EnemyArchetypeRaider,
		Name:           "Raider",
		Role:           "Melee Chaser",
		MaxHP:          58,
		Damage:         8,
		MoveSpeed:      128,
		AttackCooldown: 1.0,
		AttackRange:    56,
		AggroRange:     320,
		PreferredRange: 46,
		RetreatRange:   0,
		Width:          30,
		Height:         30,
		AttackMode:     EnemyAttackMelee,
		DamageType:     DamagePhysical,
		XPReward:       20,
		ThreatValue:    11,
	},
	EnemyArchetypePikeman: {
		Type:           EnemyArchetypePikeman,
		Name:           "Pikeman",
		Role:           "Melee Chaser",
		MaxHP:          86,
		Damage:         10,
		MoveSpeed:      95,
		AttackCooldown: 1.25,
		AttackRange:    66,
		AggroRange:     300,
		PreferredRange: 56,
		RetreatRange:   0,
		Width:          32,
		Height:         34,
		AttackMode:     EnemyAttackMelee,
		DamageType:     DamagePhysical,
		XPReward:       22,
		ThreatValue:    14,
	},
	EnemyArchetypeArcher: {
		Type:               EnemyArchetypeArcher,
		Name:               "Archer",
		Role:               "Ranged",
		MaxHP:              68,
		Damage:             9,
		MoveSpeed:          104,
		AttackCooldown:     1.5,
		AttackRange:        360,
		AggroRange:         390,
		PreferredRange:     280,
		RetreatRange:       140,
		Width:              30,
		Height:             30,
		AttackMode:         EnemyAttackProjectile,
		ProjectileSpeed:    270,
		ProjectileRadius:   6,
		ProjectileLifetime: 2.2,
		DamageType:         DamagePhysical,
		XPReward:           24,
		ThreatValue:        16,
	},
	EnemyArchetypeHexCaller: {
		Type:           EnemyArchetypeHexCaller,
		Name:           "Hex Caller",
		Role:           "Caster",
		MaxHP:          72,
		Damage:         7,
		MoveSpeed:      88,
		AttackCooldown: 2.4,
		AttackRange:    300,
		AggroRange:     360,
		PreferredRange: 230,
		RetreatRange:   160,
		Width:          30,
		Height:         32,
		AttackMode:     EnemyAttackCasterAOE,
		DamageType:     DamageMagical,
		OnHitEffects: []EffectSpec{
			{
				Type:      EffectSlow,
				Duration:  2.0,
				Magnitude: 0.25,
			},
		},
		XPReward:    26,
		ThreatValue: 18,
	},
	EnemyArchetypeBrute: {
		Type:           EnemyArchetypeBrute,
		Name:           "Brute",
		Role:           "Tank Bruiser",
		MaxHP:          136,
		Damage:         16,
		MoveSpeed:      72,
		AttackCooldown: 1.9,
		AttackRange:    64,
		AggroRange:     300,
		PreferredRange: 52,
		RetreatRange:   0,
		Width:          40,
		Height:         40,
		AttackMode:     EnemyAttackMelee,
		DamageType:     DamagePhysical,
		XPReward:       32,
		ThreatValue:    24,
	},
	EnemyArchetypeSwarmling: {
		Type:           EnemyArchetypeSwarmling,
		Name:           "Swarmling",
		Role:           "Swarmer",
		MaxHP:          30,
		Damage:         4,
		MoveSpeed:      170,
		AttackCooldown: 0.8,
		AttackRange:    44,
		AggroRange:     320,
		PreferredRange: 34,
		RetreatRange:   0,
		Width:          22,
		Height:         22,
		AttackMode:     EnemyAttackMelee,
		DamageType:     DamagePhysical,
		XPReward:       10,
		ThreatValue:    6,
	},
}

var eliteModifierOrder = []EliteModifierType{
	EliteModifierScorching,
	EliteModifierCrippling,
}

var eliteModifiers = map[EliteModifierType]EliteModifier{
	EliteModifierScorching: {
		Type:          EliteModifierScorching,
		Name:          "Scorching",
		HPMultiplier:  1.45,
		DmgMultiplier: 1.25,
		OnHitEffects: []EffectSpec{
			{
				Type:      EffectBurn,
				Duration:  3.0,
				Magnitude: 3.0,
				TickRate:  1.0,
			},
		},
	},
	EliteModifierCrippling: {
		Type:          EliteModifierCrippling,
		Name:          "Crippling",
		HPMultiplier:  1.5,
		DmgMultiplier: 1.15,
		OnHitEffects: []EffectSpec{
			{
				Type:      EffectMoveSpeedReduction,
				Duration:  2.4,
				Magnitude: 0.25,
			},
		},
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

func GetEnemyArchetype(archetype EnemyArchetypeType) EnemyArchetype {
	value, ok := enemyArchetypes[archetype]
	if !ok {
		return enemyArchetypes[EnemyArchetypeRaider]
	}
	return value
}

func EnemyArchetypeTypes() []EnemyArchetypeType {
	out := make([]EnemyArchetypeType, len(enemyArchetypeOrder))
	copy(out, enemyArchetypeOrder)
	return out
}

func GetEliteModifier(modifierType EliteModifierType) EliteModifier {
	value, ok := eliteModifiers[modifierType]
	if !ok {
		return eliteModifiers[EliteModifierScorching]
	}
	return value
}

func EliteModifierTypes() []EliteModifierType {
	out := make([]EliteModifierType, len(eliteModifierOrder))
	copy(out, eliteModifierOrder)
	return out
}
