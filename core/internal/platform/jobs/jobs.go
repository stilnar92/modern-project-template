// Package jobs holds the observability of the River jobs (backend standard
// 06, ADR 0014): the job in the log context, and the queue metrics for the
// alerts "the age of the oldest job increases" and "jobs in the dead-letter
// queue".
package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
)

// LogContext puts the job ID and kind into the context of each job, so
// each log record of a job has job_id and job_kind.
func LogContext() rivertype.WorkerMiddleware {
	return river.WorkerMiddlewareFunc(func(ctx context.Context, job *rivertype.JobRow, doInner func(context.Context) error) error {
		return doInner(observability.WithJob(ctx, job.ID, job.Kind))
	})
}

// The states that the alerts read. "completed" and "cancelled" are history.
var queueStates = []string{"available", "retryable", "running", "scheduled", "discarded"}

const queueQuery = `
SELECT state::text, count(*),
       coalesce(extract(epoch FROM now() - min(scheduled_at) FILTER (WHERE scheduled_at <= now())), 0)::float8
FROM river_job
WHERE state IN ('available', 'retryable', 'running', 'scheduled', 'discarded')
GROUP BY state`

// scrapeTimeout keeps /metrics fast when the database is slow.
const scrapeTimeout = 2 * time.Second

// RegisterQueueMetrics adds two gauges that read river_job at each scrape:
//   - river.queue.jobs{state}: the number of jobs in each state;
//   - river.queue.oldest_age{state}: the age in seconds of the oldest job
//     that could run, for the states available and retryable.
//
// The pool can be the app pool: the River tables have no RLS (ADR 0009).
func RegisterQueueMetrics(meter metric.Meter, pool *pgxpool.Pool) error {
	jobs, err := meter.Int64ObservableGauge("river.queue.jobs",
		metric.WithDescription("River jobs in each state"))
	if err != nil {
		return err
	}
	age, err := meter.Float64ObservableGauge("river.queue.oldest_age",
		metric.WithDescription("Age of the oldest job that could run, by state"), metric.WithUnit("s"))
	if err != nil {
		return err
	}
	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		ctx, cancel := context.WithTimeout(ctx, scrapeTimeout)
		defer cancel()
		rows, err := pool.Query(ctx, queueQuery)
		if err != nil {
			return fmt.Errorf("queue metrics: %w", err)
		}
		defer rows.Close()
		// Each state gets a value, also 0: an alert needs a series, not a gap.
		counts := map[string]int64{}
		oldest := map[string]float64{}
		for rows.Next() {
			var state string
			var n int64
			var age float64
			if err := rows.Scan(&state, &n, &age); err != nil {
				return err
			}
			counts[state], oldest[state] = n, age
		}
		if err := rows.Err(); err != nil {
			return err
		}
		for _, state := range queueStates {
			attrs := metric.WithAttributes(attribute.String("state", state))
			o.ObserveInt64(jobs, counts[state], attrs)
			if state == "available" || state == "retryable" {
				o.ObserveFloat64(age, oldest[state], attrs)
			}
		}
		return nil
	}, jobs, age)
	return err
}
