package jobs

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
	"github.com/stilnar92/modern-project-template/core/internal/platform/testdb"
)

type testArgs struct{}

func (testArgs) Kind() string { return "test.job" }

func TestQueueMetrics(t *testing.T) {
	d := testdb.Start(t)
	ctx := t.Context()
	client, err := river.NewClient[pgx.Tx](riverpgxv5.New(d.App), &river.Config{})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if _, err := client.Insert(ctx, testArgs{}, nil); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := d.Owner.Exec(ctx, "UPDATE river_job SET scheduled_at = now() - interval '90 seconds' WHERE id = (SELECT min(id) FROM river_job)"); err != nil {
		t.Fatal(err)
	}

	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	if err := RegisterQueueMetrics(provider.Meter("test"), d.App); err != nil {
		t.Fatal(err)
	}
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		t.Fatal(err)
	}

	got := map[string]any{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			switch data := m.Data.(type) {
			case metricdata.Gauge[int64]:
				for _, p := range data.DataPoints {
					state, _ := p.Attributes.Value("state")
					got[m.Name+"/"+state.AsString()] = p.Value
				}
			case metricdata.Gauge[float64]:
				for _, p := range data.DataPoints {
					state, _ := p.Attributes.Value("state")
					got[m.Name+"/"+state.AsString()] = p.Value
				}
			}
		}
	}
	if got["river.queue.jobs/discarded"] != int64(0) {
		t.Fatalf("an empty state must report 0: %v", got)
	}
	if got["river.queue.jobs/available"] != int64(2) {
		t.Fatalf("available jobs: %v", got)
	}
	if age, _ := got["river.queue.oldest_age/available"].(float64); age < 89 {
		t.Fatalf("oldest age = %v, want about 90 s: %v", age, got)
	}
}

// Each log record of a job has job_id and job_kind.
func TestLogContext(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(observability.NewContextHandler(slog.NewJSONHandler(&buf, nil)))
	err := LogContext().Work(t.Context(), &rivertype.JobRow{ID: 7, Kind: "test.job"}, func(ctx context.Context) error {
		logger.InfoContext(ctx, "work")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"job_id":7`) || !strings.Contains(buf.String(), `"job_kind":"test.job"`) {
		t.Fatalf("record: %s", buf.String())
	}
}
