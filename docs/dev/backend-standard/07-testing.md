# 07 — Tests

**Status:** DRAFT
**Decisions:** B12

| Level | What it checks | Tool |
| --- | --- | --- |
| Unit | Pure domain logic without I/O: status transitions, calculations, validation rules | `go test` |
| Integration — the main level | A whole module with a real Postgres: sqlc queries, transactions, River jobs | testcontainers-go |
| Isolation | Account A does not see and does not change the data of account B, in each module | testcontainers-go |
| Contract | API responses match `openapi.yaml` | Response validation against the schema |
| Adapters | Parsing of webhooks and source responses | Recorded real responses |
| End-to-end | Spec scenarios through the public page and the dashboard | Playwright ([frontend standard](../frontend-standard/29-testing-strategy.md)) |

## Rules

- Do not mock the database: errors hide exactly in SQL, transactions and constraints.
- Each OpenSpec scenario links to at least one test or check. The PR shows the link ([05 of the frontend standard](../frontend-standard/05-spec-driven-development.md#6-traceability)).
- Write new code with TDD ([development process](../dev-process.md#cycle-for-each-task)): a failing test → the smallest fix → repeat.
- Fix a bug only after a test reproduces it.
- Each job handler has an idempotency test: two runs give the same result.

## Tools in the repository

| Tool | Where | What it does |
| --- | --- | --- |
| `testdb.Start(t)` | `core/internal/platform/testdb` | Starts Postgres 16 in Docker with the same roles script as `deploy/` and applies all migrations. Returns one pool per role: `Owner`, `App`, `System` |
| `testdb.AssertIsolation(t, d, case)` | the same | The isolation test for one table: account B does not see, change, delete or insert the rows of account A. Each module calls it for each of its tables |
| RLS check | `core/migrations/rls_test.go` | Fails when a table with `account_id` has no rule, or RLS is not enabled and forced ([ADR 0009](../adr/0009-account-isolation-rls.md)). It blocks the merge |
| `db.WithAccount` | `core/internal/platform/db` | The one entry point to account data. Tests use it as the real code does |
| Playwright | `web/e2e` | End-to-end tests. `mise run up`, then `mise run test:e2e`. `E2E_API_URL` selects the target |
| Prism | `mise run mock` | A mock API from `api/openapi.yaml` on port 4010, for frontend work before the backend is ready |

`mise run test` runs the Go tests with Postgres in Docker. It needs only Docker, no local Postgres.
