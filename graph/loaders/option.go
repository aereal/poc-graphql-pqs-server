package loaders

import "github.com/aereal/poc-graphql-pqs-server/domain"

type config struct {
	characterRepo *domain.CharacterRepository
}

type Option func(*config)

func WithCharacterRepository(r *domain.CharacterRepository) Option {
	return func(c *config) { c.characterRepo = r }
}
