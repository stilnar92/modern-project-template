package config

import (
	"log/slog"
	"strings"
	"testing"
)

type testConfig struct {
	Base
	DatabaseURL string `env:"DATABASE_URL,required,notEmpty"`
	Port        int    `env:"PORT" envDefault:"8080"`
}

func TestLoadDefaults(t *testing.T) {
	for _, k := range []string{"APP_ENV", "LOG_LEVEL", "SENTRY_DSN", "OTEL_EXPORTER_OTLP_ENDPOINT", "METRICS_PORT", "PORT"} {
		t.Setenv(k, "")
	}
	t.Setenv("DATABASE_URL", "postgres://x")

	cfg, err := Load[testConfig]()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppEnv != "local" || cfg.LogLevel != slog.LevelInfo || cfg.MetricsPort != 9090 || cfg.Port != 8080 {
		t.Fatalf("defaults: %+v", cfg)
	}
}

func TestLoadValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("PORT", "9000")

	cfg, err := Load[testConfig]()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.LogLevel != slog.LevelDebug || cfg.Port != 9000 || cfg.DatabaseURL != "postgres://x" {
		t.Fatalf("values: %+v", cfg)
	}
}

// All errors come in one message, so the operator fixes them at once.
func TestLoadReportsAllErrors(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PORT", "not-a-number")
	t.Setenv("LOG_LEVEL", "loud")

	_, err := Load[testConfig]()
	if err == nil {
		t.Fatal("want an error")
	}
	for _, want := range []string{"DATABASE_URL", "PORT", "LOG_LEVEL", ".env.example"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not name %s: %v", want, err)
		}
	}
}
