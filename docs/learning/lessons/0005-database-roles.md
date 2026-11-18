# 0005 — The application must not be able to do everything in the database: roles and least privilege

From PRs #44, #106 · 12.11.2026 · Topics: database security, least privilege

## Problem

A typical web project has one database user, and the application does everything with it: it creates tables, reads data and deletes data. Our database protects each account with rules inside Postgres ([lesson 0006](0006-tenant-isolation.md)). Those rules depend on who is connected. So we had to decide which database users exist and what each can do.

## Predict first

A migration creates a function and then runs `REVOKE ALL ON FUNCTION ... FROM PUBLIC`. Before the `REVOKE`, the migration gives the function to a different owner. Postgres prints a warning, not an error, and the migration succeeds. Who can call the function now?

## Theory

**Least privilege** means that each part of a system has only the rights that its job needs. Then a bug or an attack in that part can do only what that part can do.

In Postgres, a **role** is a database user with rights. Four facts are important:

- A **superuser** passes all checks, including row-level security.
- The **owner** of an object can change it, drop it and change its rights. Only the owner, or a superuser, can `GRANT` and `REVOKE` on it.
- `PUBLIC` means "each role". Some defaults are open: `PUBLIC` can execute each new function.
- **Default privileges** say "each table that this role creates later gets these rights for that role". Without them, you give rights after each new table.

A `SECURITY DEFINER` function runs with the rights of its owner, not of its caller. It is `sudo` for one command. It is a narrow door through a wall, and it has two risks. The first is the open default: you must revoke `PUBLIC` yourself. The second is the `search_path`: a caller can put its own objects before the real ones, so the function must set a safe `search_path`.

Sources:

- [PostgreSQL documentation, "Privileges"](https://www.postgresql.org/docs/current/ddl-priv.html): owners, `GRANT`, `REVOKE`, `PUBLIC` and the default rights for each object type.
- [PostgreSQL documentation, CREATE FUNCTION](https://www.postgresql.org/docs/current/sql-createfunction.html), section "Writing SECURITY DEFINER Functions Safely".
- [PostgreSQL documentation, ALTER DEFAULT PRIVILEGES](https://www.postgresql.org/docs/current/sql-alterdefaultprivileges.html).
- [Saltzer and Schroeder, "The Protection of Information in Computer Systems"](https://www.cs.virginia.edu/~evans/cs551/saltzer/), 1975: the source of the least privilege principle.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| The application connects as the superuser | One user, no setup | A superuser ignores row-level security. Each bug has all rights |
| One owner role for migrations and requests | The usual Django setup | A bug in a request handler can change the schema and the rules |
| Two roles: owner and application | The owner migrates. The application only reads and writes rows | Some jobs must cross accounts, for example scheduled deletion. They have no role |
| Three roles: owner, application, system | The system role ignores the rules and is used only in listed places | Three passwords, three connection strings, and each use of the system role needs a review |

## Edge cases

| What if | What happens |
| --- | --- |
| The application role tries to create a table | "permission denied for schema public". The script revokes all rights on the schema from `PUBLIC`, and only the owner creates objects |
| A migration adds a table | The application role can use it immediately. The first migration sets default privileges for the tables that the owner creates |
| A request must find its account before the account is known | A system connection for each request is too wide. One `SECURITY DEFINER` function returns one account ID and nothing more |
| `REVOKE` runs after the function has a new owner | The migration role is no longer the owner. Postgres only warns, and `PUBLIC` keeps the right to execute. We found this in the local database. The rights now come before the hand-over |
| The tests call the function and pass | They passed with the bug too: a call proves that the application can use the function, not that others cannot. A test now reads the list of rights |
| The owner must give a function to the system role | The owner is a member of the system role with `INHERIT FALSE, SET TRUE`. It gets the power only through an explicit `SET ROLE`, not by default |

## Decision

We use three roles. The schema owner runs only the migrations. The application role runs the API and the worker: it does not own the tables and cannot ignore the rules. The system role has `BYPASSRLS` and is used only in the places that the ADR lists. The `postgres` superuser is for administration only, and no code uses it.

The session lookup is the one narrow exception: a `SECURITY DEFINER` function that the system role owns and only the application role can execute. A new function of this type needs a new ADR.

This would be too much for a database with one tenant and no row rules. There, two roles (owner and application) give most of the gain.

Record: [ADR 0009](../../dev/adr/0009-account-isolation-rls.md) and [ADR 0013](../../dev/adr/0013-session-lookup-function.md).

## In our code

1. `deploy/postgres/init/01-roles.sh`: the three roles and the `REVOKE` on the schema and the database.
2. `core/migrations/00001_role_grants.sql`: the default privileges for the application and system roles.
3. `core/migrations/00002_accounts_sessions.sql`, the function `session_account`: read the comments about the order of the rights.
4. `core/internal/accounts/accounts_test.go`, `TestLookupFunctionRights`: the test that reads the rights.

## Check yourself

1. Why does the API not connect as `postgres`, when that is the simplest setup?
2. In the "Predict first" case, who can call the function, and why did no test fail?
3. Why is one `SECURITY DEFINER` function safer than a system-role connection for the session lookup?

<details><summary>Answers</summary>

1. `postgres` is a superuser, and a superuser ignores row-level security. A query with a missing filter would then return the rows of all accounts. The API could also drop tables.
2. Each role can call it. `PUBLIC` has the execute right by default, and the `REVOKE` did nothing because the migration role was no longer the owner. The tests only called the function as the application role, and that call works in both cases. Only a test that reads the rights sees the difference.
3. The function can do one thing: return the account ID of one live session for an exact token hash. A system connection can read each row of each table. If the calling code has a bug, the function limits the damage to that one answer.

</details>
