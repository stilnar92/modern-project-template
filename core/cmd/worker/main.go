// Command worker runs the background jobs of the core with River (ADR 0012,
// ADR 0014): the event dispatch and the scheduled jobs. It also serves
// /metrics and the liveness check /healthz on the internal port.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
	"github.com/riverqueue/rivercontrib/otelriver"
	"go.opentelemetry.io/otel"

	"github.com/stilnar92/modern-project-template/core/internal/accounts"
	"github.com/stilnar92/modern-project-template/core/internal/events"
	"github.com/stilnar92/modern-project-template/core/internal/platform/config"
	"github.com/stilnar92/modern-project-template/core/internal/platform/health"
	"github.com/stilnar92/modern-project-template/core/internal/platform/jobs"
	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
)

// workerConfig is the environment of the worker (.env.example).
type workerConfig struct {
	config.Base
	DatabaseURL string `env:"DATABASE_URL,required,notEmpty"`
	// SystemDatabaseURL uses the BYPASSRLS role. Only the jobs that ADR 0009
	// lists get this pool.
	SystemDatabaseURL string `env:"SYSTEM_DATABASE_URL,required,notEmpty"`
	MaxWorkers        int    `env:"WORKER_MAX_WORKERS" envDefault:"10"`
}

// stopTimeout gives running jobs time to finish on SIGTERM. It is shorter
// than stop_grace_period in deploy/compose.yaml.
const stopTimeout = 10 * time.Second

func main() {
	healthcheck := flag.Bool("healthcheck", false,
		"call /healthz of the running worker and exit with 0 when it answers 200 (the Docker healthcheck: the image has no shell)")
	flag.Parse()
	if *healthcheck {
		os.Exit(runHealthcheck())
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runHealthcheck needs only METRICS_PORT, so it does not load the whole config.
func runHealthcheck() int {
	port := os.Getenv("METRICS_PORT")
	if port == "" {
		port = "9090"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := health.Check(ctx, "http://127.0.0.1:"+port+"/healthz"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// liveness answers /healthz of the worker (backend standard 06, lesson 0016).
// The worker is alive while its River client runs: after Start and before
// the client stops. It does not check the database: River waits for the
// database itself, and a restart does not help it.
type liveness struct {
	stopped atomic.Pointer[<-chan struct{}]
}

// running starts the check. Give it the channel of client.Stopped() right
// after Start: River gives a new channel after Stop.
func (l *liveness) running(stopped <-chan struct{}) { l.stopped.Store(&stopped) }

func (l *liveness) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	alive := false
	if ch := l.stopped.Load(); ch != nil {
		select {
		case <-*ch:
		default:
			alive = true
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if !alive {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"down"}` + "\n"))
		return
	}
	_, _ = w.Write([]byte(`{"status":"ok"}` + "\n"))
}

func run() error {
	cfg, err := config.Load[workerConfig]()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tel, err := observability.Setup(ctx, "worker", cfg.Base)
	if err != nil {
		return err
	}
	defer tel.Shutdown(context.WithoutCancel(ctx))
	logger := tel.Logger

	app, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("app pool: %w", err)
	}
	defer app.Close()
	system, err := pgxpool.New(ctx, cfg.SystemDatabaseURL)
	if err != nil {
		return fmt.Errorf("system pool: %w", err)
	}
	defer system.Close()

	client, err := newRiver(app, system, logger, cfg.MaxWorkers)
	if err != nil {
		return err
	}
	// Queue length and the age of the oldest job (backend standard 06).
	if err := jobs.RegisterQueueMetrics(otel.Meter("river-queue"), app); err != nil {
		return fmt.Errorf("queue metrics: %w", err)
	}

	live := &liveness{}
	metrics := tel.MetricsServer(cfg.MetricsPort, live)
	errc := make(chan error, 1)
	go func() {
		if err := metrics.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- fmt.Errorf("listen %s: %w", metrics.Addr, err)
		}
	}()
	if err := client.Start(ctx); err != nil {
		return fmt.Errorf("start river: %w", err)
	}
	live.running(client.Stopped())
	logger.InfoContext(ctx, "worker started", "metrics", metrics.Addr, "max_workers", cfg.MaxWorkers)

	var runErr error
	select {
	case <-ctx.Done():
	case runErr = <-errc:
		observability.ReportError(ctx, logger, "worker stopped", runErr)
	}

	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), stopTimeout)
	defer cancel()
	// Stop waits for the running jobs. After the timeout River cancels them,
	// and they run again later: each job must accept a second run.
	if err := client.Stop(stopCtx); err != nil {
		runErr = errors.Join(runErr, fmt.Errorf("stop river: %w", err))
	}
	if err := metrics.Shutdown(stopCtx); err != nil {
		runErr = errors.Join(runErr, err)
	}
	logger.InfoContext(stopCtx, "worker stopped")
	return runErr
}

// newRiver registers each job of the core. A new job kind goes here.
func newRiver(app, system *pgxpool.Pool, logger *slog.Logger, maxWorkers int) (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()
	river.AddWorker(workers, events.NewDispatchWorker(app, logger))
	river.AddWorker(workers, accounts.NewCleanupWorker(system, logger))

	client, err := river.NewClient(riverpgxv5.New(app), &river.Config{
		Logger:       logger,
		Queues:       map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: maxWorkers}},
		Workers:      workers,
		PeriodicJobs: []*river.PeriodicJob{accounts.CleanupJob()},
		// A span and metrics for each job. The trace context goes in the job
		// metadata, so a trace runs from the request to the job (standard 06).
		// Then the job ID and kind go into the log context of each job.
		Middleware: []rivertype.Middleware{otelriver.NewMiddleware(nil), jobs.LogContext()},
	})
	if err != nil {
		return nil, fmt.Errorf("river client: %w", err)
	}
	return client, nil
}
