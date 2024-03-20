//go:build wireinject

package main

import (
	"context"

	"github.com/aereal/poc-graphql-pqs-server/config"
	"github.com/aereal/poc-graphql-pqs-server/domain"
	"github.com/aereal/poc-graphql-pqs-server/graph"
	"github.com/aereal/poc-graphql-pqs-server/graph/loaders"
	"github.com/aereal/poc-graphql-pqs-server/graph/persistedquery/apollo"
	"github.com/aereal/poc-graphql-pqs-server/graph/resolvers"
	"github.com/aereal/poc-graphql-pqs-server/infra"
	"github.com/aereal/poc-graphql-pqs-server/otel/otelinstrument"
	"github.com/aereal/poc-graphql-pqs-server/web"
	"github.com/google/wire"
)

func initialize(ctx context.Context) (*app, error) {
	wire.Build(
		provideApp,
		web.ProvideServer,
		config.EnvironmentProvider,
		graph.ProvideExecutableSchema,
		resolvers.Provide,
		loaders.Provide,
		domain.ProvideCharacterRepository,
		apollo.ProvideQueryCacheFromManifest,
		apollo.ProvideManifestFromPath,
		infra.ProvideDB,
		otelinstrument.Provider,
	)
	return &app{}, nil
}
