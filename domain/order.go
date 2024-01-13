package domain

type OrderDirection string

const (
	OrderDirectionAsc  OrderDirection = "ASC"
	OrderDirectionDesc OrderDirection = "DESC"
)

type CharacterOrderField string

const (
	CharacterOrderFieldHealth             CharacterOrderField = "HEALTH"
	CharacterOrderFieldAttack             CharacterOrderField = "ATTACK"
	CharacterOrderFieldDefence            CharacterOrderField = "DEFENCE"
	CharacterOrderFieldElementEnergy      CharacterOrderField = "ELEMENT_ENERGY"
	CharacterOrderFieldUniqueAbilityScore CharacterOrderField = "UNIQUE_ABILITY_SCORE"
)

func (f CharacterOrderField) column() string {
	switch f {
	case CharacterOrderFieldHealth:
		return "health"
	case CharacterOrderFieldAttack:
		return "attack"
	case CharacterOrderFieldDefence:
		return "defence"
	case CharacterOrderFieldElementEnergy:
		return "element_energy"
	case CharacterOrderFieldUniqueAbilityScore:
		return "unique_ability_score"
	default:
		return ""
	}
}
