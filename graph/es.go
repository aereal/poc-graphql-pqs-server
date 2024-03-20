package graph

import (
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/aereal/poc-graphql-pqs-server/graph/executableschema"
	"github.com/aereal/poc-graphql-pqs-server/graph/resolvers"
)

func ProvideExecutableSchema(resolversRoot *resolvers.Resolver) (graphql.ExecutableSchema, error) {
	if resolversRoot == nil {
		return nil, errors.New("resolvers.Resolver is required")
	}
	cfg := executableschema.Config{Resolvers: resolversRoot}
	es := executableschema.NewExecutableSchema(cfg)
	return es, nil
}
