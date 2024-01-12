package main

import (
	"context"
	"log/slog"
	"os"
	"time"

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

	srv := web.New(web.WithPort(os.Getenv("PORT")))
	if err := srv.Start(ctx); err != nil {
		slog.Error("server is closed", slog.String("error", err.Error()))
		return 1
	}
	return 0
}
