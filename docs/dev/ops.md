# Operations

**Status:** local only. The core has no server yet: hosting is a later decision ([questions](questions.md#infrastructure-choices)). This document grows when the core gets a server.

## Local environment

The stack is like production: Postgres, S3 (SeaweedFS), Mailpit, the core image (api, worker, migrate) and Caddy ([deploy/README.md](../../deploy/README.md)). Tool versions come from `mise.toml`: run `mise install` once.

| What | Command |
| --- | --- |
| Start, with migrations | `task up`. It creates `.env` when it does not exist and waits until each healthcheck passes |
| Stop, keep the data | `task down` |
| Stop and delete the data | `task reset` (it asks first) |
| Logs | `task logs`, or `task logs -- api` for one service |
| Migrations | `task migrate`, `task migrate -- status`, `task migrate -- down` |
| Health | `docker compose -f deploy/compose.yaml --env-file .env ps`: each service except the worker shows `healthy` |
| Version of the running core | `docker compose -f deploy/compose.yaml --env-file .env exec api /app/api -version` |

**Addresses:** http://api.localhost:8088 (core), http://mail.localhost:8088 (Mailpit), http://s3.localhost:8088 (S3). Postgres is on `localhost:54320`.

**Healthchecks.** The api image has no shell, so the api binary checks itself: `/app/api -healthcheck` calls `/healthz`. The worker has no healthcheck until it has an HTTP endpoint.

**Version.** `task up` builds the core image with `VERSION` from `git describe --always --dirty`. The version is in `api -version`, in each log line (`"version"`) and in Sentry as the release.

## A second worktree

`task env` sets `COMPOSE_PROJECT_NAME` to the folder name, so a new worktree gets its own containers and volumes. Ports do not change on their own: in the new `.env`, set other values for `HTTP_PORT`, `POSTGRES_PORT`, `S3_PORT` and `SMTP_PORT`. If you forget, `task up` stops with "port is already allocated", and no data mixes. The web dev servers use fixed ports: 5173 (dashboard), 4330 (site), 4010 (Prism mock).

## Backups and restore

| What | Command |
| --- | --- |
| Dump the local database | `task db:dump`. The file goes into `backups/`, which git ignores |
| Replace the local database with a dump | `task db:restore -- backups/<file>.dump`. It asks first and restores in one transaction, so a failed restore keeps the old data |

To check a dump without changes to the main database, restore it into a scratch database:

```sh
dc() { docker compose -f deploy/compose.yaml --env-file .env "$@"; }
dc exec -T postgres createdb -U postgres -O app_owner restore_check
dc exec -T postgres pg_restore -U postgres -d restore_check --single-transaction < backups/<file>.dump
dc exec -T postgres psql -U postgres -d restore_check -c "SELECT version_id, is_applied FROM goose_db_version"
dc exec -T postgres dropdb -U postgres restore_check
```

The S3 volume (`s3data`) has no backup task yet. It holds only local test files.

`TestDownAndUp` (`core/migrations/roundtrip_test.go`) runs each migration down and up again in `task test`, so a broken `Down` section fails before the merge.

## Errors and logs

- **Logs:** JSON on stdout, with the fields `service` and `version`. Caddy writes a JSON access log. Docker rotates the logs (10 MB, 3 files).
- **Error tracking:** set `SENTRY_DSN` in `.env`. Then api, worker and migrate send errors to Sentry. Without it, Sentry is off. The dashboard uses `VITE_SENTRY_DSN` in the same way.
- **API errors:** the client gets Problem Details (RFC 9457) without internal text. The real error is in the log and in Sentry ([backend standard 02](backend-standard/02-api.md)).

## When the core gets a server

The plan is in the [tech stack](tech-stack.md#infrastructure-and-operations): TLS with Caddy, scheduled backups with a restore check, an image registry, rollback to the previous image version, and the alerts of [backend standard 06](backend-standard/06-observability.md#alerts).
