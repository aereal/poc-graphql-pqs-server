package infra

import (
	"net/url"
	"strconv"

	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const driverPgx = "pgx"

type Option func(dbURL *url.URL)

func WithUser(user string) Option {
	return func(dbURL *url.URL) {
		if passwd, ok := dbURL.User.Password(); ok {
			dbURL.User = url.UserPassword(user, passwd)
			return
		}
		dbURL.User = url.User(user)
	}
}

func WithPassword(passwd string) Option {
	return func(dbURL *url.URL) { dbURL.User = url.UserPassword(dbURL.User.Username(), passwd) }
}

func WithDBName(name string) Option {
	return func(dbURL *url.URL) { dbURL.Path = "/" + url.PathEscape(name) }
}

func WithSSLMode(mode string) Option {
	return func(dbURL *url.URL) {
		params := dbURL.Query()
		params.Set("sslmode", mode)
		dbURL.RawQuery = params.Encode()
	}
}

func WithAddr(addr string) Option {
	return func(dbURL *url.URL) { dbURL.Host = addr }
}

func OpenDB(opts ...Option) (*sqlx.DB, error) {
	dbURL := &url.URL{Scheme: "postgres"}
	for _, o := range opts {
		o(dbURL)
	}
	db, err := otelsql.Open(driverPgx, dbURL.String(), otelsql.WithAttributes(buildDefaultAttrs(dbURL)...))
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(db, driverPgx), nil
}

func buildDefaultAttrs(dbURL *url.URL) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 6)
	attrs = append(attrs, semconv.NetTransportTCP)
	attrs = append(attrs, semconv.DBSystemPostgreSQL)
	attrs = append(attrs, semconv.ServerAddress(dbURL.Hostname()))
	if port, err := strconv.Atoi(dbURL.Port()); err == nil {
		attrs = append(attrs, semconv.ServerPort(port))
	}
	if username := dbURL.User.Username(); username != "" {
		attrs = append(attrs, semconv.DBUser(username))
	}
	if path := dbURL.Path; len(path) > 1 {
		attrs = append(attrs, semconv.DBName(path[1:]))
	}
	return attrs
}
