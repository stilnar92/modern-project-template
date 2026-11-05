# Tech stack

The public site on Astro, the dashboard on Vite + React, the core in Go, Postgres for data and jobs. Tool versions are in `mise.toml`. The containers are in the [architecture](architecture.md).

## Overview

| Part | Technology | What it does |
| --- | --- | --- |
| Core | Go | API, background jobs (River), integrations |
| Public site | Node (Astro, SSR) | Renders public pages from the core API data |
| Dashboard | Vite + React + Mantine ([ADR 0008](adr/0008-dashboard-mantine.md)) | The app for signed-in users, from standard components |
| Data and infrastructure | PostgreSQL, S3 (SeaweedFS locally), Caddy | Data, the River queue, files, reverse proxy and TLS |

One contract, `api/openapi.yaml`, generates the Go server code and the TypeScript client (`web/packages/api-client`) for the site and the dashboard.

## Frontend

One TypeScript monorepo (pnpm workspace in `web/`): `apps/dashboard`, `apps/site`, `packages/*`, `e2e/`.

**Frontend standard.** The whole frontend follows the [frontend standard](frontend-standard/README.md): specs before code, API types only from code generation, UI separate from logic, Tailwind + tailwind-variants, tokens as CSS variables, fast tooling (tsgo, oxlint, oxfmt), observability from day one. The standard applies in full. A deviation needs an ADR:

| Deviation | Status |
| --- | --- |
| The dashboard on a ready-made UI library (Mantine) instead of our own kit on Tailwind + tailwind-variants | [ADR 0001](adr/0001-dashboard-ui-library.md), [ADR 0008](adr/0008-dashboard-mantine.md) |
| The public site on Astro with React islands: the standard describes a React app | An ADR is necessary: which rules of the standard apply to Astro pages |
| Specs in the OpenSpec format, but Pocock's skills write them, not the OpenSpec commands | [ADR 0002](adr/0002-specs-openspec-hybrid.md) |
| ADRs are in `docs/dev/adr/` instead of `docs/adr/` | [ADR 0004](adr/0004-adr-location.md) |
| The AI standard replaces group B (agent work) for the whole repository | [ADR 0006](adr/0006-ai-standard-project-wide.md) |

| Module | Technologies | Why | Alternative |
| --- | --- | --- | --- |
| Public site | Astro in SSR mode (`@astrojs/node`), Tailwind CSS, React islands where a page needs interaction | Light pages with almost no JS, good for SEO and slow mobile internet, independent of the dashboard | Next.js |
| Dashboard | Vite + React SPA on Mantine, TanStack Query, React Router, React Compiler, Sentry | Many ready-made components, so it is fast to write. No SEO is necessary | Next.js |
| API client | openapi-typescript + openapi-fetch | Types come only from the contract | — |
| Tests | Vitest, Testing Library, Playwright (`web/e2e`), Prism mock server from the contract | Unit, component and end-to-end tests. The frontend can work before the backend exists | — |

The site has no Tailwind and no React islands yet. Add them with the first real page.

## Backend and data

| Module | Technologies | Why | Alternative |
| --- | --- | --- | --- |
| API and core | Go 1.27+, `net/http` (method and path patterns of Go 1.22+) | Little memory per client, one binary, stable for years | chi |
| API contract | `openapi.yaml` + oapi-codegen (Go) + openapi-typescript (frontend) | One source of truth for all clients | — |
| Database access | pgx + sqlc | SQL by hand, typed Go code, no ORM | GORM, ent |
| Migrations | goose (SQL files embedded in the `migrate` binary) | Versioned migrations. A test runs each Down section | Atlas |
| Main database | PostgreSQL 16 with RLS ([ADR 0009](adr/0009-account-isolation-rls.md)) | Data, account isolation, the River queue | — |
| Background jobs and cron | River (a queue on Postgres) | No separate broker. The job goes into the same transaction as the request | Asynq (Redis) |
| Cache and limits | Redis or Valkey, only when Postgres is not sufficient | One less service at the start | — |
| Files | S3-compatible storage (aws-sdk-go-v2) | Any provider, SeaweedFS locally | — |
| Login | A one-time code by email, a session in an httpOnly cookie | No passwords, no tokens in the browser | OAuth provider |
| Integrations | Own HTTP clients, webhooks with a signature and idempotency | Simple APIs need no SDK | — |
| Observability | slog (JSON), OpenTelemetry, Prometheus, Sentry | Logs, traces, metrics, errors | — |
| Tests | testcontainers-go | Tests with a real Postgres | — |

