// Package testdb starts a real Postgres for integration tests
// (backend standard 07). It uses the same roles script as the local
// environment and applies all migrations, so tests see the real schema.
package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/stilnar92/modern-project-template/core/migrations"
)

// DB is a migrated test database with one pool per role of ADR 0009.
type DB struct {
	Owner  *pgxpool.Pool // the schema owner: migrations and test setup
	App    *pgxpool.Pool // the api and the worker: RLS applies
	System *pgxpool.Pool // BYPASSRLS: only the marked places
}

var passwords = map[string]string{
	"app_owner":  "test-owner",
	"app_app":    "test-app",
	"app_system": "test-system",
}

// Start starts Postgres in Docker, applies the migrations and returns the
// pools. Docker stops the container when the test ends.
func Start(t testing.TB) *DB {
	t.Helper()
	ctx := context.Background()

	_, self, _, _ := runtime.Caller(0)
	initScript := filepath.Join(filepath.Dir(self), "..", "..", "..", "..", "deploy", "postgres", "init", "01-roles.sh")

	ctr, err := postgres.Run(ctx, "postgres:16.15-alpine",
		postgres.WithUsername("postgres"),
		postgres.WithPassword("test-admin"),
		postgres.WithDatabase("postgres"),
		postgres.WithInitScripts(initScript),
		testcontainers.WithEnv(map[string]string{
			"OWNER_DB_PASSWORD":  passwords["app_owner"],
			"APP_DB_PASSWORD":    passwords["app_app"],
			"SYSTEM_DB_PASSWORD": passwords["app_system"],
		}),
		postgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, ctr)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	host, err := ctr.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := ctr.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatal(err)
	}
	dsn := func(role string) string {
		return fmt.Sprintf("postgres://%s:%s@%s:%s/app?sslmode=disable", role, passwords[role], host, port.Port())
	}

	migrate(t, dsn("app_owner"))

	d := &DB{
		Owner:  pool(t, dsn("app_owner")),
		App:    pool(t, dsn("app_app")),
		System: pool(t, dsn("app_system")),
	}
	return d
}

func migrate(t testing.TB, dsn string) {
	t.Helper()
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = sqlDB.Close() }()
	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, migrations.FS)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

func pool(t testing.TB, dsn string) *pgxpool.Pool {
	t.Helper()
	p, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}
