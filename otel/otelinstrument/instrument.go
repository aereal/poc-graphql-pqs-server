package otelinstrument

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName = "graphql-pqs"
	environment = "local"
)

type Config struct {
	ShutdownGrace time.Duration
}

func ProvideInstrumentation(ctx context.Context, cfg *Config) (*Instrumentation, error) {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("otlptracegrpc.New: %w", err)
	}
	res, err := resource.New(
		ctx,
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
	otel.SetTracerProvider(tp)
	i := &Instrumentation{tp: tp}
	if cfg.ShutdownGrace > 0 {
		i.shutdownGrace = cfg.ShutdownGrace
	}
	return i, nil
}

type Instrumentation struct {
	tp            *sdktrace.TracerProvider
	shutdownGrace time.Duration
}

func (i *Instrumentation) TracerProvider() trace.TracerProvider { return i.tp }

func (i *Instrumentation) Cleanup() {
	ctx, cancel := context.WithCancel(context.Background())
	if i.shutdownGrace > 0 {
		ctx, cancel = context.WithTimeout(ctx, i.shutdownGrace)
	}
	defer cancel()
	if err := i.tp.Shutdown(ctx); err != nil {
		slog.LogAttrs(ctx, slog.LevelWarn, "failed to shutdown TracerProvider", slog.String("error", err.Error()))
	}
}
