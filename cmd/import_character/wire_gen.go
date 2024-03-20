// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"context"
	"github.com/aereal/poc-graphql-pqs-server/config"
	"github.com/aereal/poc-graphql-pqs-server/infra"
	"github.com/aereal/poc-graphql-pqs-server/otel/otelinstrument"
)

import (
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

// Injectors from wire.go:

func initialize(contextContext context.Context) (*app, error) {
	configConfig, err := config.ProvideConfigFromEnv()
	if err != nil {
		return nil, err
	}
	otelinstrumentConfig := configConfig.OtelConfig
	instrumentation, err := otelinstrument.ProvideInstrumentation(contextContext, otelinstrumentConfig)
	if err != nil {
		return nil, err
	}
	tracerProvider := otelinstrument.ProvideTracerProvider(instrumentation)
	dbConnectInfo := configConfig.DBConnectInfo
	db, err := infra.ProvideDB(tracerProvider, dbConnectInfo)
	if err != nil {
		return nil, err
	}
	mainApp, err := provideApp(db, instrumentation)
	if err != nil {
		return nil, err
	}
	return mainApp, nil
}
