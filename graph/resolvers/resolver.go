//go:generate go run github.com/99designs/gqlgen generate

package resolvers

import (
	"errors"

	"github.com/aereal/poc-graphql-pqs-server/domain"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	characterRepo *domain.CharacterRepository
}

func Provide(characterRepo *domain.CharacterRepository) (*Resolver, error) {
	if characterRepo == nil {
		return nil, errors.New("domain.CharacterRepository is required")
	}
	r := &Resolver{characterRepo: characterRepo}
	return r, nil
}
