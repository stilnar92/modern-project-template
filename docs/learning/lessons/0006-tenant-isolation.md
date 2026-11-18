# 0006 — One database, many accounts: how account A never sees the data of account B

From PRs #30, #46 · 28.10.2026 · Topics: multi-tenancy, row-level security

## Problem

The data of all users is in the same tables. Each row has an `account_id`. A query that forgets `WHERE account_id = ...` returns the clients of another user, and that is a leak of personal data. Agents write most of the queries. We had to decide who adds the last filter: the code in each query, or the database.

## Predict first

A table has a correct rule: "show only the rows of the current account". Somebody adds a second rule for reports: `FOR SELECT USING (true)`. Which rows does account A see now: its own rows, no rows, or all rows?

## Theory

A system where many customers share one database is **multi-tenant**. Each customer is a tenant. The cheapest form is shared tables with a tenant column. Its risk is that the isolation depends on each query.

**Row-level security** (RLS) moves the filter into the database. You enable it on a table and add a **policy**: a condition that Postgres adds to each query on that table. A policy has two parts. `USING` filters the rows that a query can see, change or delete. `WITH CHECK` tests the rows that a query writes. If `WITH CHECK` is missing, Postgres uses the `USING` condition for writes too.

The database must know the current tenant. A common way is a setting on the transaction: the code calls `set_config('app.account_id', id, true)`, and the policy reads it with `current_setting`. The `true` makes the value local to the transaction. This is important with a connection pool, because the next request gets the same connection.

Three rules of Postgres change the design:

- The owner of a table ignores its policies, unless the table has `FORCE ROW LEVEL SECURITY`.
- Superusers and roles with `BYPASSRLS` always ignore policies.
- Normal (permissive) policies of one table combine with OR. One wide policy opens what the others close.

The Django analogy is a custom model manager that always filters by tenant. The difference is that raw SQL can go around a manager, and nothing goes around RLS.

Sources:

- [PostgreSQL documentation, "Row Security Policies"](https://www.postgresql.org/docs/current/ddl-rowsecurity.html): `ENABLE`, `FORCE`, owners and `BYPASSRLS`.
- [PostgreSQL documentation, CREATE POLICY](https://www.postgresql.org/docs/current/sql-createpolicy.html): `USING`, `WITH CHECK`, and how permissive policies combine.
- [Multi-tenant data isolation with PostgreSQL Row Level Security](https://aws.amazon.com/blogs/database/multi-tenant-data-isolation-with-postgresql-row-level-security/), AWS Database Blog: the pattern with a session setting.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| A filter in code only | Each query has `account_id = $1` | One missing filter is a leak |
| Database rules only | RLS on each table, no filter in the queries | One layer. The queries do not show their scope, and indexes on `account_id` are easy to forget |
| A filter now, rules before the pilot | Start simple and add RLS later | Somebody must remember the switch, and all old tables need the rules at one time |
| Both from the first query | Two independent layers | Each table needs three more SQL lines. An empty result is harder to debug |

## Edge cases

| What if | What happens |
| --- | --- |
| No account is set | The policy matches no rows. The result is empty, not an error |
| The transaction ended and the connection returns to the pool | The setting stays as an empty string, not NULL. A cast of `''` to `uuid` fails, so the next query on that connection failed. The rule uses `nullif(..., '')`. Only the real database showed this |
| The schema owner reads the table | With `FORCE`, the owner sees 0 rows without an account. It also cannot insert a row, so test setup that must ignore the rules uses the system role |
| A second, wide policy is added | The policies combine with OR, so each account sees all rows. The check reads the text of each policy and fails |
| Account B inserts a row with the ID of account A | `WITH CHECK` refuses it. Without that check, B could write rows that A then sees |
| A view reads the table | A view runs with the rights of its owner by default, so RLS does not apply to its user. The check requires `security_invoker` on each view |
| A request must find its account from a cookie | No account is known, so RLS shows nothing. One narrow function does this lookup ([lesson 0005](0005-database-roles.md)) |

## Decision

We use both layers from the first query. Each query has the filter, and each table with `account_id` has `ENABLE`, `FORCE` and one policy with `USING` and `WITH CHECK`, in the same migration as the table. All access goes through one function that opens a transaction and sets the account for that transaction only. A test reads the Postgres system tables and fails the merge when a table has no rule or an open rule ([lesson 0007](0007-tests-with-real-database.md)).

The owner selected this option over "a filter now, rules before the pilot". The intermediate option needs somebody to remember the switch, and its gain is not worth that risk.

This would be wrong where each tenant has its own database or schema. There the isolation is already physical.

Record: [ADR 0009](../../dev/adr/0009-account-isolation-rls.md).

## In our code

1. `core/internal/platform/db/db.go`, `WithAccount`: the one entry point. It sets `app.account_id` for one transaction.
2. `core/migrations/00002_accounts_sessions.sql`: two real tables with `ENABLE`, `FORCE` and the policy.
3. `core/internal/platform/testdb/rls.go`, `CheckRLS`: the query that reads each table, each policy and each view.
4. `core/internal/platform/testdb/isolation.go`, `CheckIsolation`: account B tries to read, change, delete and insert the rows of account A.

## Check yourself

1. Why do we keep the filter in code when the database rule already hides the rows?
2. What is the answer to the "Predict first" question, and why?
3. Why does the entry point set the account for the transaction and not for the connection?

<details><summary>Answers</summary>

1. The layers fail in different ways. A table can lose its rule in a bad migration, or a place can use the system role. Then the filter in code still works. The filter also lets the query use an index that starts with `account_id`.
2. All rows of all accounts. Permissive policies combine with OR: a row is visible if any policy accepts it, and `USING (true)` accepts each row. The correct policy next to it changes nothing.
3. A pool gives the same connection to the next request. A value on the connection would stay there, and the next request, maybe of a different account, would see the data of the first. A value on the transaction ends with the transaction.

</details>
