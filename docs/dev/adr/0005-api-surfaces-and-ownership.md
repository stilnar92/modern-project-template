# 0005 — API sections, data ownership and compatibility

Date: 2026-10-28
Status: accepted

## Context

The core has several clients: the dashboard, public pages, and external services that send webhooks. Many accounts share one platform. We need one rule that tells the core who owns the data and who calls it. The rule must stop one account from seeing or changing the data of another account.

The template has no native mobile app: users open the dashboard in the browser or as a PWA. So we do not need token login.

## Decision

**Three API sections.** Each section has its own access check:

| Section | Who calls it | Where the account comes from |
| --- | --- | --- |
| `/v1/dashboard/...` | The dashboard (PWA) | The session in an httpOnly cookie |
| `/v1/public/...` | Public pages and other anonymous clients | The address: domain or slug, or a signed parameter |
| `/v1/webhooks/...` | External services: payments, messengers, other providers | The signature and the source parameters |

**The server finds the account, not the request.** The API never accepts `account_id` from the browser: only from the session, the address or a signed parameter.

**Ownership in the database:**

- each account-owned table has a required `account_id`;
- foreign keys are composite, `(account_id, id)`: a record cannot point to a record of another account, even if the code has a bug;
- each module must have the test "other accounts' data is not visible and cannot be changed".

**Public addresses** (a domain or a slug that resolves to an account) are unique on the whole platform. Reserved names are forbidden. After a rename, the old address becomes a 301 redirect.

**Compatibility:** inside `/v1`, changes only add things. To remove or rename a field, use a deprecation period and a new version.

**Our support** is a separate role outside accounts. Each time support enters a customer's account to help, the audit log records who, when and why.

## Consequences

- A new frontend connects to one of the three sections and needs no new access logic.
- Isolation has three layers: the server finds the account, the database does not let records point to other accounts, and tests check each module. The choice between RLS and a filter in code stays open and does not affect this decision.
- Composite keys make the schema and migrations a little more complex.
- If we need a native app, a new ADR adds token login to `/v1/dashboard`.
