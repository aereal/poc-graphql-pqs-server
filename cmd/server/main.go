package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aereal/poc-graphql-pqs-server/logging"
	"github.com/aereal/poc-graphql-pqs-server/web"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	serviceName = "graphql-pqs"
	environment = "local"
)

var (
	otelShutdownGrace = time.Second * 5
)

func main() {
	os.Exit(run())
}

func run() int {
	isDebug := os.Getenv("DEBUG") != ""
	isVerbose := os.Getenv("VERBOSE") != ""

	logging.Init(logging.WithOutput(os.Stdout), logging.WithDebug(isDebug), logging.WithStacktrace(isVerbose))

	ctx := context.Background()

	tp, err := setupOtel(ctx)
	if err != nil {
		slog.Error("failed to instrument OpenTelemetry", slog.String("error", err.Error()))
		return 1
	}
	otel.SetTracerProvider(tp)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), otelShutdownGrace)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			slog.Warn("failed to shutdown TracerProvider", slog.String("error", err.Error()))
		}
	}()

	srv := web.New(web.WithPort(os.Getenv("PORT")))
	if err := srv.Start(ctx); err != nil {
		slog.Error("server is closed", slog.String("error", err.Error()))
		return 1
	}
	return 0
}

func setupOtel(ctx context.Context) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("otlptracegrpc.New: %w", err)
	}
	res, err := resource.New(
		ctx,
		// resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("resource.New: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	return tp, nil
}
