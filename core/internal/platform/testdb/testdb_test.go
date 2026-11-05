package testdb

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

// One container for all tests of this package.
func TestTestDB(t *testing.T) {
	d := Start(t)
	ctx := context.Background()

	exec := func(t *testing.T, sql string) {
		t.Helper()
		if _, err := d.Owner.Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}
	insert := func(table string) func(context.Context, pgx.Tx, string) error {
		return func(ctx context.Context, tx pgx.Tx, accountID string) error {
			_, err := tx.Exec(ctx, "INSERT INTO "+table+" (account_id) VALUES ($1)", accountID)
			return err
		}
	}

	t.Run("RLS check fails on a table without a rule, passes with the rule", func(t *testing.T) {
		exec(t, "CREATE TABLE probe_rls (account_id uuid NOT NULL)")
		t.Cleanup(func() { _, _ = d.Owner.Exec(ctx, "DROP TABLE probe_rls") })

		steps := []struct {
			sql  string
			want int
		}{
			{"", 1}, // no RLS
			{"ALTER TABLE probe_rls ENABLE ROW LEVEL SECURITY", 1},
			{"ALTER TABLE probe_rls FORCE ROW LEVEL SECURITY", 1}, // no rule yet
			{`CREATE POLICY account_isolation ON probe_rls
			  USING (account_id = nullif(current_setting('app.account_id', true), '')::uuid)
			  WITH CHECK (account_id = nullif(current_setting('app.account_id', true), '')::uuid)`, 0},
		}
		for _, s := range steps {
			if s.sql != "" {
				exec(t, s.sql)
			}
			problems, err := CheckRLS(ctx, d.Owner)
			if err != nil {
				t.Fatal(err)
			}
			if len(problems) != s.want {
				t.Fatalf("after %q: got %v, want %d problems", s.sql, problems, s.want)
			}
		}
	})

	t.Run("RLS check fails on a rule that does not use the account", func(t *testing.T) {
		exec(t, "CREATE TABLE probe_rule (account_id uuid NOT NULL)")
		t.Cleanup(func() { _, _ = d.Owner.Exec(ctx, "DROP TABLE probe_rule CASCADE") })
		exec(t, "ALTER TABLE probe_rule ENABLE ROW LEVEL SECURITY")
		exec(t, "ALTER TABLE probe_rule FORCE ROW LEVEL SECURITY")
		exec(t, `CREATE POLICY account_isolation ON probe_rule
		  USING (account_id = nullif(current_setting('app.account_id', true), '')::uuid)`)

		steps := []struct {
			name, sql string
			want      int
		}{
			{"the correct rule", "", 0},
			// Policies add up with OR, so a second open policy opens all rows.
			{"an open second rule", "CREATE POLICY open ON probe_rule USING (true)", 1},
			{"open rule removed", "DROP POLICY open ON probe_rule", 0},
			// Without nullif, the empty setting after a transaction breaks the
			// next query on a pooled connection (ADR 0009).
			{"a rule without nullif", `CREATE POLICY no_nullif ON probe_rule FOR SELECT
			  USING (account_id = current_setting('app.account_id', true)::uuid)`, 1},
			{"rule without nullif removed", "DROP POLICY no_nullif ON probe_rule", 0},
			{"an open insert check", `CREATE POLICY open_insert ON probe_rule FOR INSERT
			  WITH CHECK (true)`, 1},
			{"open insert check removed", "DROP POLICY open_insert ON probe_rule", 0},
			// A view runs with the rights of its owner, and RLS does not apply.
			{"a view without security_invoker", "CREATE VIEW probe_view AS SELECT account_id FROM probe_rule", 1},
			{"the view with security_invoker", "ALTER VIEW probe_view SET (security_invoker = true)", 0},
		}
		for _, s := range steps {
			if s.sql != "" {
				exec(t, s.sql)
			}
			problems, err := CheckRLS(ctx, d.Owner)
			if err != nil {
				t.Fatal(err)
			}
			if len(problems) != s.want {
				t.Fatalf("%s: got %v, want %d problems", s.name, problems, s.want)
			}
		}
	})

	t.Run("isolation holds with the rule", func(t *testing.T) {
		exec(t, `CREATE TABLE probe_iso (account_id uuid NOT NULL)`)
		t.Cleanup(func() { _, _ = d.Owner.Exec(ctx, "DROP TABLE probe_iso") })
		exec(t, "ALTER TABLE probe_iso ENABLE ROW LEVEL SECURITY")
		exec(t, "ALTER TABLE probe_iso FORCE ROW LEVEL SECURITY")
		exec(t, `CREATE POLICY account_isolation ON probe_iso
			USING (account_id = nullif(current_setting('app.account_id', true), '')::uuid)
			WITH CHECK (account_id = nullif(current_setting('app.account_id', true), '')::uuid)`)
		AssertIsolation(t, d, IsolationCase{Table: "probe_iso", Insert: insert("probe_iso")})
	})

	t.Run("isolation check finds a leak without the rule", func(t *testing.T) {
		exec(t, `CREATE TABLE probe_leak (account_id uuid NOT NULL)`)
		t.Cleanup(func() { _, _ = d.Owner.Exec(ctx, "DROP TABLE probe_leak") })
		err := CheckIsolation(ctx, d, IsolationCase{Table: "probe_leak", Insert: insert("probe_leak")})
		if !errors.Is(err, errLeak) {
			t.Fatalf("got %v, want a leak", err)
		}
	})

	t.Run("a duplicate key does not count as isolation", func(t *testing.T) {
		exec(t, `CREATE TABLE probe_dup (id int PRIMARY KEY, account_id uuid NOT NULL)`)
		t.Cleanup(func() { _, _ = d.Owner.Exec(ctx, "DROP TABLE probe_dup") })
		exec(t, "ALTER TABLE probe_dup ENABLE ROW LEVEL SECURITY")
		exec(t, "ALTER TABLE probe_dup FORCE ROW LEVEL SECURITY")
		// Reads are isolated, but inserts are open to all accounts.
		exec(t, `CREATE POLICY p ON probe_dup
			USING (account_id = nullif(current_setting('app.account_id', true), '')::uuid)
			WITH CHECK (true)`)
		fixedKey := func(ctx context.Context, tx pgx.Tx, accountID string) error {
			_, err := tx.Exec(ctx, "INSERT INTO probe_dup (id, account_id) VALUES (1, $1)", accountID)
			return err
		}
		err := CheckIsolation(ctx, d, IsolationCase{Table: "probe_dup", Insert: fixedKey})
		if err == nil {
			t.Fatal("want an error: the second insert fails on the key, not on RLS")
		}
	})

	t.Run("app role cannot read without an account", func(t *testing.T) {
		var n int
		exec(t, `CREATE TABLE probe_none (account_id uuid NOT NULL)`)
		t.Cleanup(func() { _, _ = d.Owner.Exec(ctx, "DROP TABLE probe_none") })
		exec(t, "ALTER TABLE probe_none ENABLE ROW LEVEL SECURITY")
		exec(t, "ALTER TABLE probe_none FORCE ROW LEVEL SECURITY")
		exec(t, `CREATE POLICY p ON probe_none USING (account_id = nullif(current_setting('app.account_id', true), '')::uuid)`)
		// FORCE blocks the owner too, so the setup row goes in as the system role.
		if _, err := d.System.Exec(ctx, "INSERT INTO probe_none VALUES ($1)", AccountA); err != nil {
			t.Fatal(err)
		}
		if err := d.App.QueryRow(ctx, "SELECT count(*) FROM probe_none").Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Fatalf("app sees %d rows without an account", n)
		}
		if err := d.System.QueryRow(ctx, "SELECT count(*) FROM probe_none").Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatalf("system sees %d rows, want 1 (BYPASSRLS)", n)
		}
	})
}
