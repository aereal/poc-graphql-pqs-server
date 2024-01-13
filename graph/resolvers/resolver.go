package resolvers

import "github.com/aereal/poc-graphql-pqs-server/domain"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	characterRepo *domain.CharacterRepository
}

type Option func(*Resolver)

func WithCharacterRepository(repo *domain.CharacterRepository) Option {
	return func(r *Resolver) { r.characterRepo = repo }
}

func New(opts ...Option) *Resolver {
	r := &Resolver{}
	for _, o := range opts {
		o(r)
	}
	return r
}
