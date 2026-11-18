# Operations

**Status:** local only. The core has no server yet: hosting is a later decision ([questions](questions.md#infrastructure-choices)). This document grows when the core gets a server.

## Local environment

The stack is like production: Postgres, S3 (SeaweedFS), Mailpit, the core image (api, worker, migrate) and Caddy ([deploy/README.md](../../deploy/README.md)). Tool versions come from `mise.toml`: run `mise install` once.

| What | Command |
| --- | --- |
| Start, with migrations | `mise run up`. It creates `.env` when it does not exist and waits until each healthcheck passes |
| Stop, keep the data | `mise run down` |
| Stop and delete the data | `mise run reset` (it asks first) |
| Logs | `mise run logs`, or `mise run logs api` for one service |
| Migrations | `mise run migrate`, `mise run migrate status`, `mise run migrate down` |
| Health | `docker compose -f deploy/compose.yaml --env-file .env ps`: each service shows `healthy` |
| Version of the running core | `docker compose -f deploy/compose.yaml --env-file .env exec api /app/api -version` |

**Addresses:** <http://api.localhost:8088> (core), <http://mail.localhost:8088> (Mailpit), <http://s3.localhost:8088> (S3). Postgres is on `localhost:54320`.

**Healthchecks.** The core image has no shell, so each binary checks itself. `/app/api -healthcheck` calls `/healthz` of the api. `/app/worker -healthcheck` calls `/healthz` on the metrics port of the worker. It answers 200 while the River client runs.

**Version.** `mise run up` builds the core image with `VERSION` from `git describe --always --dirty`. The version is in `api -version`, in each log line (`"version"`) and in Sentry as the release.

## A second worktree

`mise run env` (`deploy/worktree-env.sh`) gives each git worktree a slot number. The main checkout has slot 0. The slot list is in the shared git folder (`.git/app-worktree-slots`). A slot becomes free when its folder no longer exists. The new `.env` gets:

- `COMPOSE_PROJECT_NAME` = the folder name, so the worktree has its own containers and volumes;
- each port = the default plus the slot: `HTTP_PORT` (8088), `POSTGRES_PORT` (54320), `S3_PORT` (8333), `SMTP_PORT` (1025), `MOCK_PORT` (4010), `DASHBOARD_PORT` (5173), `SITE_PORT` (4330);
- `CORE_URL` and the database URLs with these ports.

So slot 2 has the core on <http://api.localhost:8090>. `mise run dev`, `mise run mock` and `mise run test:e2e` read the ports from `.env`. A worktree with an old `.env` keeps the old ports: delete its `.env` and run `mise run env` again.

## Backups and restore

| What | Command |
| --- | --- |
| Dump the local database | `mise run db:dump`. The file goes into `backups/`, which git ignores |
| Replace the local database with a dump | `mise run db:restore backups/<file>.dump`. It asks first and restores in one transaction, so a failed restore keeps the old data |

To check a dump without changes to the main database, restore it into a scratch database:

```sh
dc() { docker compose -f deploy/compose.yaml --env-file .env "$@"; }
dc exec -T postgres createdb -U postgres -O app_owner restore_check
dc exec -T postgres pg_restore -U postgres -d restore_check --single-transaction < backups/<file>.dump
dc exec -T postgres psql -U postgres -d restore_check -c "SELECT version_id, is_applied FROM goose_db_version"
dc exec -T postgres dropdb -U postgres restore_check
```

The S3 volume (`s3data`) has no backup task yet. It holds only local test files.

`TestDownAndUp` (`core/migrations/roundtrip_test.go`) runs each migration down and up again in `mise run test`, so a broken `Down` section fails before the merge.

## Errors and logs

- **Logs:** JSON on stdout, with the fields `service` and `version`. Caddy writes a JSON access log. Docker rotates the logs (10 MB, 3 files).
- **Error tracking:** set `SENTRY_DSN` in `.env`. Then api, worker and migrate send errors to Sentry. Without it, Sentry is off. The dashboard uses `VITE_SENTRY_DSN` in the same way.
- **API errors:** the client gets Problem Details (RFC 9457) without internal text. The real error is in the log and in Sentry ([backend standard 02](backend-standard/02-api.md)).

## When the core gets a server

The plan is in the [tech stack](tech-stack.md#infrastructure-and-operations): TLS with Caddy, scheduled backups with a restore check, an image registry, rollback to the previous image version, and the alerts of [backend standard 06](backend-standard/06-observability.md#alerts).
