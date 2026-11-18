# accounts

Accounts and server sessions. A core module (ADR 0012, ADR 0013).

Tables: accounts, sessions

## Boundaries

- Now: an empty account (`id`, `created_at`), sessions in Postgres with an `HttpOnly` cookie, and the middleware that puts the account of the session into the request context.
- Not here yet: the login method (a code by email, SMS or Telegram), owners, staff and roles. These are product decisions. The login flow calls `StartSession` after it checks the user, and `EndSession` for the old cookie, so a login always gets a new token.

## Interface

- `Service.CreateAccount`, `Service.StartSession`, `Service.SessionAccount`, `Service.EndSession`, `Service.EndAllSessions`.
- `Service.RequireSession`: the middleware for `/v1/dashboard/...`. It answers 401 with a Problem body when the cookie has no live session. Handlers read the account with `accounts.AccountID(ctx)` and pass it to `db.WithAccount`.
- `Service.SetCookie`, `Service.ClearCookie`.
- `NewCleanupWorker(systemPool, logger)` and `CleanupJob()`: the worker process registers them.

## Security

- The cookie holds a random token (128 bits). The table keeps only its SHA-256 hash, so a database leak gives no working session.
- The lookup by token runs before the account is known. The SQL function `session_account` is the only way around RLS for it ([ADR 0013](../../../docs/dev/adr/0013-session-lookup-function.md)).
- Expired sessions do not work: the function checks `expires_at`. `CleanupWorker` deletes them each hour with the system role (ADR 0009: scheduled deletion). It is the only system-role place of the module.

## Events

None yet.

## Dependencies

`platform/db`, `platform/problem`, `platform/observability`, River.
