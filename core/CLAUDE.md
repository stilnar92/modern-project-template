# core (Go)

The rules for the Go core. The full standard: [docs/dev/backend-standard](../docs/dev/backend-standard/README.md).

## Commands (from the repo root)

- `task lint:core`: golangci-lint and `tools/archcheck` (import rules, table ownership, coupling report).
- `task test:core`: `go vet` and `go test`. Integration tests start Postgres in Docker through `internal/platform/testdb`.
- `task gen:core`: regenerate `internal/platform/openapi` from `api/openapi.yaml`.
- `task migrate`: apply migrations in the local environment.

## Layout

- `internal/<module>`: a core module. `internal/product/<module>`: a product module. `internal/platform/<name>`: shared infrastructure ([standard 01](../docs/dev/backend-standard/01-modules.md)).
- Other modules import only the root package of a module, never its subpackages. The core never imports a product module.
- Each module README has a line `Tables: a, b`. SQL in `<module>/queries/*.sql` uses only these tables.

## Rules

- All access to account data goes through `db.WithAccount` (ADR 0009). The system role only in the places that ADR 0009 lists.
- Each table with `account_id` gets `ENABLE` and `FORCE ROW LEVEL SECURITY` and a rule with `nullif(current_setting('app.account_id', true), '')::uuid`. Each rule of the table uses this filter: rules add up with OR, so one open rule opens all rows. A view gets `security_invoker = true`. `migrations/rls_test.go` checks all of this.
- Each module calls `testdb.AssertIsolation` for each of its tables.
- Do not mock the database. Use `testdb.Start(t)`. Write new code with TDD ([standard 07](../docs/dev/backend-standard/07-testing.md)).
- Migrations only add things: expand → migrate → contract ([standard 03](../docs/dev/backend-standard/03-data.md)).
- Do not edit `*.gen.go`. Change the contract and run `task gen`.
- Settings come only from environment variables. Add each new one to `.env.example`.
