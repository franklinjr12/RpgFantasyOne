package gamedata

type SkillType int

const (
	SkillTypePowerStrike SkillType = iota
	SkillTypeGuardStance
	SkillTypeBloodOath
	SkillTypeShockwaveSlam
	SkillTypeQuickShot
	SkillTypeRetreatRoll
	SkillTypeFocusedAim
	SkillTypePoisonTip
	SkillTypeArcaneBolt
	SkillTypeManaShield
	SkillTypeFrostField
	SkillTypeArcaneDrain
)

type Skill struct {
	Type            SkillType
	Name            string
	Cooldown        float32
	CurrentCooldown float32
	ManaCost        int
	Targeting       TargetingSpec
	Delivery        DeliverySpec
	DamageSpec      *DamageSpec
	Effects         []EffectSpec
}

func NewSkill(skillType SkillType) *Skill {
	switch skillType {
	case SkillTypePowerStrike:
		return &Skill{
			Type:     SkillTypePowerStrike,
			Name:     "Power Strike",
			Cooldown: 7.0,
			ManaCost: 0,
			Targeting: TargetingSpec{
				Type:       TargetEnemy,
				Range:      60,
				MaxTargets: 1,
			},
			Delivery: DeliverySpec{
				Type: DeliveryInstant,
			},
			DamageSpec: &DamageSpec{
				Base:       30,
				Scaling:    map[StatType]float32{StatTypeSTR: 1.5},
				DamageType: DamagePhysical,
				CritChance: 0.15,
				CritMult:   1.5,
			},
		}
	case SkillTypeGuardStance:
		return &Skill{
			Type:     SkillTypeGuardStance,
			Name:     "Guard Stance",
			Cooldown: 12.0,
			ManaCost: 0,
			Targeting: TargetingSpec{
				Type: TargetSelf,
			},
			Delivery: DeliverySpec{
				Type: DeliveryInstant,
			},
			Effects: []EffectSpec{
				{Type: int(EffectDamageReduction), Duration: 4.0, Magnitude: 0.4},
				{Type: int(EffectMoveSpeedReduction), Duration: 4.0, Magnitude: 0.3},
			},
		}
	case SkillTypeBloodOath:
		return &Skill{
			Type:     SkillTypeBloodOath,
			Name:     "Blood Oath",
			Cooldown: 12.0,
			ManaCost: 0,
			Targeting: TargetingSpec{
				Type: TargetSelf,
			},
			Delivery: DeliverySpec{
				Type: DeliveryInstant,
			},
			Effects: []EffectSpec{
				{Type: int(EffectLifesteal), Duration: 5.0, Magnitude: 0.3},
			},
		}
	case SkillTypeShockwaveSlam:
		return &Skill{
			Type:     SkillTypeShockwaveSlam,
			Name:     "Shockwave Slam",
			Cooldown: 17.0,
			ManaCost: 0,
			Targeting: TargetingSpec{
				Type:       TargetArea,
				Radius:     100,
				MaxTargets: 10,
			},
			Delivery: DeliverySpec{
				Type: DeliveryInstant,
			},
			DamageSpec: &DamageSpec{
				Base:       25,
				Scaling:    map[StatType]float32{StatTypeSTR: 1.0},
				DamageType: DamagePhysical,
			},
			Effects: []EffectSpec{
				{Type: int(EffectSlow), Duration: 2.0, Magnitude: 0.3},
			},
		}
	case SkillTypeQuickShot:
		return &Skill{
			Type:     SkillTypeQuickShot,
			Name:     "Quick Shot",
			Cooldown: 6.0,
			ManaCost: 0,
			Targeting: TargetingSpec{
				Type:       TargetEnemy,
				Range:      250,
				MaxTargets: 1,
			},
			Delivery: DeliverySpec{
				Type:  DeliveryProjectile,
				Speed: 500,
			},
			DamageSpec: &DamageSpec{
				Base:       20,
				Scaling:    map[StatType]float32{StatTypeDEX: 1.2},
				DamageType: DamagePhysical,
			},
		}
	case SkillTypeRetreatRoll:
		return &Skill{
			Type:     SkillTypeRetreatRoll,
			Name:     "Retreat Roll",
			Cooldown: 11.0,
			ManaCost: 0,
			Targeting: TargetingSpec{
				Type: TargetSelf,
			},
			Delivery: DeliverySpec{
				Type: DeliveryInstant,
			},
			Effects: []EffectSpec{
				{Type: int(EffectMoveSpeedBoost), Duration: 2.0, Magnitude: 0.5},
			},
		}
	case SkillTypeFocusedAim:
		return &Skill{
			Type:     SkillTypeFocusedAim,
			Name:     "Focused Aim",
			Cooldown: 12.0,
			ManaCost: 0,
			Targeting: TargetingSpec{
				Type: TargetSelf,
			},
			Delivery: DeliverySpec{
				Type: DeliveryInstant,
			},
			Effects: []EffectSpec{
				{Type: int(EffectDamageBoost), Duration: 5.0, Magnitude: 0.5},
				{Type: int(EffectMoveSpeedReduction), Duration: 5.0, Magnitude: 0.4},
			},
		}
	case SkillTypePoisonTip:
		return &Skill{
			Type:     SkillTypePoisonTip,
			Name:     "Poison Tip",
			Cooldown: 16.0,
			ManaCost: 0,
			Targeting: TargetingSpec{
				Type:       TargetEnemy,
				Range:      250,
				MaxTargets: 1,
			},
			Delivery: DeliverySpec{
				Type:  DeliveryProjectile,
				Speed: 400,
			},
			DamageSpec: &DamageSpec{
				Base:       15,
				Scaling:    map[StatType]float32{StatTypeDEX: 0.8},
				DamageType: DamagePhysical,
			},
			Effects: []EffectSpec{
				{Type: int(EffectPoison), Duration: 5.0, Magnitude: 3.0, TickRate: 1.0},
			},
		}
	case SkillTypeArcaneBolt:
		return &Skill{
			Type:     SkillTypeArcaneBolt,
			Name:     "Arcane Bolt",
			Cooldown: 7.0,
			ManaCost: 15,
			Targeting: TargetingSpec{
				Type:       TargetEnemy,
				Range:      200,
				MaxTargets: 1,
			},
			Delivery: DeliverySpec{
				Type:  DeliveryProjectile,
				Speed: 450,
			},
			DamageSpec: &DamageSpec{
				Base:       40,
				Scaling:    map[StatType]float32{StatTypeINT: 1.8},
				DamageType: DamageMagical,
			},
		}
	case SkillTypeManaShield:
		return &Skill{
			Type:     SkillTypeManaShield,
			Name:     "Mana Shield",
			Cooldown: 12.0,
			ManaCost: 30,
			Targeting: TargetingSpec{
				Type: TargetSelf,
			},
			Delivery: DeliverySpec{
				Type: DeliveryInstant,
			},
		}
	case SkillTypeFrostField:
		return &Skill{
			Type:     SkillTypeFrostField,
			Name:     "Frost Field",
			Cooldown: 15.0,
			ManaCost: 25,
			Targeting: TargetingSpec{
				Type:       TargetArea,
				Radius:     120,
				MaxTargets: 10,
			},
			Delivery: DeliverySpec{
				Type: DeliveryInstant,
			},
			Effects: []EffectSpec{
				{Type: int(EffectSlow), Duration: 4.0, Magnitude: 0.5},
			},
		}
	case SkillTypeArcaneDrain:
		return &Skill{
			Type:     SkillTypeArcaneDrain,
			Name:     "Arcane Drain",
			Cooldown: 18.0,
			ManaCost: 20,
			Targeting: TargetingSpec{
				Type:       TargetArea,
				Radius:     80,
				MaxTargets: 10,
			},
			Delivery: DeliverySpec{
				Type: DeliveryInstant,
			},
			DamageSpec: &DamageSpec{
				Base:       20,
				Scaling:    map[StatType]float32{StatTypeINT: 1.0},
				DamageType: DamageMagical,
			},
		}
	default:
		return nil
	}
}

