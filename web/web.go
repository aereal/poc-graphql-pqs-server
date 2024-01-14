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
)

const defaultPort = "8080"

var shutdownGrace = time.Second * 5

type Option func(*Server)

func WithPort(port string) Option { return func(s *Server) { s.port = port } }

func WithExecutableSchema(es graphql.ExecutableSchema) Option {
	return func(s *Server) { s.executableSchema = es }
}

func WithLoaderRoot(r *loaders.Root) Option { return func(s *Server) { s.loaderRoot = r } }

func New(opts ...Option) *Server {
	s := &Server{}
	for _, o := range opts {
		o(s)
	}
	if s.port == "" {
		s.port = defaultPort
	}
	return s
}

type Server struct {
	port             string
	executableSchema graphql.ExecutableSchema
	loaderRoot       *loaders.Root
}

func (s *Server) handlerRoot() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") })
}

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", s.handlerRoot())
	mux.Handle("/graphql", s.handlerGraphql(false))
	mux.Handle("/public/graphql", s.handlerGraphql(true))
	return withOtel(mux)
}

func (s *Server) handlerGraphql(public bool) http.Handler {
	h := handler.New(s.executableSchema)
	h.AddTransport(transport.POST{})
	if public {
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

func withOtel(next http.Handler) http.Handler {
	return otelhttp.NewMiddleware("", otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents), otelhttp.WithSpanNameFormatter(formatSpanName), otelhttp.WithPublicEndpoint(), otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace { return otelhttptrace.NewClientTrace(ctx) }))(next)
}

func formatSpanName(_ string, r *http.Request) string {
	return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
}