What is in `core/go.mod` now: pgx, goose, sentry-go, testcontainers-go. Add sqlc, River and the other libraries with the first real use ([questions](questions.md#deferred-setup)).

## Infrastructure and operations

| Task | Technologies | Note |
| --- | --- | --- |
| Hosting | Not selected | Keep the first deploy fast: everything is in Docker |
| Containers | Docker, Docker Compose | One or two servers at the start. No Kubernetes until the load needs it |
| Proxy and TLS | Caddy with Let's Encrypt | Local Caddy is HTTP only |
| DNS | Registrar or cloud DNS with an API | Necessary for a wildcard certificate (DNS-01) |
| CI | GitHub Actions (`.github/workflows/ci.yml`) | Lint, tests, code generation check, builds, audits |
| Errors | Sentry (own server) or GlitchTip | Frontend and backend |
| Metrics and logs | Prometheus, Grafana, Loki | Queue length, error rate, response time |
| Backups | Daily Postgres dumps to storage | Check a restore once a month |

**Requirements for the server choice:**

- [ ] Managed PostgreSQL with backups, or our own on a VM
- [ ] S3-compatible storage with lifecycle rules and a CDN
- [ ] DNS with an API
- [ ] One VM is sufficient at the start: api, worker, site, Caddy, Postgres
- [ ] We can move: everything in Docker, no dependency on the services of one provider

### Platform services

| Service | How it is deployed | Scaling |
| --- | --- | --- |
| Core API | A Go binary in a distroless image behind Caddy | Several replicas. State only in Postgres |
| Worker | A separate container from the same image | Separate queues by job type, so slow jobs do not delay fast ones |
| Scheduler | Periodic River jobs in the worker | — |
| Public site | One Astro Node server | Several replicas behind Caddy or a CDN |
| Dashboard | Vite static files in S3 + CDN, or served by Caddy | File names have hashes, so the cache is permanent |

### Environments, CI and secrets

- **Environments:** local (Docker Compose), staging and production.
- **Monorepo** by [ADR 0007](adr/0007-monorepo-layout.md): `api/openapi.yaml`, `core/` (Go: `cmd/api`, `cmd/worker`, `cmd/migrate`, `internal/<module>`), `web/` (pnpm workspace: `apps/dashboard`, `apps/site`, `packages/*`), `deploy/`. Shared commands are in `Taskfile.yml`.
- **Database migrations** are a separate deploy step, and only backward compatible.
- **Platform secrets** are in the CI secret store or SOPS. **Account secrets** (third-party tokens and keys) are encrypted in the database. The encryption key is outside the database.
- **Observability:** queue length, undelivered webhooks and certificate expiry, with alerts.

## API conventions

- REST, with the version in the path prefix: `/v1/...`.
- Resources are plural nouns, without extra segments and implementation details: `/v1/dashboard/projects`, `/v1/dashboard/projects/{id}`.
- Cursor pagination through query parameters: `?cursor=...&limit=20`.
- A long operation is asynchronous: `POST` returns a job ID; `GET .../{id}` or SSE gives the status and the result.
- Three API sections ([ADR 0005](adr/0005-api-surfaces-and-ownership.md)): `/v1/dashboard/...` with the account from the session; `/v1/public/...` for the public site, with the account from the address; `/v1/webhooks/<source>` for external sources, with a signature check and idempotency. The API does not accept `account_id` from the browser.
- Compatibility: inside `/v1`, changes only add things. To remove or rename a field, use a deprecation period.
- Errors are Problem Details by RFC 9457 (`application/problem+json`). The API implementation rules are in the [backend standard](backend-standard/02-api.md).
