# 0013 — Session lookup through one SECURITY DEFINER function

Date: 2026-11-12
Status: accepted. It adds one case to [ADR 0009](0009-account-isolation-rls.md)

## Context

ADR 0009 isolates accounts with RLS: each query of the app role runs inside `db.WithAccount`, and the rules show only the rows of that account. ADR 0012 selects server sessions in Postgres: a `sessions` table with the hash of the token and `account_id`.

A request with a session cookie must find its account first. At that moment the account is not known, so RLS shows no rows. ADR 0009 allows the system role (`BYPASSRLS`) only in a closed list of places, and the session lookup is not in it. Each dashboard request needs the lookup, so the case is frequent and must be narrow.

## Decision

- The SQL function `session_account(token_hash bytea) RETURNS uuid` is the only way around RLS for sessions. It returns the account of one live session (`expires_at > now()`), or NULL. It returns no other column and no other row.
- It is `SECURITY DEFINER`, and `app_system` owns it, so its body reads `sessions` without RLS. It sets `search_path = public, pg_temp`. Only `app_app` can execute it: `PUBLIC` has no right.
- To hand the function to the system role, the schema owner needs two things. The roles script gives `GRANT app_system TO app_owner WITH INHERIT FALSE, SET TRUE`: the owner gets no `BYPASSRLS` by default, only through an explicit `SET ROLE`. The migration gives the system role `CREATE` on the schema only for the hand-over, in the same transaction.
- All other session queries (create, end one, end all) run in `db.WithAccount` with the normal rules.
- `accounts` is a minimal table (`id`, `created_at`) with RLS on `id`, so `sessions.account_id` has a real foreign key. The product adds its columns later.

## Trade-off

- **Prefer:** a small attack surface, one connection pool, the normal RLS rules for all other session queries.
- **Give up:** some simplicity: a database function, a role grant in the roles script, and the rule that changes to this function need the same review as a system-role place.

## Stress round

| Stressor | Affects | Verdict | Resolution |
| --- | --- | --- | --- |
| An attacker calls the function with guessed hashes | Sessions | survives | The token has 128 random bits, and the function needs the exact SHA-256 hash. It returns only an account ID, never a token |
| A later migration changes the function to return more data | Isolation | accepted | The function is in `core/migrations/`: only the orchestrator changes it, and the review checks it like a system-role place |
| The owner role is stolen | The whole database | accepted | The owner already owns all tables. `SET ROLE app_system` adds `BYPASSRLS`, but the owner can already change the rules. The owner runs only migrations |
| An existing local database has no role grant | Local environment | accepted | The roles script runs only on an empty volume. `mise run reset` or one `GRANT` command fixes it ([deploy/README.md](../../../deploy/README.md)) |

## Consequences

- `testdb.CheckRLS` keeps all rules strict: `sessions` has the normal `account_isolation` rule.
- The dashboard middleware (`accounts.RequireSession`) calls the function once for each request, through the app pool.
- The system-role list of ADR 0009 now has one more case: this function. A new `SECURITY DEFINER` function needs a new ADR.
