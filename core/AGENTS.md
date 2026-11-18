# core (Go)

The rules for the Go core. The full standard: [docs/dev/backend-standard](../docs/dev/backend-standard/README.md).

## Commands (from the repo root)

- `mise run lint:core`: golangci-lint and `tools/archcheck` (import rules, table ownership, coupling report).
- `mise run test:core`: `go vet` and `go test`. Integration tests start Postgres in Docker through `internal/platform/testdb`.
- `mise run gen:core`: regenerate `internal/platform/openapi` from `api/openapi.yaml`, and the sqlc code of each module (`sqlc.yaml`).
- `mise run migrate`: apply migrations in the local environment.

## Layout

- `internal/<module>`: a core module. `internal/product/<module>`: a product module. `internal/platform/<name>`: shared infrastructure ([standard 01](../docs/dev/backend-standard/01-modules.md)).
- Other modules import only the root package of a module, never its subpackages. The core never imports a product module.
- Each module README has a line `Tables: a, b`. SQL in `<module>/queries/*.sql` uses only these tables. sqlc compiles it into the private `<module>/store` package: add each new module to `sqlc.yaml`. Do not edit the generated files.
- IDs are `uuid.NewV7()` from the standard library `uuid` package. `internal/accounts` is the reference module: README, queries, store, service, tests.

## Rules

- All access to account data goes through `db.WithAccount` (ADR 0009). The system role only in the places that ADR 0009 lists.
- Each table with `account_id` gets `ENABLE` and `FORCE ROW LEVEL SECURITY` and a rule with `nullif(current_setting('app.account_id', true), '')::uuid`. Each rule of the table uses this filter: rules add up with OR, so one open rule opens all rows. A view gets `security_invoker = true`. `migrations/rls_test.go` checks all of this.
- Each module calls `testdb.AssertIsolation` for each of its tables.
- Dashboard handlers get the account from `accounts.RequireSession` and `accounts.AccountID(ctx)`. A new way around RLS (the system role or a `SECURITY DEFINER` function) needs an ADR (ADR 0009, ADR 0013). Set the rights of a function before you change its owner.
- Do not mock the database. Use `testdb.Start(t)`. Write new code with TDD ([standard 07](../docs/dev/backend-standard/07-testing.md)).
- Migrations only add things: expand → migrate → contract ([standard 03](../docs/dev/backend-standard/03-data.md)).
- Do not edit `*.gen.go`. Change the contract and run `mise run gen`.
- Settings come only from environment variables. A process describes them as a struct that embeds `config.Base` and reads them once with `config.Load`. Add each new variable to `.env.example`.
- Log with the `*Context` methods (`logger.InfoContext(ctx, ...)`), so each record gets `request_id` and `trace_id`. Report errors with `observability.ReportError(ctx, ...)`.
- Background work is a River job. Register each job kind in `newRiver` in `cmd/worker/main.go`. The job arguments carry `account_id`, and the job reads data through `db.WithAccount`. Only the jobs that ADR 0009 lists get the system pool. A job must accept a second run.
- A list endpoint uses `internal/platform/cursor`: the package comment has the SQL pattern ([backend standard 02](../docs/dev/backend-standard/02-api.md#lists)).
- To record an event, call `events.Recorder.Record` with the transaction of the action, and add the name to `docs/dev/events.md` (ADR 0014).
- The api middleware order is in the `internal/platform/httpmw` package comment. `/healthz` is liveness only. `/readyz` checks the database. `/metrics` is on the internal `METRICS_PORT`, never through the proxy.
