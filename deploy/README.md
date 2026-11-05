# deploy

A local environment like production ([ADR 0007](../docs/dev/adr/0007-monorepo-layout.md)).

## Commands

| Command | What it does |
| --- | --- |
| `task up` | Creates `.env` from `.env.example` if necessary, starts Postgres, S3 and Mailpit, applies migrations, then starts the api, the worker and Caddy |
| `task migrate` | Applies migrations. `task migrate -- status` shows them |
| `task logs -- api` | Shows the logs of one service |
| `task down` | Stops the environment. The data stays |
| `task reset` | Stops the environment and deletes its data |

## Addresses

| Address | Service |
| --- | --- |
| http://api.localhost:8088/healthz | Core API |
| http://mail.localhost:8088 | Mailpit: all emails that the core sends |
| http://s3.localhost:8088 | S3 API (SeaweedFS) |
| `localhost:54320` | Postgres, database `app` |

`*.localhost` goes to `127.0.0.1` without changes to the hosts file. To use other ports, change `.env`.

## What is in it

- **Postgres 16.** `postgres/init/01-roles.sh` creates the database and the three roles of [ADR 0009](../docs/dev/adr/0009-account-isolation-rls.md): `app_owner` (migrations), `app_app` (api and worker, RLS applies), `app_system` (`BYPASSRLS`, only marked places). The `postgres` user is a superuser for administration only, because superusers ignore RLS. Production creates the same roles before the first migration.
- **SeaweedFS** for S3. MinIO no longer publishes Docker images. The core uses only the S3 API, so the server can change.
- **Mailpit** instead of a real mail server.
- **Caddy** as the proxy. In production, it also issues the certificates.
- **The core image** (`core/Dockerfile`): one distroless image with the `api`, `worker` and `migrate` binaries. Compose selects the binary with `command`.

## Settings

All settings come from environment variables. The list with comments is in [.env.example](../.env.example). Errors go to Sentry only when `SENTRY_DSN` is set. Logs are JSON in stdout.

The compose file has no containers for the site and the dashboard yet. The app shells in `web/apps` run with `pnpm dev` in `web/` ([web/CLAUDE.md](../web/CLAUDE.md)). Add their containers, the dashboard hosting and the session cookie path (same origin or CORS) with the first feature that needs them.
