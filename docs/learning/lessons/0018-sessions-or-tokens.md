# 0018 — Who sends this request: server-side sessions or self-contained tokens

From PR #106 · 12.11.2026 · Topics: authentication, sessions

## Problem

A user logs in one time and then sends hundreds of requests. HTTP has no memory: each request arrives alone. The server must know the account of each request. It must also be able to end a login, for example when the user logs out or loses a phone.

## Predict first

A user logs out, or an administrator blocks the account. The browser of an attacker still has a copy of the old credential. With a server-side session, what does the server do? With a signed token that is valid for one hour, what can the server do?

## Theory

There are two models.

**Server-side session.** At login the server makes a long random value, the session token, and saves a row: token, account, end time. The browser gets only the token, in a cookie. For each request the server finds the row. The token has no meaning without the row. This is the model of Django sessions with the database backend.

**Self-contained token.** At login the server makes a document with claims, for example "account 42, valid until 13:00", and signs it. The usual format is JWT. For each request the server checks the signature and the end time, and reads the account from the token. It needs no database query, and each service that has the key can check the token.

The difference is **revocation**. A session ends when the server deletes the row. A signed token stays valid until its end time, because nothing asks the server. The usual fixes bring the state back: a short lifetime with a second, stored "refresh" token, or a list of revoked tokens that each request must read.

How the browser keeps the credential is a separate question. A cookie with `HttpOnly` is not readable from JavaScript, so a script injection cannot copy it. But the browser sends a cookie automatically, which opens CSRF ([lesson 0020](0020-requests-from-another-site.md)).

Sources:

- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html): session IDs, cookie flags, expiry, a new ID at login.
- [RFC 7519: JSON Web Token](https://www.rfc-editor.org/rfc/rfc7519): the token format and its claims.
- [RFC 8725: JSON Web Token Best Current Practices](https://www.rfc-editor.org/rfc/rfc8725): the known mistakes with JWT.
- [Django: How to use sessions](https://docs.djangoproject.com/en/5.2/topics/http/sessions/): the same model in a framework that the reader knows.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Self-contained token (JWT) | The server checks a signature. No lookup | No revocation before the end time. Key management |
| JWT with a refresh token or a revoke list | A short token, and stored state to end it | The state and the lookup come back, with more moving parts |
| A session library (alexedwards/scs, gorilla/sessions) | Ready-made store and middleware | scs has a fixed table with no account column, and it stores the plain token |
| Our own sessions table | A row for each session, with the account and the end time | We write and test the code. One database lookup for each request |

## Edge cases

| What if | What happens |
| --- | --- |
| The user logs out | The server deletes the row and clears the cookie. A copy of the token stops working immediately |
| The login address of an account changes | One `DELETE` by `account_id` ends each session of the account |
| The session is past its end time | The lookup checks `expires_at`, so the session does not work. A job deletes old rows each hour |
| The cookie is missing, unknown or forged | The middleware answers 401 in the Problem format |
| The database fails during the lookup | The answer is 500, not 401. A database failure must not look like "log in again" |
| The lookup runs before the server knows the account | The row rules (RLS) show no rows then. One narrow database function does this lookup ([ADR 0013](../../dev/adr/0013-session-lookup-function.md)) |
| A user logs in again | The login flow starts a new session and ends the old one, so a login always gets a new token |

## Decision

We use server-side sessions in Postgres, with our own code. The cookie has `HttpOnly`, `Secure` and `SameSite=Lax`. The table keeps `account_id`, so SQL can end all sessions of an account, and the code fits the account isolation rules. The record compares session libraries. It does not record a comparison with JWT.

The cost is one database lookup for each dashboard request. This would be wrong if many separate services had to check a login without access to this database. Then a signed token with a short lifetime is the usual choice.

The module does not have a login yet. How a user proves who they are is a product decision.

Record: [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md), [ADR 0013](../../dev/adr/0013-session-lookup-function.md).

## In our code

1. `core/migrations/00002_accounts_sessions.sql`: the `sessions` table and the lookup function.
2. `core/internal/accounts/accounts.go`: `StartSession`, `SessionAccount`, `EndSession`, `EndAllSessions`.
3. `core/internal/accounts/http.go`: `SetCookie` and the middleware `RequireSession`.
4. `core/internal/accounts/accounts_test.go`, `TestSessionLifecycle` and `TestRequireSession`: lookup, expiry, logout, end all.

## Check yourself

1. Why is "log out on all devices" one SQL statement in our design? What would it need with a JWT that is valid for one day?
2. The database is down for a moment during the session lookup. Why must the answer not be 401?
3. What does the server-side session cost for each request, compared to a signed token?

<details><summary>Answers</summary>

1. Each session is a row with `account_id`. A `DELETE` by that column removes all of them, and the next request with an old token finds no row. A JWT stays valid until its end time, so the server would need a revoke list that each request reads, or a change of the signing key for all users.
2. A 401 tells the client "log in again", and the client can drop a session that is still good. The real problem is on the server. A 500 says so, and the error goes to the log and to Sentry.
3. One database query for each request. A signed token needs only a signature check in memory. We accept the query because it gives immediate revocation.

</details>
