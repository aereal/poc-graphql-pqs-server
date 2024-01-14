package apollo

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
)

type queryList map[string]string

func New(manifest *Manifest) graphql.Cache {
	list := make(queryList)
	for _, op := range manifest.Operations {
		list[op.ID] = op.Body
	}
	return list
}

var _ graphql.Cache = (queryList)(nil)

func (l queryList) Get(_ context.Context, hash string) (any, bool) {
	query, ok := l[hash]
	return query, ok
}

func (queryList) Add(context.Context, string, any) {}

type Manifest struct {
	Format     string      `json:"format"`
	Version    int         `json:"version"`
	Operations []Operation `json:"operations"`
}

type Operation struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Body string `json:"body"`
}
