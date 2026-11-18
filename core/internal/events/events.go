// Package events owns the event log of the core, which is also the outbox
// (README.md, ADR 0003, ADR 0014).
package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/stilnar92/modern-project-template/core/internal/events/store"
	"github.com/stilnar92/modern-project-template/core/internal/platform/db"
)

// Event is a fact that happened in an account. Add each name to
// docs/dev/events.md in the same PR.
type Event struct {
	// Name is <module>.<thing>_<past tense verb>, for example "accounts.session_started".
	Name    string
	Version int64
	// Data holds IDs and the changed fields. It is stored as JSON.
	Data any
}

// Stored is an event as the log keeps it.
type Stored struct {
	ID         uuid.UUID
	AccountID  uuid.UUID
	Name       string
	Version    int64
	OccurredAt time.Time
	Data       json.RawMessage
}

// Recorder writes events. It needs a River client: an insert-only client
// is enough in the api.
type Recorder struct {
	river *river.Client[pgx.Tx]
}

// NewRecorder returns a recorder.
func NewRecorder(client *river.Client[pgx.Tx]) *Recorder {
	return &Recorder{river: client}
}

// Record writes the event and its dispatch job in tx: the transaction of
// the action inside db.WithAccount. So the action, the event and the job
// commit together or not at all (backend standard 03).
func (r *Recorder) Record(ctx context.Context, tx pgx.Tx, accountID uuid.UUID, e Event) (uuid.UUID, error) {
	data, err := json.Marshal(e.Data)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("event %s data: %w", e.Name, err)
	}
	if e.Version == 0 {
		e.Version = 1
	}
	id := uuid.NewV7()
	err = store.New(tx).InsertEvent(ctx, store.InsertEventParams{
		ID:         id,
		AccountID:  accountID,
		Name:       e.Name,
		Version:    e.Version,
		OccurredAt: time.Now(),
		Data:       data,
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("record event %s: %w", e.Name, err)
	}
	if _, err := r.river.InsertTx(ctx, tx, DispatchArgs{EventID: id, AccountID: accountID}, nil); err != nil {
		return uuid.UUID{}, fmt.Errorf("queue event %s: %w", e.Name, err)
	}
	return id, nil
}

// Sink receives events, for example the adapter of the analytics tool. A
// sink decides what may leave the core: the personal data of clients never
// does (ADR 0003). Send must accept the same event twice: a job can retry.
type Sink interface {
	Name() string
	Send(ctx context.Context, e Stored) error
}

// DispatchArgs is the job that sends one event to the sinks. It carries
// the account, so the worker reads the event through db.WithAccount (ADR 0009).
type DispatchArgs struct {
	EventID   uuid.UUID `json:"event_id"`
	AccountID uuid.UUID `json:"account_id"`
}

// Kind is the River job name.
func (DispatchArgs) Kind() string { return "events.dispatch" }

// DispatchWorker sends events to the sinks. A failed sink fails the job,
// and River retries it. A job that runs out of attempts is "discarded":
// the metrics and the alert of backend standard 06 show it.
type DispatchWorker struct {
	river.WorkerDefaults[DispatchArgs]
	pool   *pgxpool.Pool
	sinks  []Sink
	logger *slog.Logger
}

// NewDispatchWorker returns the worker. With no sinks it only reads the event.
func NewDispatchWorker(pool *pgxpool.Pool, logger *slog.Logger, sinks ...Sink) *DispatchWorker {
	return &DispatchWorker{pool: pool, sinks: sinks, logger: logger}
}

// Work sends the event to each sink.
func (w *DispatchWorker) Work(ctx context.Context, job *river.Job[DispatchArgs]) error {
	var ev Stored
	err := db.WithAccount(ctx, w.pool, job.Args.AccountID.String(), func(tx pgx.Tx) error {
		row, err := store.New(tx).GetEvent(ctx, job.Args.EventID)
		if err != nil {
			return err
		}
		ev = Stored{ID: row.ID, AccountID: row.AccountID, Name: row.Name, Version: row.Version, OccurredAt: row.OccurredAt, Data: row.Data}
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		// The account was deleted with its events. Nothing to send.
		w.logger.InfoContext(ctx, "event is gone", "event_id", job.Args.EventID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("read event: %w", err)
	}
	var errs []error
	for _, s := range w.sinks {
		if err := s.Send(ctx, ev); err != nil {
			errs = append(errs, fmt.Errorf("sink %s: %w", s.Name(), err))
		}
	}
	return errors.Join(errs...)
}
