package migrations_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/stilnar92/modern-project-template/core/internal/platform/testdb"
	"github.com/stilnar92/modern-project-template/core/migrations"
)

// schemaQuery lists what migrations change: tables, columns, rules,
// indexes and the rights on the schema. Two equal results mean the same schema.
const schemaQuery = `
SELECT coalesce(string_agg(x, E'\n' ORDER BY x), '') FROM (
  SELECT format('table %s', table_name) FROM information_schema.tables WHERE table_schema = 'public'
  UNION ALL
  SELECT format('column %s.%s %s null=%s', table_name, column_name, data_type, is_nullable)
  FROM information_schema.columns WHERE table_schema = 'public'
  UNION ALL
  SELECT format('rule %s.%s %s %s', tablename, policyname, qual, with_check) FROM pg_policies WHERE schemaname = 'public'
  UNION ALL
  SELECT format('index %s', indexdef) FROM pg_indexes WHERE schemaname = 'public'
  UNION ALL
  SELECT format('schema rights %s', nspacl) FROM pg_namespace WHERE nspname = 'public'
  UNION ALL
  SELECT format('default rights %s %s', defaclobjtype, defaclacl) FROM pg_default_acl
) s(x)`

func schema(t *testing.T, db *sql.DB) string {
	t.Helper()
	var s string
	if err := db.QueryRowContext(context.Background(), schemaQuery).Scan(&s); err != nil {
		t.Fatal(err)
	}
	return s
}

// TestDownAndUp rolls back all migrations and applies them again. Each Down
// section must run, and the schema after the round trip must be the same
// (backend standard 03).
func TestDownAndUp(t *testing.T) {
	d := testdb.Start(t) // applies all migrations
	ctx := context.Background()
	db := stdlib.OpenDBFromPool(d.Owner)
	defer func() { _ = db.Close() }()
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		t.Fatal(err)
	}

	before := schema(t, db)
	if _, err := provider.DownTo(ctx, 0); err != nil {
		t.Fatalf("down to version 0: %v", err)
	}
	if v, err := provider.GetDBVersion(ctx); err != nil || v != 0 {
		t.Fatalf("version after down = %d, %v; want 0", v, err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("up after down: %v", err)
	}
	if after := schema(t, db); after != before {
		t.Fatalf("the schema changed after down and up.\nbefore:\n%s\nafter:\n%s", before, after)
	}
}
