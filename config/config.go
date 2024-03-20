package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aereal/poc-graphql-pqs-server/graph/persistedquery/apollo"
	"github.com/aereal/poc-graphql-pqs-server/infra"
	"github.com/aereal/poc-graphql-pqs-server/otel/otelinstrument"
	"github.com/aereal/poc-graphql-pqs-server/web"
	"github.com/google/wire"
)

type Config struct {
	Port                       web.Port
	PersistedQueryManifestFile apollo.ManifestFilePath
	DBConnectInfo              *infra.DBConnectInfo
	OtelConfig                 *otelinstrument.Config
}

var EnvironmentProvider = wire.NewSet(
	ProvideConfigFromEnv,
	wire.FieldsOf(
		new(*Config),
		"Port",
		"PersistedQueryManifestFile",
		"DBConnectInfo",
		"OtelConfig",
	),
)

func ProvideConfigFromEnv() (*Config, error) {
	c := &Config{DBConnectInfo: new(infra.DBConnectInfo), OtelConfig: new(otelinstrument.Config)}
	err := errors.Join(
		getEnv((*string)(&c.Port), "PORT", optional(true)),
		getEnv((*string)(&c.PersistedQueryManifestFile), "PERSISTED_QUERY_MANIFEST_FILE"),
		getEnv(&c.DBConnectInfo.Addr, "DB_ADDR"),
		getEnv(&c.DBConnectInfo.DBName, "DB_NAME"),
		getEnv(&c.DBConnectInfo.Username, "DB_USER"),
		getEnv(&c.DBConnectInfo.Passwd, "DB_PASSWORD"),
		getEnv(&c.DBConnectInfo.SSLMode, "DB_SSL_MODE", optional(true)),
		getVar(&c.OtelConfig.ShutdownGrace, "OTEL_SHUTDOWN_GRACE", time.ParseDuration, optional(true)),
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type getEnvOption func(c *getEnvConfig)

func optional(v bool) getEnvOption { return func(c *getEnvConfig) { c.optional = v } }

type getEnvConfig struct {
	optional bool
}

type EnvironmentVariableNotFoundError struct {
	Name string
}

func (e *EnvironmentVariableNotFoundError) Error() string {
	return fmt.Sprintf("environment variable %s is not given", e.Name)
}

func getEnv(ptr *string, name string, opts ...getEnvOption) error {
	return getVar(ptr, name, asString, opts...)
}

func getVar[T any](ptr *T, name string, transform func(string) (T, error), opts ...getEnvOption) error {
	cfg := new(getEnvConfig)
	for _, o := range opts {
		o(cfg)
	}
	s, ok := os.LookupEnv(name)
	if cfg.optional && !ok {
		return &EnvironmentVariableNotFoundError{Name: name}
	}
	val, err := transform(s)
	if err != nil {
		return err
	}
	*ptr = val
	return nil
}

func asString(s string) (string, error) { return s, nil }
