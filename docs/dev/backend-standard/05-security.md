# 05 — Security

**Status:** DRAFT
**Decisions:** B6, B10

## Account isolation

- The server finds the account: from the session, the address or a signed parameter. Do not accept `account_id` from the request ([ADR 0005](../adr/0005-api-surfaces-and-ownership.md)).
- Four layers of protection: the server finds the account; an `account_id` filter in each query; Postgres rules (RLS) on each table with `account_id` ([ADR 0009](../adr/0009-account-isolation-rls.md)); an isolation test in each module ([07](07-testing.md)). Composite keys stop records from pointing to other accounts.
- Access to account data goes only through one entry point, which sets the account for the transaction. Use the system role with `BYPASSRLS` only in the places that ADR 0009 lists.

## Login and sessions

- Login with a one-time code: a short lifetime, a limit on attempts per code and per address, a delay after errors.
- The session is a cookie with `HttpOnly`, `Secure`, `SameSite=Lax`. Requests that change data also check the `Origin` header.
- Rate limits on the public API, login and webhooks.

## Data

- Do not write the personal data of users to logs, traces, error texts or external sinks.
- Each new place that stores personal data needs an entry in `docs/dev/data-map.md` before the merge.
- The audit log records our support access to a user's account and the creation, change and deletion of personal data: who, when, what.

## External calls

- Check the signature of each incoming webhook before you save it.
- Platform secrets are in the CI secret store. Account secrets are encrypted in the database.
- Agents do not get production secrets during development ([AI standard](../ai-standard/02-harness.md#secrets)).

## Checks

- CI runs the dependency audit on each change: `task audit:go` (govulncheck on the core, only the vulnerabilities that our code calls) and `task audit:deps` (osv-scanner on the web lockfile). An accepted finding goes into `osv-scanner.toml` with a reason and an end date. Dependabot security updates open fix PRs.
- Before the pilot and before we connect payments: an audit with the skill [cloudflare/security-audit-skill](https://github.com/cloudflare/security-audit-skill) in an isolated sandbox ([AI standard](../ai-standard/03-review-verification.md)).
- For early-stage priorities, use [security-101-for-saas-startups](https://github.com/forter/security-101-for-saas-startups) as a guide. Check its last update date before you rely on specific advice.
