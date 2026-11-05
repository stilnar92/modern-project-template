// Package observability sets up the logs and the error reports of a process
// (backend standard 06): JSON logs to stdout, and Sentry when SENTRY_DSN is set.
package observability

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/stilnar92/modern-project-template/core/internal/platform/config"
	"github.com/stilnar92/modern-project-template/core/internal/platform/version"
)

// Setup returns the logger of the service and a function that sends the
// buffered error reports. Call the function before the process exits.
func Setup(service string) (*slog.Logger, func(), error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.ToUpper(config.String("LOG_LEVEL", "INFO")))); err != nil {
		return nil, nil, fmt.Errorf("LOG_LEVEL: %w", err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})).
		With("service", service, "version", version.Version)

	flush := func() {}
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         dsn,
			Environment: config.String("APP_ENV", "local"),
			Release:     version.Version,
			ServerName:  service,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("sentry: %w", err)
		}
		flush = func() { sentry.Flush(2 * time.Second) }
		logger.Info("sentry enabled")
	}
	return logger, flush, nil
}

// ReportError logs the error and sends it to Sentry when Sentry is enabled.
func ReportError(logger *slog.Logger, msg string, err error) {
	logger.Error(msg, "err", err)
	sentry.CaptureException(err)
}
