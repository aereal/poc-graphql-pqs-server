package infra

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const driverPgx = "pgx"

type DBConnectInfo struct {
	Addr     string
	Username string
	Passwd   string
	DBName   string
	SSLMode  string
}

func ProvideDB(tp trace.TracerProvider, info *DBConnectInfo) (*sqlx.DB, error) {
	if tp == nil {
		return nil, errors.New("trace.TracerProvider is required")
	}
	if info == nil {
		return nil, errors.New("DBConnectInfo is required")
	}
	dbURL := &url.URL{Scheme: "postgres"}
	dbURL.User = url.UserPassword(info.Username, info.Passwd)
	dbURL.Host = info.Addr
	dbURL.Path = "/" + url.PathEscape(info.DBName)
	if info.SSLMode != "" {
		params := dbURL.Query()
		params.Set("sslmode", info.SSLMode)
		dbURL.RawQuery = params.Encode()
	}
	db, err := otelsql.Open(driverPgx, dbURL.String(), otelsql.WithTracerProvider(tp), otelsql.WithAttributes(buildDefaultAttrs(dbURL)...))
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
