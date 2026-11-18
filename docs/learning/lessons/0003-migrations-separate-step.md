# 0003 — Who changes the database schema and when: migrations as a separate step

From PR #44 · 28.10.2026 · Topics: database migrations, deploy

## Problem

The schema of the database changes with the code: a new table, a new column. Something must apply these changes before the new code runs. The simple answer is "the application does it when it starts". We had to decide if that answer is safe.

## Predict first

Two copies of the API start at the same moment after a deploy. Each copy runs the migrations at startup. What can happen to a migration that adds a table? Which database rights does the API need for this design?

## Theory

A **migration** is a numbered file that changes the structure of the database. A migration tool keeps a table of the files that it applied, and applies only the new ones. Django does this with `manage.py migrate`. The difference here is that we write the files by hand in SQL.

There are two moments when migrations can run:

- **At application startup.** Each process applies the new migrations before it serves requests. No deploy step can be forgotten.
- **As a separate step.** One command runs one time, before the new version starts. Django works this way: `migrate` is a command in the deploy, not a part of `runserver`.

Startup migrations have three problems. Several copies start together and compete to apply the same file. A slow migration delays the start of each copy, and a platform that sees no answer restarts the process. The third problem is rights: the process that serves user requests must be able to change the schema. So an SQL injection or a bug can drop a table.

The separate step solves all three. It runs one time, it can take as long as it needs, and it uses its own database user. The price is an order that the deploy must keep: first the database, then the migrations, then the application.

Sources:

- [The Twelve-Factor App, "XII. Admin processes"](https://12factor.net/admin-processes): one-off tasks such as migrations run as separate processes of the same release.
- [Django documentation, "Migrations"](https://docs.djangoproject.com/en/5.2/topics/migrations/): the same model with a separate `migrate` command.
- [goose documentation](https://pressly.github.io/goose/): the migration tool that we use, with `Up` and `Down` sections in one SQL file.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| The application migrates at startup | The API applies new migrations before it listens | Copies compete. The application user needs the rights to change the schema |
| A person runs the SQL by hand | Somebody connects and applies the files | Steps are forgotten, and environments drift apart |
| A separate program, one step in the deploy | A `migrate` binary runs after the database starts and before the application starts | The deploy must keep the order. One more program to build |

## Edge cases

| What if | What happens |
| --- | --- |
| The step runs two times | The second run prints "no new migrations". The tool remembers the applied files |
| The image has no SQL files and no shell | `go:embed` puts the SQL files into the binary at build time, like a bundler that puts assets into the build |
| A migration waits for a lock that a long query holds | Each migration sets `lock_timeout` and `statement_timeout`, so it fails fast and does not block the application |
| The roles that the first migration names do not exist | The migration fails. The roles must exist before the first migration, so a script creates them when the database is new |
| A local database is older than a change in the roles script | A later migration failed on an old volume with a documented error. One documented `GRANT` or a reset fixes it |
| A `Down` section is broken | Production never runs `Down`. A test runs each migration down and up again and compares the schema, so we find the problem early |

## Decision

Migrations are a separate step. The program `migrate` is in the same image as the API and the worker. `mise run up` starts the database, runs `migrate`, and only then starts the core. The API cannot change the schema: it connects as the application role, and only `migrate` connects as the schema owner ([lesson 0005](0005-database-roles.md)).

This would be wrong for a program that the user installs, for example a desktop app with an embedded database. There is no deploy step there, so the program must migrate at startup.

There is no ADR for this decision. The role part is in [ADR 0009](../../dev/adr/0009-account-isolation-rls.md).

## In our code

1. `core/cmd/migrate/main.go`: the program. It reads `MIGRATE_DATABASE_URL`, the connection of the schema owner.
2. `core/migrations/migrations.go`: the `go:embed` line that puts the SQL files into the binary.
3. `core/migrations/00001_role_grants.sql`: the first migration, with its `Up` and `Down` sections and the time limits.
4. `deploy/compose.yaml`, the `migrate` service: a one-off container in the `tools` profile.

## Check yourself

1. Why does the API not run the migrations when it starts?
2. The `migrate` step fails in the middle of a deploy. What does the old version of the application do?
3. Why do we test `Down` sections that production never runs?

<details><summary>Answers</summary>

1. Two copies can start together and try to apply the same migration. Also, the API would need the rights of the schema owner, so a bug in a request handler could change or drop tables. A separate step runs one time with its own database user.
2. It continues to run. The step is before the start of the new version, so a failed migration stops the deploy and not the service. This is true only when the migration did not break the old code, which is why a schema change must stay compatible with the version that runs.
3. A rollback is rare and happens in a bad moment. A `Down` section that nobody ran is probably broken at that moment. The test makes the check cheap and constant.

</details>
