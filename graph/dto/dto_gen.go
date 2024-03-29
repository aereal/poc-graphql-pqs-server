// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package dto

import (
	"github.com/aereal/poc-graphql-pqs-server/domain"
)

type CharacterConnection struct {
	Nodes    []*domain.Character `json:"nodes"`
	PageInfo *PageInfo           `json:"pageInfo"`
}

type CharactersOrder struct {
	Field     domain.CharacterOrderField `json:"field"`
	Direction domain.OrderDirection      `json:"direction"`
}

type PageInfo struct {
	HasNext   bool    `json:"hasNext"`
	EndCursor *string `json:"endCursor,omitempty"`
}

type Query struct {
}
