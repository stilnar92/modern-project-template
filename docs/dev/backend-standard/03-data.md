# 03 — Data and migrations

**Status:** DRAFT
**Decisions:** B5, B6

## Schema

- Each table with account data has `account_id NOT NULL`.
- Foreign keys between account data are composite, `(account_id, id)`. Unique constraints include `account_id` where uniqueness applies inside one account ([ADR 0005](../adr/0005-api-surfaces-and-ownership.md)).
- IDs are UUIDv7: they are ordered by time, work well in indexes and do not show the number of records.
- Time is `timestamptz`; money is `bigint` in minor units (for example, cents).
- Attributes that change between customer segments are `jsonb`, and the code checks their schema. We do not build a field builder.
- Each table with `account_id` gets an RLS rule (`ENABLE` + `FORCE` + a policy on `app.account_id`) in the same migration that creates the table ([ADR 0009](../adr/0009-account-isolation-rls.md)).
- Delete the personal data of users physically when its period ends. Do not use "soft delete" for it.

## Queries

- Write SQL by hand; sqlc compiles it. Do not use an ORM.
- Each query to account data filters by `account_id`. Indexes start with `account_id`.
- A new query on a hot path needs `EXPLAIN` in the PR description. The basis is [SQL Performance Explained](https://sql-performance-explained.com/): column order in an index, selectivity, sorting by index.
- Paginate with a cursor on `(created_at, id)`, not with `OFFSET`.
- The `events` log is partitioned by month ([tech stack](../tech-stack.md#backend-and-data)). For now it is a plain table: [ADR 0014](../adr/0014-river-and-outbox.md) gives the trigger for the partitions.

## Transactions

- Put the action, its outbox event and its River jobs in one transaction ([04](04-async-integrations.md)).
- External calls inside a transaction are forbidden: the network must not hold locks.

## Migrations

- Only backward-compatible migrations, in three steps: **expand** (a new column or table) → **migrate** (the code writes to both, a job fills in the data) → **contract** (a separate release removes the old part).
- Do not edit a migration that was applied. Write a new one.
- Check each migration on a copy of the production database before the deploy ([development process](../dev-process.md#cycle-for-each-task)).
- Migrations and the schema are shared files: only the orchestrator changes them ([AI standard](../ai-standard/01-workflow.md#parallel-work)).
