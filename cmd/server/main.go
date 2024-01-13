package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/aereal/poc-graphql-pqs-server/domain"
	"github.com/aereal/poc-graphql-pqs-server/graph"
	"github.com/aereal/poc-graphql-pqs-server/graph/loaders"
	"github.com/aereal/poc-graphql-pqs-server/graph/resolvers"
	"github.com/aereal/poc-graphql-pqs-server/infra"
	"github.com/aereal/poc-graphql-pqs-server/logging"
	"github.com/aereal/poc-graphql-pqs-server/otel/otelinstrument"
	"github.com/aereal/poc-graphql-pqs-server/web"
)

func main() {
	os.Exit(run())
}

func run() int {
	isDebug := os.Getenv("DEBUG") != ""
	isVerbose := os.Getenv("VERBOSE") != ""

	logging.Init(logging.WithOutput(os.Stdout), logging.WithDebug(isDebug), logging.WithStacktrace(isVerbose))

	ctx := context.Background()

	shutdown, err := otelinstrument.Instrument(ctx, otelinstrument.WithShutdownGrace(time.Second*5), otelinstrument.WithSetGlobalTracerProvider(true))
	if err != nil {
		slog.Error("failed to instrument OpenTelemetry", slog.String("error", err.Error()))
		return 1
	}
	defer shutdown()

	db, err := infra.OpenDB(infra.WithAddr(os.Getenv("DB_ADDR")), infra.WithDBName(os.Getenv("DB_NAME")), infra.WithUser(os.Getenv("DB_USER")), infra.WithPassword(os.Getenv("DB_PASSWORD")), infra.WithSSLMode("disable"))
	if err != nil {
		slog.Error("failed to open DB", slog.String("error", err.Error()))
		return 1
	}
	characterRepo := domain.NewCharacterRepository(domain.WithDB(db))
	loaderRoot := loaders.New(loaders.WithCharacterRepository(characterRepo))
	resolverRoot := resolvers.New(resolvers.WithCharacterRepository(characterRepo))
	es := graph.NewExecutableSchema(graph.Config{Resolvers: resolverRoot})
	srv := web.New(web.WithPort(os.Getenv("PORT")), web.WithExecutableSchema(es), web.WithLoaderRoot(loaderRoot))
	if err := srv.Start(ctx); err != nil {
		slog.Error("server is closed", slog.String("error", err.Error()))
		return 1
	}
	return 0
}
