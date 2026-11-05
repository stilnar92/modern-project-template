# 00 — Backend standard overview

**Status:** DRAFT

## Fixed decisions

The sections are written around these limits. Only an ADR can change them.

| ID | Decision | Section |
| --- | --- | --- |
| B1 | A modular monolith in Go. A new service needs an ADR and a specific reason: load, a separate release cycle, a different team | 01 |
| B2 | Contract first: `api/openapi.yaml` → oapi-codegen → handlers implement the generated interfaces. CI checks that the generated code matches the contract | 02 |
| B3 | API errors are Problem Details by RFC 9457 (`application/problem+json`) | 02 |
| B4 | Inside `/v1`, changes only add things. To remove or rename, use deprecation ([ADR 0005](../adr/0005-api-surfaces-and-ownership.md)) | 02 |
| B5 | Postgres, pgx + sqlc, no ORM. Migrations are only backward compatible: expand → migrate → contract | 03 |
| B6 | The server finds the account; an `account_id` filter in code and Postgres rules (RLS) on each table; composite foreign keys `(account_id, id)`; an isolation test in each module ([ADR 0005](../adr/0005-api-surfaces-and-ownership.md), [ADR 0009](../adr/0009-account-isolation-rls.md)) | 03, 05 |
| B7 | Each side effect outside the database is a River job, added in the same transaction. Job handlers are idempotent | 04 |
| B8 | Events go to the outbox in the same transaction as the action ([ADR 0003](../adr/0003-analytics-outbox-and-sinks.md)) | 04 |
| B9 | External sources connect only through adapters in four directions ([architecture](../architecture.md#rules-for-the-code-structure)) | 04 |
| B10 | The personal data of users does not go into logs, traces or external sinks | 05, 06 |
| B11 | Observability from day one: structured logs, traces through HTTP and jobs, errors in Sentry, queue metrics | 06 |
| B12 | Tests use a real Postgres (testcontainers), no database mocks; external sources use recorded responses | 07 |

## Sections

| # | Section | File | Status |
| --- | --- | --- | --- |
| 01 | Modules and coupling | [01-modules.md](01-modules.md) | DRAFT |
| 02 | API | [02-api.md](02-api.md) | DRAFT |
| 03 | Data and migrations | [03-data.md](03-data.md) | DRAFT |
| 04 | Background jobs, events and integrations | [04-async-integrations.md](04-async-integrations.md) | DRAFT |
| 05 | Security | [05-security.md](05-security.md) | DRAFT |
| 06 | Observability | [06-observability.md](06-observability.md) | DRAFT |
| 07 | Tests | [07-testing.md](07-testing.md) | DRAFT |

## Priority

1. ADRs in [`docs/dev/adr/`](../adr/)
2. Fixed decisions B1–B12
3. Sections 01–07
4. Code, comments, habits

A conflict that you find is a defect: fix the document that loses.

## What to read

Selected from the owner's bookmarks on 28.10.2026 and checked for currency. These books and courses are for practice, not for interviews.

| Source | Why | Section |
| --- | --- | --- |
| Vlad Khononov, *Balancing Coupling in Software Design* (2024) | Where to draw module boundaries: strength, distance, volatility | 01 |
| John Ousterhout, *A Philosophy of Software Design* (2nd ed., 2021) | Deep modules: a simple interface that hides a lot of complexity | 01 |
| [Clean Code: First Edition Critique](https://bugzmanov.github.io/cleancode-critique/index.html) | Why not to copy the advice of Clean Code (1st ed.) into code as-is | 01 |
| [Software Engineering at Google](https://abseil.io/resources/swe-book) | Review, tests, API deprecation, large changes | 02, 07 |
| [RFC 9457 — Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457) | Error format; replaces RFC 7807 from an old bookmark | 02 |
| [SQL Performance Explained](https://sql-performance-explained.com/) (use-the-index-luke) | Indexes and query plans | 03 |
| [Course "Asynchronous architecture"](https://tough-dev.school/architecture) | Idempotency, retries, outbox, dead-letter queue, event versions; the examples use Kafka, the ideas apply to River | 04 |
| [Business Oriented System Design](https://vvsevolodovich.dev/business-oriented-system-design-course/) (Vladimir Ivanov) | Requirements, estimates, C4, ADRs and RFCs, defending decisions to the business | 00 |
| [Caching best practices](https://jakearchibald.com/2016/caching-best-practices/) (Jake Archibald) | `Cache-Control` and hashed file names; still correct, although from 2016 | 02 |
| [security-101-for-saas-startups](https://github.com/forter/security-101-for-saas-startups) | Security by startup stage; we could not check the last update date | 05 |
