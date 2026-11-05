// Command migrate applies the database migrations. It is a separate step
// before the start of the api and the worker (issue #25).
//
// Usage: migrate [up|down|status|version] (default: up)
// It connects as the schema owner: MIGRATE_DATABASE_URL.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
	"github.com/pressly/goose/v3"

	"github.com/stilnar92/modern-project-template/core/internal/platform/config"
	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
	"github.com/stilnar92/modern-project-template/core/migrations"
)

func main() {
	logger, flush, err := observability.Setup("migrate")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer flush()
	if err := run(context.Background()); err != nil {
		observability.ReportError(logger, "migrate failed", err)
		flush()
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	dsn, err := config.Required("MIGRATE_DATABASE_URL")
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
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
