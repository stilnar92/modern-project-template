package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/stilnar92/modern-project-template/core/internal/platform/config"
)

func TestContextHandlerAddsIDs(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewContextHandler(slog.NewJSONHandler(&buf, nil))).With("service", "test")

	ctx := WithRequestID(t.Context(), "req-1")
	ctx, span := sdktrace.NewTracerProvider().Tracer("test").Start(ctx, "op")
	defer span.End()
	logger.InfoContext(ctx, "hello")

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	if rec["request_id"] != "req-1" || rec["trace_id"] != span.SpanContext().TraceID().String() || rec["service"] != "test" {
		t.Fatalf("record: %v", rec)
	}
}

func TestContextHandlerAddsJob(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewContextHandler(slog.NewJSONHandler(&buf, nil)))
	logger.InfoContext(WithJob(t.Context(), 42, "events.dispatch"), "work")

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	if rec["job_id"] != float64(42) || rec["job_kind"] != "events.dispatch" {
		t.Fatalf("record: %v", rec)
	}
}

func TestContextHandlerWithoutIDs(t *testing.T) {
	var buf bytes.Buffer
	slog.New(NewContextHandler(slog.NewJSONHandler(&buf, nil))).InfoContext(t.Context(), "hello")
	if strings.Contains(buf.String(), "request_id") || strings.Contains(buf.String(), "trace_id") {
		t.Fatalf("empty IDs must not appear: %s", buf.String())
	}
}

func TestSetupServesMetrics(t *testing.T) {
	tel, err := Setup(t.Context(), "test", config.Base{AppEnv: "test", MetricsPort: 9090})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tel.Shutdown(t.Context()) })

	// A metric from our code reaches the Prometheus output.
	counter, err := otel.Meter("test").Int64Counter("test_events")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(t.Context(), 1)

	rec := httptest.NewRecorder()
	tel.MetricsServer(0, nil).Handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil))
	body := rec.Body.String()
	for _, want := range []string{"test_events", "go_goroutine", `service_name="test"`} {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics has no %s", want)
		}
	}
}
