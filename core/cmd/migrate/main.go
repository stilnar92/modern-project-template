// Command migrate applies the database migrations. It is a separate step
// before the start of the api and the worker.
//
// Usage: migrate [up|down|status|version] (default: up)
// It connects as the schema owner: MIGRATE_DATABASE_URL. "up" applies the
// goose migrations, then the River tables (migrations.River). "down",
// "status" and "version" work on the goose migrations only.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/stilnar92/modern-project-template/core/internal/platform/config"
	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
	"github.com/stilnar92/modern-project-template/core/migrations"
)

// migrateConfig is the environment of migrate (.env.example).
type migrateConfig struct {
	config.Base
	DatabaseURL string `env:"MIGRATE_DATABASE_URL,required,notEmpty"`
}

func main() {
	cfg, err := config.Load[migrateConfig]()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx := context.Background()
	tel, err := observability.Setup(ctx, "migrate", cfg.Base)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := run(ctx, cfg.DatabaseURL); err != nil {
		observability.ReportError(ctx, tel.Logger, "migrate failed", err)
		tel.Shutdown(ctx)
		os.Exit(1)
	}
	tel.Shutdown(ctx)
}

func run(ctx context.Context, dsn string) error {
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer pool.Close()
	db := stdlib.OpenDBFromPool(pool)
	defer func() { _ = db.Close() }()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		return err
	}
	switch command {
	case "up":
		results, err := provider.Up(ctx)
		for _, r := range results {
			fmt.Printf("applied %s (%s)\n", r.Source.Path, r.Duration)
		}
		if err == nil && len(results) == 0 {
			fmt.Println("no new migrations")
		}
		if err != nil {
			return err
		}
		versions, err := migrations.River(ctx, pool)
		if len(versions) > 0 {
			fmt.Printf("applied river migrations %v\n", versions)
		}
		return err
	case "down":
		r, err := provider.Down(ctx)
		if r != nil {
			fmt.Printf("rolled back %s\n", r.Source.Path)
		}
		return err
	case "status":
		statuses, err := provider.Status(ctx)
		for _, s := range statuses {
			fmt.Printf("%-8s %s\n", s.State, s.Source.Path)
		}
		return err
	case "version":
		v, err := provider.GetDBVersion(ctx)
		fmt.Println(v)
		return err
	default:
		return errors.New("unknown command " + command + ": use up, down, status or version")
	}
}
