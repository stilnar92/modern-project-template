package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

// River applies the tables of the River job queue (ADR 0012). River keeps
// its own version table (river_migration), so a River upgrade brings its
// migrations with it, and nobody copies its SQL. Run it as the schema owner
// after goose: the app role cannot change the schema. The River tables have
// no account data and no RLS rules (ADR 0009).
func River(ctx context.Context, pool *pgxpool.Pool) ([]int, error) {
	m, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		return nil, fmt.Errorf("river migrator: %w", err)
	}
	res, err := m.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return nil, fmt.Errorf("river migrate: %w", err)
	}
	versions := make([]int, 0, len(res.Versions))
	for _, v := range res.Versions {
		versions = append(versions, v.Version)
	}
	return versions, nil
}
