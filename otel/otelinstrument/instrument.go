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
)

const (
	serviceName = "graphql-pqs"
	environment = "local"
)

type config struct {
	shutdownGrace           time.Duration
	setGlobalTracerProvider bool
}

type Option func(*config)

func WithShutdownGrace(grace time.Duration) Option {
	return func(c *config) { c.shutdownGrace = grace }
}

func WithSetGlobalTracerProvider(on bool) Option {
	return func(c *config) { c.setGlobalTracerProvider = on }
}

func Instrument(ctx context.Context, opts ...Option) (func(), error) {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}

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
	if cfg.setGlobalTracerProvider {
		otel.SetTracerProvider(tp)
	}
	return func() {
		ctx, cancel := context.WithCancel(context.Background())
		if cfg.shutdownGrace > 0 {
			ctx, cancel = context.WithTimeout(ctx, cfg.shutdownGrace)
		}
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			slog.Warn("failed to shutdown TracerProvider", slog.String("error", err.Error()))
		}
	}, nil
}
