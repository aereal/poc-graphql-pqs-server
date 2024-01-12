package logging

import (
	"context"
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

type config struct {
	handlerOptions *slog.HandlerOptions
	output         io.Writer
}

type Option func(*config)

func WithStacktrace(on bool) Option { return func(c *config) { c.handlerOptions.AddSource = on } }

func WithDebug(on bool) Option {
	return func(c *config) { c.handlerOptions.Level = constLeveler(slog.LevelDebug) }
}

func WithOutput(w io.Writer) Option { return func(c *config) { c.output = w } }

func Init(optFns ...Option) {
	cfg := &config{handlerOptions: new(slog.HandlerOptions)}
	for _, f := range optFns {
		f(cfg)
	}
	if cfg.output == nil {
		cfg.output = io.Discard
	}
	logger := slog.New(&otelTraceIDHandler{Handler: slog.NewJSONHandler(cfg.output, cfg.handlerOptions)})
	slog.SetDefault(logger)
}

type otelTraceIDHandler struct{ slog.Handler }

var _ slog.Handler = (*otelTraceIDHandler)(nil)

func (h *otelTraceIDHandler) Handle(ctx context.Context, record slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		record.AddAttrs(slog.String("otel.trace_id", sc.TraceID().String()), slog.String("otel.span_id", sc.SpanID().String()))
	}
	return h.Handler.Handle(ctx, record)
}

type constLeveler slog.Level

var _ slog.Leveler = (constLeveler)(0)

func (l constLeveler) Level() slog.Level { return slog.Level(l) }