func GetClassSkills(classType ClassType) []*Skill {
	switch classType {
	case ClassTypeMelee:
		return []*Skill{
			NewSkill(SkillTypePowerStrike),
			NewSkill(SkillTypeGuardStance),
			NewSkill(SkillTypeBloodOath),
			NewSkill(SkillTypeShockwaveSlam),
		}
	case ClassTypeRanged:
		return []*Skill{
			NewSkill(SkillTypeQuickShot),
			NewSkill(SkillTypeRetreatRoll),
			NewSkill(SkillTypeFocusedAim),
			NewSkill(SkillTypePoisonTip),
		}
	case ClassTypeCaster:
		return []*Skill{
			NewSkill(SkillTypeArcaneBolt),
			NewSkill(SkillTypeManaShield),
			NewSkill(SkillTypeFrostField),
			NewSkill(SkillTypeArcaneDrain),
		}
	default:
		return []*Skill{}
	}
}

func (s *Skill) Update(deltaTime float32) {
	if s.CurrentCooldown > 0 {
		s.CurrentCooldown -= deltaTime
		if s.CurrentCooldown < 0 {
			s.CurrentCooldown = 0
		}
	}
}

func (s *Skill) CanUse() bool {
	return s.CurrentCooldown <= 0
}

func (s *Skill) Use() {
	s.CurrentCooldown = s.Cooldown
}
