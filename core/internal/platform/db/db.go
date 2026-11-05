// Package db is the one entry point to account data (ADR 0009).
package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithAccount runs fn in a transaction that sees only the rows of one
// account. It sets app.account_id for this transaction only, so the pooled
// connection does not keep it. The RLS rules of each table read the value.
func WithAccount(ctx context.Context, pool *pgxpool.Pool, accountID string, fn func(pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT set_config('app.account_id', $1, true)", accountID); err != nil {
			return err
		}
		return fn(tx)
	})
}
