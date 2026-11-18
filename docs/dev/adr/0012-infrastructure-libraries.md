# 0012 — Libraries for the infrastructure building blocks

Date: 2026-11-12
Status: accepted

## Context

The infrastructure audit of 12.11.2026 found building blocks that do not exist yet: HTTP middleware, sessions, config validation, metrics, IDs, the outbox, pagination, forms validation, i18n, design tokens and error handling ([questions](../questions.md#deferred-setup)). Each block needs a library or our own code. Agents must not select a library because it is popular. So two research passes compared at least three options for each block: maintenance, license, size, fit with the repository, and how well agents use it. The owner selected each option.

Choices that other ADRs already fix stay: pgx + sqlc, River, goose, slog, Sentry ([backend standard](../backend-standard/README.md)), Mantine and `@mantine/form` ([ADR 0008](0008-dashboard-mantine.md)), Tailwind + tailwind-variants for the public pages ([ADR 0001](0001-dashboard-ui-library.md)), react-router in data mode (the code already uses `createBrowserRouter`).

## Decision

### Core (Go)

| Block | Choice | Main reason | Rejected |
| --- | --- | --- | --- |
| Request ID, panic recovery, access log | Our own middleware (about 80 lines) and `sentryhttp` | A panic becomes a Problem response and a Sentry report | chi/middleware + httplog (plain 500), samber/slog-http (extra dependencies) |
| Sessions | Our own code on pgx + sqlc: a `sessions` table with the SHA-256 hash of the token, `account_id`, `expires_at` | The table keeps `account_id`, so SQL can revoke all sessions of an account, and the code fits RLS | alexedwards/scs (fixed table without `account_id`, plain token), gorilla/sessions |
| CSRF | `net/http.CrossOriginProtection` from the standard library, with `SameSite=Lax` | No dependency. The deny handler returns a Problem response | gorilla/csrf (token in the SPA, past CVE), filippo.io/csrf |
| Config | `github.com/caarlos0/env` | No dependencies. It checks all variables at startup and reports all errors | Our own `platform/config` only, sethvargo/go-envconfig |
| Metrics and HTTP instrumentation | OpenTelemetry SDK with the Prometheus exporter, `otelhttp`, `otelriver` | One API for traces and metrics, from HTTP through River | prometheus/client_golang directly |
| UUIDv7 | The `uuid` package of the Go 1.27 standard library | Official type, no dependency | google/uuid (no release since 2024), `uuidv7()` (needs Postgres 18) |
| Outbox | Our own `events` table and River `InsertTx` in the same transaction, with a periodic sweep | River already gives the transactional insert ([ADR 0003](0003-analytics-outbox-and-sinks.md)) | watermill-sql (a second queue next to River) |
| Cursor pagination | Our own helper: an opaque base64url cursor of `(created_at, id)` | No maintained library fits pgx + sqlc | — |

### Frontends

| Block | Choice | Main reason | Rejected |
| --- | --- | --- | --- |
| Form validation | zod 4 schemas generated from `api/openapi.yaml` by `@hey-api/openapi-ts` (zod plugin only), used through `@mantine/form` | The rules come from the contract and are not copied by hand | valibot (agents mix its API with zod), zod schemas by hand |
| i18n | Paraglide (inlang): one message package for the dashboard and the site | Compiled typed functions, no runtime on the site, Russian plurals through `Intl.PluralRules` | Lingui, i18next, our own code |
| Design tokens | Our own `packages/tokens` in TypeScript: one source gives the Mantine theme object and the CSS variables for the site | No dependency. The project does not use a design tool that needs DTCG files | Terrazzo, Style Dictionary |
| Error handling | Route `ErrorBoundary` of react-router and one `toProblem()` mapper in `api-client`. Field errors go to the form, service errors go to a Mantine notification | Built in. One place maps the Problem format | react-error-boundary (add it only when one widget needs isolation) |

## Trade-off

- **Prefer:** few dependencies, fit with the contract and RLS, code that agents use correctly.
- **Give up:** some ready-made code: we write and test the middleware, sessions, outbox and pagination ourselves.

## Stress round

| Stressor | Affects | Verdict | Resolution |
| --- | --- | --- | --- |
| sqlc cannot map the standard library `uuid.UUID` | IDs | accepted | Check it when the sqlc config is added. If it fails, use a `go_type` override or google/uuid, with a new ADR |
| hey-api is still before 1.0 and its API changes | Validation | accepted | Pin the version. Generate only zod schemas. Run it under TypeScript 6, as `api-client` does for openapi-typescript |
| The OTel Prometheus exporter is v0.x | Metrics | accepted | Pin the version. The exporter is one file of setup, so a change is cheap |
| Field errors need a place in the Problem schema | Error handling | accepted | Add an `errors` extension to `Problem` in the contract. Only the orchestrator changes it |

## Consequences

- Each block gets its tests when it is built. The session and outbox code need integration tests with `testdb`.
- A new library for one of these blocks needs a new ADR.
- The research notes are in the PR of this ADR.
