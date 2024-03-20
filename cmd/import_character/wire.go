//go:build wireinject

package main

import (
	"context"

	"github.com/aereal/poc-graphql-pqs-server/config"
	"github.com/aereal/poc-graphql-pqs-server/infra"
	"github.com/aereal/poc-graphql-pqs-server/otel/otelinstrument"
	"github.com/google/wire"
)

func initialize(context.Context) (*app, error) {
	wire.Build(
		provideApp,
		config.EnvironmentProvider,
		infra.ProvideDB,
		otelinstrument.Provider,
	)
	return &app{}, nil
}
