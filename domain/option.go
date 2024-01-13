package domain

import "github.com/jmoiron/sqlx"

type DBOption interface {
	CharacterRepositoryOption
}

type CharacterRepositoryOption interface {
	applyCharacterRepositoryOption(*CharacterRepository)
}

type LimitOption interface {
	SearchCharactersOption
}

type SearchCharactersOption interface {
	applySearchCharactersOption(*searchCharactersArgs)
}

type withDBOpt struct{ db *sqlx.DB }

func (o *withDBOpt) applyCharacterRepositoryOption(r *CharacterRepository) { r.db = o.db }

func WithDB(db *sqlx.DB) DBOption { return &withDBOpt{db} }

type withLimitOpt struct{ limit uint }

func (o *withLimitOpt) applySearchCharactersOption(args *searchCharactersArgs) {
	args.limit = o.limit
}

func WithLimit(limit uint) LimitOption { return &withLimitOpt{limit: limit} }

type withCharacterOrderOpt struct {
	field     CharacterOrderField
	direction OrderDirection
}

func (o *withCharacterOrderOpt) applySearchCharactersOption(args *searchCharactersArgs) {
	args.orderDirection = o.direction
	args.orderField = o.field
}

func WithCharacterOrder(field CharacterOrderField, direction OrderDirection) SearchCharactersOption {
	return &withCharacterOrderOpt{field: field, direction: direction}
}

type withCharacterFilterCriteriaOpt struct{ criteria *CharacterFilterCriteria }

func (o *withCharacterFilterCriteriaOpt) applySearchCharactersOption(args *searchCharactersArgs) {
	args.criteria = o.criteria
}

func WithCharacterFilterCriteria(criteria *CharacterFilterCriteria) SearchCharactersOption {
	return &withCharacterFilterCriteriaOpt{criteria}
}
