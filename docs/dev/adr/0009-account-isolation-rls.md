# 0009 — Account isolation: a filter in code and Postgres rules from the start

Date: 2026-10-28
Status: accepted

## Context

The data of all accounts is in the same tables. [ADR 0005](0005-api-surfaces-and-ownership.md) already decided that the server finds the account, foreign keys are composite, and each module has an isolation test. One question stayed open: who adds the last filter "only the rows of this account"? The code in each query, or the database itself (Row-Level Security in Postgres)? Agents write most of the code, and a filter that is missing in one query leaks the personal data of another account.

We compared: a filter in code only; database rules only; both from the start; a filter now and rules before the pilot. The owner selected both from the start. The intermediate option requires us to remember the switch, and its gain is not worth that risk.

## Decision

Two independent layers from the first query.

**1. A filter in code.** Each query to account data contains `account_id = $1`. Indexes start with `account_id` ([backend standard 03](../backend-standard/03-data.md)).

**2. Postgres rules (RLS)** on each table with `account_id`:

```sql
ALTER TABLE items ENABLE ROW LEVEL SECURITY;
ALTER TABLE items FORCE ROW LEVEL SECURITY;
CREATE POLICY account_isolation ON items
  USING (account_id = nullif(current_setting('app.account_id', true), '')::uuid)
  WITH CHECK (account_id = nullif(current_setting('app.account_id', true), '')::uuid);
```

`nullif(..., '')` is necessary. After a transaction ends, the connection keeps the setting as an empty string, not NULL. Without `nullif`, the next query on the same pooled connection fails with "invalid input syntax for type uuid". We found this in the local environment.

**One entry point.** All access to account data goes through one function. The function opens a transaction, sets `set_config('app.account_id', <id>, true)` for this transaction only, and does the work. The connection pool is safe: the value lives until the transaction ends. If no account is set, the rule lets no rows through.

**Database roles:**

| Role | Used for | RLS |
| --- | --- | --- |
| Schema owner | Migrations | It owns the tables; `FORCE` applies to it too |
| Application | Normal work of the core and the workers | Applies; not the owner, no `BYPASSRLS` |
| System | Only marked places: account lookup by public address, matching an incoming webhook to an account, scheduled deletion, support access with the audit log | `BYPASSRLS`; use only in these places, and the review checks each use |

**River background jobs:** the job arguments carry `account_id`, and the handler uses the same entry point. The internal River tables have no rules.

## Consequences

- A missing filter in code does not cause a leak: the database returns an empty result. The isolation test checks both layers: a query without the filter, in the context of account A, does not see the rows of account B.
- "Why is the result empty?" is harder to debug, so the query logs show which account is set.
- Each new table with `account_id` gets its rule in the same migration. The review checks this with the checklist.
- The system role is the most dangerous place. Its use is limited to the listed cases, and the code shows it through a separate function.
- RLS is a topic in the optional learning map ([level 3](../../learning/backend-path.md#level-3--postgres-and-sql)).
