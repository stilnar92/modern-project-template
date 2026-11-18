# 0014 — River tables, the outbox without a sweep, and a plain events table

Date: 2026-11-12
Status: accepted. It refines [ADR 0012](0012-infrastructure-libraries.md) (outbox) and defers the partitions of [backend standard 03](../backend-standard/03-data.md#queries)

## Context

ADR 0012 selects River and an outbox of our own: an `events` table, a River job inserted in the same transaction (`InsertTx`), and a periodic sweep. Backend standard 03 says that the `events` log is partitioned by month. Three questions came up when we built it:

1. How do the River tables get into the database? goose runs our migrations, and the app role cannot change the schema.
2. Is the sweep necessary?
3. Who creates the monthly partitions? Only the owner role can run DDL, and the Postgres image has no `pg_partman`.

## Decision

- **River tables.** The `migrate` command applies them with River's own migrator (`rivermigrate`) after goose, as the schema owner. River keeps its version in `river_migration`, so a River upgrade brings its migrations, and nobody copies its SQL. `down` and `status` work on the goose migrations only. The River tables have no account data and no RLS (ADR 0009).
- **No sweep.** The event row and its job commit in one transaction, so no event is lost. A job that fails retries with a back-off. After its last attempt River marks it "discarded", and the dead-letter alert of backend standard 06 watches it. So the log has no `dispatched_at` column and no updates.
- **A plain `events` table for now.** An index on `(account_id, occurred_at)`. We partition it by month when one of these is true: the table has more than 10 million rows, a report on it gets slow, or a retention rule must delete old months. The migration then creates a partitioned table and moves the rows. This is cheap while the table is small.
- **Jobs that ADR 0009 lists** (scheduled deletion) get the system pool in the worker. All other jobs use the app pool and `db.WithAccount` with the `account_id` from the job arguments.

## Trade-off

- **Prefer:** one migration path that the owner runs, fewer moving parts, the guarantees that River already gives.
- **Give up:** the partitions of standard 03 for now, and a second migration tool next to goose (River's own, which we do not edit).

## Stress round

| Stressor | Affects | Verdict | Resolution |
| --- | --- | --- | --- |
| A sink is down for a day | Dispatch | survives | The jobs retry with a back-off. After the last attempt they are discarded and the alert fires. An operator retries them |
| The events table grows fast in the pilot | Reports | accepted | The trigger above starts the partition migration |
| A River upgrade changes its schema | Migrations | survives | `mise run migrate` applies River's new migrations. The tests apply them too (`testdb`) |
| A job runs twice after a crash | Sinks | accepted | A sink must accept the same event twice (README of `internal/events`) |

## Consequences

- `cmd/worker` registers each job kind in `newRiver`.
- The api uses an insert-only River client for `events.Recorder` when the first handler records an event.
- Queue length and the age of the oldest job (standard 06) need a metric of their own: otelriver gives job counts and durations, not queue state. It comes with the alerts.
