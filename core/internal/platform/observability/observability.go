// Package observability sets up the logs, traces, metrics and error reports
// of a process (backend standard 06, ADR 0012):
//
//   - JSON logs to stdout. Each record gets request_id and trace_id from the
//     context, so log calls must use the *Context methods.
//   - OpenTelemetry traces. Each process has a tracer, so trace IDs exist in
//     the logs. The export over OTLP starts only when
//     OTEL_EXPORTER_OTLP_ENDPOINT is set.
//   - OpenTelemetry metrics, read by Prometheus on the internal /metrics port.
//   - Sentry, when SENTRY_DSN is set.
package observability

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"

	"github.com/stilnar92/modern-project-template/core/internal/platform/config"
	"github.com/stilnar92/modern-project-template/core/internal/platform/version"
)

// Telemetry is the observability of one process.
type Telemetry struct {
	Logger *slog.Logger
	// Metrics serves the Prometheus format. Serve it on the internal
	// metrics port only (MetricsServer).
	Metrics  http.Handler
	shutdown []func(context.Context) error
}

// Setup starts the observability of the process. Call Shutdown before the
// process exits, so the buffered traces and error reports go out.
func Setup(ctx context.Context, service string, cfg config.Base) (*Telemetry, error) {
	logger := slog.New(NewContextHandler(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}),
	)).With("service", service, "version", version.Version)
	t := &Telemetry{Logger: logger}

	res := resource.NewSchemaless(
		semconv.ServiceName(service),
		semconv.ServiceVersion(version.Version),
		semconv.DeploymentEnvironmentNameKey.String(cfg.AppEnv),
	)

	// Traces. Without an endpoint the spans stay in the process: the logs
	// still get trace IDs, and nothing goes over the network.
	traceOpts := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	if cfg.OTLPEndpoint != "" {
		exp, err := otlptracehttp.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("otlp trace exporter: %w", err)
		}
		traceOpts = append(traceOpts, sdktrace.WithBatcher(exp))
		logger.InfoContext(ctx, "trace export enabled")
	}
	tp := sdktrace.NewTracerProvider(traceOpts...)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	t.shutdown = append(t.shutdown, tp.Shutdown)

	// Metrics. A registry of our own: only our metrics, not global state.
	registry := prometheus.NewRegistry()
	exp, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		return nil, fmt.Errorf("prometheus exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exp), sdkmetric.WithResource(res))
	otel.SetMeterProvider(mp)
	t.shutdown = append(t.shutdown, mp.Shutdown)
	if err := runtime.Start(runtime.WithMeterProvider(mp)); err != nil {
		return nil, fmt.Errorf("runtime metrics: %w", err)
	}
	t.Metrics = promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	if cfg.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         cfg.SentryDSN,
			Environment: cfg.AppEnv,
			Release:     version.Version,
			ServerName:  service,
		})
		if err != nil {
			return nil, fmt.Errorf("sentry: %w", err)
		}
		t.shutdown = append(t.shutdown, func(context.Context) error {
			sentry.Flush(2 * time.Second)
			return nil
		})
		logger.InfoContext(ctx, "sentry enabled")
	}
	return t, nil
}

// Shutdown sends the buffered traces and error reports.
func (t *Telemetry) Shutdown(ctx context.Context) {
	var errs []error
	for _, f := range t.shutdown {
		errs = append(errs, f(ctx))
	}
	if err := errors.Join(errs...); err != nil {
		t.Logger.ErrorContext(ctx, "telemetry shutdown", "err", err)
	}
}

// MetricsServer serves /metrics on the internal port. Do not route this
// port through the proxy. A process without its own HTTP server (the worker)
// gives a liveness handler, and the server also answers /healthz with it.
func (t *Telemetry) MetricsServer(port int, liveness http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", t.Metrics)
	if liveness != nil {
		mux.Handle("GET /healthz", liveness)
	}
	return &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

// ReportError logs the error and sends it to Sentry when Sentry is enabled.
// The report gets the request ID of the context.
func ReportError(ctx context.Context, logger *slog.Logger, msg string, err error) {
	logger.ErrorContext(ctx, msg, "err", err)
	hub := sentry.CurrentHub().Clone()
	if id := RequestID(ctx); id != "" {
		hub.Scope().SetTag("request_id", id)
	}
	hub.CaptureException(err)
}
