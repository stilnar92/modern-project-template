package events

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/stilnar92/modern-project-template/core/internal/platform/db"
	"github.com/stilnar92/modern-project-template/core/internal/platform/testdb"
)

var quiet = slog.New(slog.DiscardHandler)

// recordingSink keeps what it receives. fail makes Send return an error.
type recordingSink struct {
	mu   sync.Mutex
	got  []Stored
	fail error
}

func (s *recordingSink) Name() string { return "test" }

func (s *recordingSink) Send(_ context.Context, e Stored) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail != nil {
		return s.fail
	}
	s.got = append(s.got, e)
	return nil
}

func (s *recordingSink) events() []Stored {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Stored(nil), s.got...)
}

// newAccount inserts an account row for the test.
func newAccount(t *testing.T, d *testdb.DB) uuid.UUID {
	t.Helper()
	id := uuid.NewV7()
	err := db.WithAccount(t.Context(), d.App, id.String(), func(tx pgx.Tx) error {
		_, err := tx.Exec(t.Context(), "INSERT INTO accounts (id) VALUES ($1)", id)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// insertOnly is a River client that only inserts jobs, as the api uses it.
func insertOnly(t *testing.T, d *testdb.DB) *river.Client[pgx.Tx] {
	t.Helper()
	c, err := river.NewClient(riverpgxv5.New(d.App), &river.Config{Logger: quiet})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func countJobs(t *testing.T, d *testdb.DB) int {
	t.Helper()
	var n int
	if err := d.Owner.QueryRow(t.Context(), "SELECT count(*) FROM river_job WHERE kind = $1", DispatchArgs{}.Kind()).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestEventsIsolation(t *testing.T) {
	d := testdb.Start(t)
	testdb.AssertIsolation(t, d, testdb.IsolationCase{
		Table: "events",
		Insert: func(ctx context.Context, tx pgx.Tx, accountID string) error {
			if _, err := tx.Exec(ctx, "INSERT INTO accounts (id) VALUES ($1) ON CONFLICT DO NOTHING", accountID); err != nil {
				return err
			}
			_, err := tx.Exec(ctx, "INSERT INTO events (id, account_id, name) VALUES ($1, $2, 'test.thing_happened')", uuid.NewV7(), accountID)
			return err
		},
	})
}

// The event and its job commit with the action, or neither exists.
func TestRecordIsAtomic(t *testing.T) {
	d := testdb.Start(t)
	ctx := t.Context()
	acc := newAccount(t, d)
	rec := NewRecorder(insertOnly(t, d))
	e := Event{Name: "test.thing_happened", Data: map[string]string{"thing_id": "t1"}}

	errAction := errors.New("the action failed")
	err := db.WithAccount(ctx, d.App, acc.String(), func(tx pgx.Tx) error {
		if _, err := rec.Record(ctx, tx, acc, e); err != nil {
			return err
		}
		return errAction
	})
	if !errors.Is(err, errAction) {
		t.Fatalf("got %v", err)
	}
	if n := countJobs(t, d); n != 0 {
		t.Fatalf("a rolled-back action left %d jobs", n)
	}

	var id uuid.UUID
	err = db.WithAccount(ctx, d.App, acc.String(), func(tx pgx.Tx) error {
		id, err = rec.Record(ctx, tx, acc, e)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if n := countJobs(t, d); n != 1 {
		t.Fatalf("jobs after commit = %d, want 1", n)
	}
	var raw []byte
	if err := d.Owner.QueryRow(ctx, "SELECT args FROM river_job WHERE kind = $1", DispatchArgs{}.Kind()).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var args DispatchArgs
	if err := json.Unmarshal(raw, &args); err != nil || args.EventID != id || args.AccountID != acc {
		t.Fatalf("job args %s: %+v, %v", raw, args, err)
	}
}

func TestEventNameForm(t *testing.T) {
	d := testdb.Start(t)
	acc := newAccount(t, d)
	rec := NewRecorder(insertOnly(t, d))
	err := db.WithAccount(t.Context(), d.App, acc.String(), func(tx pgx.Tx) error {
		_, err := rec.Record(t.Context(), tx, acc, Event{Name: "Thing Happened"})
		return err
	})
	if err == nil {
		t.Fatal("a name without the <module>.<event> form must fail")
	}
}

func TestDispatchWorker(t *testing.T) {
	d := testdb.Start(t)
	ctx := t.Context()
	acc := newAccount(t, d)
	rec := NewRecorder(insertOnly(t, d))
	var id uuid.UUID
	err := db.WithAccount(ctx, d.App, acc.String(), func(tx pgx.Tx) (err error) {
		id, err = rec.Record(ctx, tx, acc, Event{Name: "test.thing_happened", Data: map[string]int{"n": 1}})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	job := &river.Job[DispatchArgs]{Args: DispatchArgs{EventID: id, AccountID: acc}}

	t.Run("each sink gets the event", func(t *testing.T) {
		sink := &recordingSink{}
		if err := NewDispatchWorker(d.App, quiet, sink).Work(ctx, job); err != nil {
			t.Fatal(err)
		}
		got := sink.events()
		if len(got) != 1 || got[0].ID != id || got[0].Version != 1 || string(got[0].Data) != `{"n": 1}` {
			t.Fatalf("sink got %+v", got)
		}
	})

	t.Run("a failed sink fails the job, so River retries it", func(t *testing.T) {
		sink := &recordingSink{fail: errors.New("tool is down")}
		if err := NewDispatchWorker(d.App, quiet, sink).Work(ctx, job); err == nil {
			t.Fatal("want an error")
		}
	})

	t.Run("an event that is gone is not an error", func(t *testing.T) {
		gone := &river.Job[DispatchArgs]{Args: DispatchArgs{EventID: uuid.NewV7(), AccountID: acc}}
		if err := NewDispatchWorker(d.App, quiet, &recordingSink{}).Work(ctx, gone); err != nil {
			t.Fatal(err)
		}
	})
}

// The whole path with a real River client under the app role: the api
// records an event, and the worker process gives it to the sink.
func TestEndToEnd(t *testing.T) {
	d := testdb.Start(t)
	ctx := t.Context()
	acc := newAccount(t, d)
	sink := &recordingSink{}

	workers := river.NewWorkers()
	river.AddWorker(workers, NewDispatchWorker(d.App, quiet, sink))
	worker, err := river.NewClient(riverpgxv5.New(d.App), &river.Config{
		Logger:            quiet,
		Queues:            map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 1}},
		Workers:           workers,
		FetchPollInterval: 200 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Start(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = worker.Stop(context.WithoutCancel(ctx)) })

	rec := NewRecorder(insertOnly(t, d))
	err = db.WithAccount(ctx, d.App, acc.String(), func(tx pgx.Tx) error {
		_, err := rec.Record(ctx, tx, acc, Event{Name: "test.thing_happened"})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for len(sink.events()) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("the worker did not deliver the event in 10 s")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
