package accounts

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/stilnar92/modern-project-template/core/internal/accounts/store"
)

// CleanupArgs is the scheduled job that deletes expired sessions.
type CleanupArgs struct{}

// Kind is the River job name.
func (CleanupArgs) Kind() string { return "accounts.cleanup_sessions" }

// CleanupInterval is how often the worker runs the clean-up.
const CleanupInterval = time.Hour

// CleanupWorker deletes the expired sessions of all accounts. It is a
// system-role place (ADR 0009: scheduled deletion): systemPool must use the
// app_system role. No other code of the module uses that pool.
type CleanupWorker struct {
	river.WorkerDefaults[CleanupArgs]
	systemPool *pgxpool.Pool
	logger     *slog.Logger
}

// NewCleanupWorker returns the worker.
func NewCleanupWorker(systemPool *pgxpool.Pool, logger *slog.Logger) *CleanupWorker {
	return &CleanupWorker{systemPool: systemPool, logger: logger}
}

// Work deletes the expired sessions.
func (w *CleanupWorker) Work(ctx context.Context, _ *river.Job[CleanupArgs]) error {
	n, err := store.New(w.systemPool).DeleteExpiredSessions(ctx)
	if err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	w.logger.InfoContext(ctx, "expired sessions deleted", "count", n)
	return nil
}

// CleanupJob is the periodic job for the River client of the worker.
func CleanupJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(CleanupInterval),
		func() (river.JobArgs, *river.InsertOpts) { return CleanupArgs{}, nil },
		&river.PeriodicJobOpts{RunOnStart: true},
	)
}
