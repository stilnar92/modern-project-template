package testdb

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/stilnar92/modern-project-template/core/internal/platform/db"
)

// Two fixed accounts for isolation tests.
const (
	AccountA = "aaaaaaaa-0000-4000-8000-000000000001"
	AccountB = "bbbbbbbb-0000-4000-8000-000000000002"
)

// IsolationCase describes one table for the isolation check.
type IsolationCase struct {
	Table string
	// Insert adds one valid row of the account to the table. Each call must
	// make a new row (new keys), so that only RLS can refuse it.
	Insert func(ctx context.Context, tx pgx.Tx, accountID string) error
}

// rlsViolation is the SQLSTATE of "new row violates row-level security policy".
const rlsViolation = "42501"

// errLeak marks a broken isolation, so a test can tell it from an error
// of the setup.
var errLeak = errors.New("isolation broken")

// CheckIsolation checks that account B does not see and does not change the
// data of account A in the table (backend standard 07). It uses the app role
// and the one entry point, as the real code does.
func CheckIsolation(ctx context.Context, d *DB, c IsolationCase) error {
	if err := db.WithAccount(ctx, d.App, AccountA, func(tx pgx.Tx) error {
		return c.Insert(ctx, tx, AccountA)
	}); err != nil {
		return fmt.Errorf("insert as account A: %w", err)
	}

	table := pgx.Identifier{c.Table}.Sanitize()
	return db.WithAccount(ctx, d.App, AccountB, func(tx pgx.Tx) error {
		var n int
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM "+table+" WHERE account_id = $1", AccountA).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return fmt.Errorf("%w: account B sees %d rows of account A in %s", errLeak, n, c.Table)
		}
		for _, stmt := range []string{
			"UPDATE " + table + " SET account_id = account_id WHERE account_id = $1",
			"DELETE FROM " + table + " WHERE account_id = $1",
		} {
			tag, err := tx.Exec(ctx, stmt, AccountA)
			if err != nil {
				return err
			}
			if tag.RowsAffected() != 0 {
				return fmt.Errorf("%w: account B changes rows of account A in %s: %s", errLeak, c.Table, stmt)
			}
		}
		// An insert with the id of account A must fail. A savepoint keeps
		// the transaction usable after the expected error.
		sp, err := tx.Begin(ctx)
		if err != nil {
			return err
		}
		err = c.Insert(ctx, sp, AccountA)
		var pgErr *pgconn.PgError
		switch {
		case err == nil:
			return fmt.Errorf("%w: account B inserts a row of account A into %s", errLeak, c.Table)
		case !errors.As(err, &pgErr) || pgErr.Code != rlsViolation:
			// Another error (for example, a duplicate key) proves nothing.
			return fmt.Errorf("insert as account B: want an RLS error (%s), got: %w", rlsViolation, err)
		}
		return sp.Rollback(ctx)
	})
}

// AssertIsolation fails the test when CheckIsolation finds a problem.
func AssertIsolation(t testing.TB, d *DB, c IsolationCase) {
	t.Helper()
	if err := CheckIsolation(context.Background(), d, c); err != nil {
		t.Fatal(err)
	}
}
