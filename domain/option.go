package domain

import "github.com/jmoiron/sqlx"

type DBOption interface {
	CharacterRepositoryOption
}

type CharacterRepositoryOption interface {
	applyCharacterRepositoryOption(*CharacterRepository)
}

type withDBOpt struct{ db *sqlx.DB }

func (o *withDBOpt) applyCharacterRepositoryOption(r *CharacterRepository) { r.db = o.db }

func WithDB(db *sqlx.DB) DBOption { return &withDBOpt{db} }
