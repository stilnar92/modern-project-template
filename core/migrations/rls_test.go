package migrations_test

import (
	"context"
	"testing"

	"github.com/stilnar92/modern-project-template/core/internal/platform/testdb"
)

// TestRLS blocks the merge when a migration adds a table with account_id
// without a forced RLS rule (ADR 0009).
func TestRLS(t *testing.T) {
	d := testdb.Start(t)
	problems, err := testdb.CheckRLS(context.Background(), d.Owner)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range problems {
		t.Error(p)
	}
}
