// Package config reads the settings of a process from environment
// variables. Settings come only from the environment. Each
// process describes its settings as a struct with env tags and calls Load
// once at startup (ADR 0012).
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Base holds the settings that each process reads.
type Base struct {
	// AppEnv is the environment name: local, staging or production.
	AppEnv   string     `env:"APP_ENV" envDefault:"local"`
	LogLevel slog.Level `env:"LOG_LEVEL" envDefault:"INFO"`
	// SentryDSN turns on error reports when it is set.
	SentryDSN string `env:"SENTRY_DSN"`
	// OTLPEndpoint turns on the trace export when it is set. The OTLP
	// exporter reads it and the other OTEL_* variables itself.
	OTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	// MetricsPort is the internal port of /metrics. The proxy does not
	// route it, so the metrics are not public.
	MetricsPort int `env:"METRICS_PORT" envDefault:"9090"`
}

// Load reads the settings of T from the environment. It checks all fields
// and reports all missing or wrong variables in one error, so a process
// with a wrong environment stops at startup.
func Load[T any]() (T, error) {
	cfg, err := env.ParseAs[T]()
	if err == nil {
		return cfg, nil
	}
	var agg env.AggregateError
	if !errors.As(err, &agg) {
		return cfg, fmt.Errorf("config: %w (see .env.example)", err)
	}
	keys := envKeys(reflect.TypeFor[T]())
	msgs := make([]string, 0, len(agg.Errors))
	for _, e := range agg.Errors {
		msgs = append(msgs, describe(e, keys))
	}
	return cfg, fmt.Errorf("config: %s (see .env.example)", strings.Join(msgs, "; "))
}

// describe names the variable in a parse error. The library names only the
// Go field, but the operator sets the variable.
func describe(err error, keys map[string]string) string {
	var pe env.ParseError
	if errors.As(err, &pe) {
		if key, ok := keys[pe.Name]; ok {
			return fmt.Sprintf("%s is not valid: %v", key, pe.Err)
		}
	}
	return err.Error()
}

// envKeys maps each field name of the struct to its variable, also in
// embedded structs such as Base.
func envKeys(t reflect.Type) map[string]string {
	keys := map[string]string{}
	for f := range t.Fields() {
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			for name, key := range envKeys(f.Type) {
				keys[name] = key
			}
			continue
		}
		if key, _, _ := strings.Cut(f.Tag.Get("env"), ","); key != "" {
			keys[f.Name] = key
		}
	}
	return keys
}
