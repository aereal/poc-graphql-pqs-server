package main

import (
	"context"
	"log/slog"
	"os"

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
	app, err := initialize(ctx)
	if err != nil {
		var attrs []slog.Attr
		if errList, ok := err.(interface{ Unwrap() []error }); ok {
			for _, e := range errList.Unwrap() {
				attrs = append(attrs, slog.String("error", e.Error()))
			}
		} else {
			attrs = append(attrs, slog.String("error", err.Error()))
		}
		slog.LogAttrs(context.Background(), slog.LevelError, "failed to initialize server", attrs...)
		return 1
	}
	defer app.instrumentation.Cleanup()
	if err := app.server.Start(ctx); err != nil {
		slog.Error("server is closed", slog.String("error", err.Error()))
		return 1
	}
	return 0
}

type app struct {
	server          *web.Server
	instrumentation *otelinstrument.Instrumentation
}

func provideApp(server *web.Server, inst *otelinstrument.Instrumentation) *app {
	return &app{server: server, instrumentation: inst}
}
