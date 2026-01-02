package gamedata

type SkillType int

const (
	SkillTypeDash SkillType = iota
	SkillTypeMultiShot
	SkillTypeManaShield
)

type Skill struct {
	Type            SkillType
	Name            string
	Cooldown        float32
	CurrentCooldown float32
	ManaCost        int
}

func NewSkill(skillType SkillType) *Skill {
	switch skillType {
	case SkillTypeDash:
		return &Skill{
			Type:     SkillTypeDash,
			Name:     "Dash",
			Cooldown: 3.0,
			ManaCost: 0,
		}
	case SkillTypeMultiShot:
		return &Skill{
			Type:     SkillTypeMultiShot,
			Name:     "Multi-Shot",
			Cooldown: 5.0,
			ManaCost: 0,
		}
	case SkillTypeManaShield:
		return &Skill{
			Type:     SkillTypeManaShield,
			Name:     "Mana Shield",
			Cooldown: 8.0,
			ManaCost: 20,
		}
	default:
		return nil
	}
}

func GetClassSkills(classType ClassType) []*Skill {
	switch classType {
	case ClassTypeMelee:
		return []*Skill{NewSkill(SkillTypeDash)}
	case ClassTypeRanged:
		return []*Skill{NewSkill(SkillTypeMultiShot)}
	case ClassTypeCaster:
		return []*Skill{NewSkill(SkillTypeManaShield)}
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
