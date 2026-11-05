# Non-functional requirements

**Status:** template. Fill in the targets for your project. A number here is a target, not a measured result. Mark a target "not verified" until a load test or production data checks it. When a target changes, update its row. Do not add a second row.

## Scope and scale

<!-- The expected load at launch. Give a reason for each number. -->

| Dimension | Target | Why this number |
| --- | --- | --- |
| Accounts | | |
| Public page requests | | |
| Concurrent dashboard sessions | | |
| Background jobs per day | | |

## Page response time

<!-- Target server time for the public site (web/apps/site), for the common case. -->

## API response time

- `/v1/dashboard` and `/v1/public` reads: <!-- target, e.g. under 200 ms at the server -->.
- `/v1/webhooks/*` must respond within a few seconds. Otherwise the source sends the event again or drops it. So a webhook handler saves the raw request, responds at once and processes the request in a River job ([architecture](architecture.md#rules-for-the-code-structure)).
- The response-time metric is one of the required metrics ([backend standard 06](backend-standard/06-observability.md)).

## Availability

<!-- The uptime target, and which part must stay up first: the public site or the dashboard. -->

The alerts that protect availability are in [backend standard 06](backend-standard/06-observability.md#alerts).

## Background job limits

- Each job that touches account data carries `account_id` in its arguments. The handler uses the same entry point as the API, so a retry stays inside one account ([ADR 0009](adr/0009-account-isolation-rls.md)).
- A job that uses all its retries goes to the dead-letter queue. It does not disappear. This is an alert condition ([backend standard 06](backend-standard/06-observability.md)).
- External calls never run inside a database transaction ([backend standard 03](backend-standard/03-data.md#transactions)). A slow external call cannot hold a lock on account data.
- Each adapter sets its own rate limit from the documented limits of its provider.

## Multi-tenancy and isolation

- Each table with account data gets an RLS policy in the same migration that creates it ([ADR 0009](adr/0009-account-isolation-rls.md)). This is a hard rule, not a target. A migration without the policy fails the review.
- Each module has an isolation test: a query in the context of account A does not see the rows of account B, even without the filter in code ([ADR 0005](adr/0005-api-surfaces-and-ownership.md)).
- The system database role (`BYPASSRLS`) is limited to the cases in [ADR 0009](adr/0009-account-isolation-rls.md#decision). A new use of this role needs its own review.

## Data residency

<!-- Where the data can be stored. The personal data categories are in data-map.md. -->

See [data-map.md](data-map.md).

## Compatibility

- Inside `/v1`, changes only add fields or endpoints. To remove or rename a field, use a deprecation period and a new version ([ADR 0005](adr/0005-api-surfaces-and-ownership.md)).
- Migrations are backward compatible in three steps: expand, migrate, contract ([backend standard 03](backend-standard/03-data.md#migrations)).

## What still needs a real decision

<!-- Targets that need a measurement or a decision. Remove an item when you record the answer above. -->

- [ ]
