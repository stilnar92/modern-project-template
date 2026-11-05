package testdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// accountRule is the start of the expression that each rule must use
// (ADR 0009). Postgres prints the stored expression with NULLIF in capitals.
const accountRule = `%nullif(current_setting('app.account_id'%`

// rlsQuery finds each table in the public schema that has an account_id
// column: its RLS flags, the number of rules (policies), and the number of
// rules that do not use the account. Rules add up with OR, so one open rule
// opens all rows.
const rlsQuery = `
SELECT c.relname, c.relrowsecurity, c.relforcerowsecurity,
       (SELECT count(*) FROM pg_policies p
        WHERE p.schemaname = 'public' AND p.tablename = c.relname),
       (SELECT count(*) FROM pg_policies p
        WHERE p.schemaname = 'public' AND p.tablename = c.relname
          AND (coalesce(p.qual, '') NOT ILIKE $1
               OR (p.with_check IS NOT NULL AND p.with_check NOT ILIKE $1)))
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public'
  AND c.relkind IN ('r', 'p')
  AND EXISTS (
    SELECT 1 FROM pg_attribute a
    WHERE a.attrelid = c.oid AND a.attname = 'account_id' AND NOT a.attisdropped)
ORDER BY c.relname`

// viewQuery finds each view in the public schema without security_invoker.
// Such a view reads its tables with the rights of its owner, so RLS does not
// apply to the user of the view.
const viewQuery = `
SELECT c.relname
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = 'public'
  AND c.relkind = 'v'
  AND NOT coalesce(c.reloptions && ARRAY['security_invoker=true', 'security_invoker=on'], false)
ORDER BY c.relname`

// CheckRLS returns one problem for each table with account_id where RLS is
// not enabled and forced, that has no rule, or that has a rule without the
// account filter, and for each view without security_invoker (ADR 0009).
func CheckRLS(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	problems, err := checkTables(ctx, pool)
	if err != nil {
		return nil, err
	}
	views, err := checkViews(ctx, pool)
	if err != nil {
		return nil, err
	}
	return append(problems, views...), nil
}

func checkTables(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx, rlsQuery, accountRule)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var problems []string
	for rows.Next() {
		var table string
		var enabled, forced bool
		var rules, open int
		if err := rows.Scan(&table, &enabled, &forced, &rules, &open); err != nil {
			return nil, err
		}
		switch {
		case !enabled:
			problems = append(problems, fmt.Sprintf("%s: RLS is not enabled (ENABLE ROW LEVEL SECURITY)", table))
		case !forced:
			problems = append(problems, fmt.Sprintf("%s: RLS is not forced (FORCE ROW LEVEL SECURITY)", table))
		case rules == 0:
			problems = append(problems, fmt.Sprintf("%s: RLS has no rule (CREATE POLICY)", table))
		case open > 0:
			problems = append(problems, fmt.Sprintf(
				"%s: %d rule(s) do not filter by nullif(current_setting('app.account_id', true), '')::uuid", table, open))
		}
	}
	return problems, rows.Err()
}

func checkViews(ctx context.Context, pool *pgxpool.Pool) ([]string, error) {
	rows, err := pool.Query(ctx, viewQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var problems []string
	for rows.Next() {
		var view string
		if err := rows.Scan(&view); err != nil {
			return nil, err
		}
		problems = append(problems, fmt.Sprintf(
			"%s: a view without security_invoker ignores RLS (ALTER VIEW ... SET (security_invoker = true))", view))
	}
	return problems, rows.Err()
}
