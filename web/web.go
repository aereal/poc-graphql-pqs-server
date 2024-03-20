package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/aereal/otelgqlgen"
	"github.com/aereal/poc-graphql-pqs-server/graph/loaders"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

const defaultPort = "8080"

var shutdownGrace = time.Second * 5

type Port string

func ProvideServer(port Port, es graphql.ExecutableSchema, loadersRoot *loaders.Root, queryList graphql.Cache, tp trace.TracerProvider) (*Server, error) {
	// TODO: check args
	s := &Server{
		port:             string(port),
		executableSchema: es,
		loaderRoot:       loadersRoot,
		queryList:        queryList,
		tp:               tp,
	}
	if s.port == "" {
		s.port = defaultPort
	}
	return s, nil
}

type Server struct {
	port             string
	executableSchema graphql.ExecutableSchema
	loaderRoot       *loaders.Root
	queryList        graphql.Cache
	tp               trace.TracerProvider
}

func (s *Server) handlerRoot() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") })
}

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", s.handlerRoot())
	mux.Handle("/graphql", s.handlerGraphql(false))
	mux.Handle("/public/graphql", s.handlerGraphql(true))
	return s.withOtel(mux)
}

func (s *Server) handlerGraphql(public bool) http.Handler {
	h := handler.New(s.executableSchema)
	h.AddTransport(transport.POST{})
	if public {
		h.Use(extension.AutomaticPersistedQuery{Cache: s.queryList})
	} else {
		h.Use(extension.Introspection{})
	}
	h.Use(otelgqlgen.New())
	h.Use(s.loaderRoot)
	opts := cors.Options{
		AllowedMethods:   []string{http.MethodPost},
		AllowCredentials: true,
	}
	return cors.New(opts).Handler(h)
}

func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:    net.JoinHostPort("", s.port),
		Handler: s.handler(),
	}
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go func() {
		<-ctx.Done()
		slog.DebugContext(ctx, "shutting down server", slog.Duration("shutdown_grace", shutdownGrace))
		ctx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.WarnContext(ctx, "cannot shutting down server gracefully", slog.String("error", err.Error()))
		}
	}()
	slog.InfoContext(ctx, "start server", slog.String("port", s.port))
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) withOtel(next http.Handler) http.Handler {
	return otelhttp.NewMiddleware("",
		otelhttp.WithTracerProvider(s.tp),
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		otelhttp.WithSpanNameFormatter(formatSpanName),
		otelhttp.WithPublicEndpoint(),
		otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace { return otelhttptrace.NewClientTrace(ctx) }),
	)(next)
}

func formatSpanName(_ string, r *http.Request) string {
	return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
}
