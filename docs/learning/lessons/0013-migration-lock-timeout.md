# 0013 — A schema change that waits for a lock stops the whole application

From PR #89 · 11.11.2026 · Topics: Postgres, migrations, locks

## Problem

A migration linter (squawk) found that our migrations set no time limits. A schema change needs a lock on the table. If it cannot get the lock, it waits, and by default it waits with no limit. While it waits, the application can stop.

## Predict first

A report query reads the table `events` for two minutes. During that time, a migration runs `ALTER TABLE events ADD COLUMN ...`. The migration waits. What happens to a plain `SELECT` on `events` that arrives one second later?

## Theory

Postgres protects each table with locks of different strength. A plain `SELECT` takes the weakest lock. Many readers and writers hold their locks at the same time with no conflict. Most forms of `ALTER TABLE` need the strongest lock, `ACCESS EXCLUSIVE`. It conflicts with all other locks, also with the lock of a `SELECT`.

The surprise is the **lock queue**. A request for a lock that cannot be granted waits in a queue. Each later request that conflicts with a waiting request waits behind it. So the sequence is:

1. A long query holds a weak lock on the table.
2. The migration asks for `ACCESS EXCLUSIVE` and waits for the long query.
3. Each new query on the table, also a plain `SELECT`, waits behind the migration.

The migration itself did nothing yet, but the table is closed. The connection pool fills with waiting queries, and the application stops answering. Analogy: a slow truck waits to turn across the road, and all cars behind it wait too, although their lane is free.

Two settings limit the damage:

- `lock_timeout` cancels a statement that waits for a lock longer than the limit. The migration fails fast, the queue clears, and you try again later.
- `statement_timeout` cancels a statement that runs longer than the limit. It protects from a change that got the lock and then holds it for a long time.

`SET LOCAL` applies a setting only until the end of the current transaction. So the limits of a migration do not change the limits of the application.

Sources:

- [Explicit Locking](https://www.postgresql.org/docs/current/explicit-locking.html), PostgreSQL documentation: the lock modes and the table of conflicts.
- [Client Connection Defaults](https://www.postgresql.org/docs/current/runtime-config-client.html), PostgreSQL documentation: `lock_timeout` and `statement_timeout`.
- [Zero-downtime Postgres schema migrations need this: lock_timeout and retries](https://postgres.ai/blog/20210923-zero-downtime-postgres-schema-migrations-lock-timeout-and-retries), Nikolay Samokhvalov: the queue problem with pictures.
- [require-timeout-settings](https://squawkhq.com/docs/require-timeout-settings), squawk: the linter rule that we use.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| No limits | The migration waits as long as necessary | One long query plus one migration closes the table for all |
| Run migrations at night | Less traffic, fewer long queries | It only lowers the chance. A night job can be the long query |
| Set the limits for the migration role or in the tool | One setting covers all migrations | The file does not show it, and a linter that reads files cannot check it |
| `SET LOCAL lock_timeout` and `statement_timeout` in each migration | Each file carries its limits. A linter fails a file without them | Two lines in each section. A failed migration needs a second try |
| The limits plus automatic retries | The tool tries again after a pause | More code around the migration tool |

## Edge cases

| What if | What happens |
| --- | --- |
| The migration cannot get the lock in 5 seconds | Postgres cancels the statement. The migration transaction rolls back, and the schema does not change. Run it again |
| The change gets the lock, but the statement runs longer than 30 seconds | `statement_timeout` cancels it. The lock is released |
| A new migration has no time limits | `mise run lint:core` fails. The deliberate test was `ADD COLUMN NOT NULL` without the limits |
| The limits stay active after the migration | They do not. `SET LOCAL` ends with the transaction of the migration |
| The `Down` section | It sets the same two limits. A rollback also takes locks |
| An old migration needs the two lines | The rule says: do not edit a migration that was applied. The owner decided to add the lines to migration 00001. The round-trip test and the RLS test pass with them |

## Decision

Each migration starts with `SET LOCAL lock_timeout = '5s'` and `SET LOCAL statement_timeout = '30s'`, in the `Up` section and in the `Down` section. squawk checks each migration file in `mise run lint:core` and in CI. A migration that waits now fails fast and does not block the application.

We did not build retries. A failed migration is visible, and a person runs it again. This would be wrong for a change that cannot finish in 30 seconds on a large table, for example one that rewrites the table or fills a column. That change needs another shape: small steps and a background job ([backend standard 03](../../dev/backend-standard/03-data.md), "expand, migrate, contract").

Record: the row "squawk" in the gates table of `docs/dev/ai-standard/03-review-verification.md`.

## In our code

1. `core/migrations/00001_role_grants.sql`: the two `SET LOCAL` lines and the comment that explains them.
2. `core/migrations/00003_events.sql`: the same lines in a migration that creates a table.
3. `mise.toml`, the task `lint:core`: `squawk migrations/*.sql`.
4. `docs/dev/backend-standard/03-data.md`, "Migrations": the three steps for a change that is not small.

## Check yourself

1. The migration did not get its lock. Why do plain `SELECT` queries on the table wait?
2. Why `SET LOCAL` and not `SET`?
3. The migration fails with a lock timeout. Is that a problem?

<details><summary>Answers</summary>

1. The migration waits in the lock queue for the strongest lock. A new `SELECT` conflicts with that waiting request, so it goes into the queue behind the migration. It does not matter that the lock of the `SELECT` does not conflict with the long query that holds the table now.
2. `SET` changes the setting for the whole session. `SET LOCAL` changes it only until the transaction ends. The limits belong to this migration and must not stay on the connection.
3. No. It is the result that we want. The schema did not change, the queue is clear, and the application works. The other result was a migration that waits with no limit while all queries on the table wait behind it. Find the long query or wait, and run the migration again.

</details>
