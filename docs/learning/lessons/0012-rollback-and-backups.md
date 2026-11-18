# 0012 — Can we go back: rollbacks and backups that somebody really tried

From PR #65 · 04.11.2026 · Topics: migrations, operations

## Problem

Each migration has a `Down` section that undoes it, but nothing ran these sections. The database also had no dump and no restore procedure. We had two ways back on paper, and no proof that one of them works.

## Predict first

A release goes wrong. You run the down migration for the first time, and it stops with a SQL error in the middle. What state is the database in, and how much time do you have to fix it?

## Theory

Code is easy to roll back: you deploy the previous version. Data is not. There are two ways back for a database:

- **A down migration** is the SQL that undoes one schema change. It is fast and exact. It cannot bring back data that the change removed.
- **A backup** is a copy of the data at one moment. A restore brings back all of it, and you lose each write after that moment.

Both have the same weak point: you use them rarely, and only when something is already wrong. Code that never runs breaks silently. A `Down` section with a mistake stays in the repository for months. A backup job writes files that nobody can read back. So the rule for both is the same: **a way back that nobody tried is not a way back**. Run the down migration in a test. Restore the backup into a real database and look at the result.

For Postgres, `pg_dump` writes a logical dump: the commands and data that build the database again. The custom format lets `pg_restore` select and reorder what it restores. `pg_restore --single-transaction` makes the restore one transaction: all of it works, or nothing changes.

Analogy: a Django migration with a `reverse_code` that nobody called. The migration looks reversible until the day you run `migrate` backward.

Sources:

- "Site Reliability Engineering" (Google), chapter 26, [Data Integrity: What You Read Is What You Wrote](https://sre.google/sre-book/data-integrity/): people need restores, not backups.
- [Postmortem of database outage of January 31](https://about.gitlab.com/blog/postmortem-of-database-outage-of-january-31/), GitLab: several backup methods existed, and none worked when it was needed.
- [SQL Dump](https://www.postgresql.org/docs/current/backup-dump.html), PostgreSQL documentation: `pg_dump` and how to restore a dump.
- [pg_restore](https://www.postgresql.org/docs/current/app-pgrestore.html), PostgreSQL documentation: `--clean`, `--if-exists`, `--single-transaction`.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| No down migrations, fix forward only | A mistake gets a new migration | No fast way back during an incident |
| Down migrations that nobody runs | The `Down` section exists | It can be broken, and you learn that at the worst moment |
| A round-trip test | A test migrates down to zero and up again, and compares the schema | A few seconds in the test run. It proves the schema, not the data |
| A dump task without a restore check | A file appears in a folder | You do not know if the file restores |
| A dump and a restore that we tried | We restored the dump into a scratch database and into the main one | Manual work now. It needs a schedule when there is a server |

## Edge cases

| What if | What happens |
| --- | --- |
| A `Down` section has a SQL error | The test fails at the step "down to version 0" |
| A `Down` section runs but removes too much, for example a right of the owner | The schema snapshot after the round trip differs, and the test shows the difference |
| A `Down` section drops a column that has data | The schema comes back and the data does not. The round-trip test cannot see this, because it compares schemas |
| The restore fails in the middle | The restore is one transaction, so the old data stays |
| Somebody runs the restore without a file name | The task stops before it touches the database. It also asks for confirmation first |
| You want to check a dump and keep the main database | Restore it into a scratch database, read the migration versions, and drop it |
| The file storage (S3) | It has no backup task yet. It holds only local test files |

## Decision

We test both ways back. `TestDownAndUp` takes a snapshot of the schema: tables, columns, RLS rules, indexes and rights. It migrates down to version 0, migrates up, and compares. It runs in `mise run test`, so a broken `Down` section fails before the merge. `mise run db:dump` and `mise run db:restore` exist, and the PR shows a real dump restored into a scratch database and into the main one.

This is not enough for production. There is no scheduled backup, because the core has no server yet. The plan for the server is a daily dump and a restore check once a month. A down migration is also the wrong tool when the change removed data. Then the way back is the backup, or a schema change in steps that never removes data in the same release ([lesson 0014](0014-breaking-api-changes.md) has the same idea for the API).

Record: the section "Backups and restore" in `docs/dev/ops.md`.

## In our code

1. `core/migrations/roundtrip_test.go`, `TestDownAndUp`: the snapshot query and the round trip.
2. `core/migrations/00003_events.sql`: an `Up` section and the `Down` section that the test runs.
3. `mise.toml`, the tasks `db:dump` and `db:restore`: the commands and the confirmation.
4. `docs/dev/ops.md`, "Backups and restore": how to check a dump in a scratch database.

## Check yourself

1. Why does the round-trip test compare a schema snapshot, and not only check that `Down` runs without an error?
2. Why is the restore one transaction?
3. The round-trip test is green. Which rollback problem can still happen in production?

<details><summary>Answers</summary>

1. A `Down` section can run with no error and still be wrong. It can leave an index, or remove a right that it did not add. Then "down and up" gives a different database. Only a comparison of the result finds this.
2. A restore with `--clean` drops each object before it creates it again. If it stops in the middle without a transaction, the database is half old and half empty. In one transaction, a failure changes nothing.
3. Data loss. The test database is empty, and the test compares schemas. A `Down` that drops a column with real data passes the test and still loses the data. The way back for data is a backup.

</details>
