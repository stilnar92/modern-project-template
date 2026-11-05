# Topic map: backend in Go

We learn through tasks ([how](README.md#how-one-task-teaches)). This map is a guide, not a schedule. It shows which topics depend on which, where to read, and which exercises fit when a topic appears in a task. The levels go from simple to complex and follow the usual order of core development. A level is mastered when its main concepts are at "can do alone" in the [concept map](concepts.md).

**The map is alive.** When a new technology comes into the project, the same change adds a topic or a level with an exercise and sources. Then the technology leaves the "Waiting their turn" list at the end. The rule is in the [AI standard](../dev/ai-standard/05-human-in-the-loop.md#learning-during-development).

The main book is [Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests): it teaches Go through tests, as we work (TDD). References: [A Tour of Go](https://go.dev/tour/), [Go by Example](https://gobyexample.com/), [Effective Go](https://go.dev/doc/effective_go).

## Level 0 — Go with simple functions

**Topics:** `go mod`, `go run`, `go test`. Types, variables, functions, `if` and `for`. Strings, slices and maps. Errors as values (`value, err`).
**Analogy:** Go returns errors and does not throw them, as the errors-as-values approach in [frontend standard 27](../dev/frontend-standard/27-error-handling.md).
**Exercises:**
- Format an amount in minor units as a string with a currency.
- Validate a slug: Latin letters, digits, hyphen, length, and reserved names (`admin`, `api`).
- A table test for each function.
**Sources:** Learn Go with Tests: Hello World, Integers, Iteration, Arrays and slices, Maps. A Tour of Go: Basics.
**Mastered when:** the learner writes a function and its table test alone.

## Level 1 — structs, methods, interfaces

**Topics:** `struct`, methods, pointers, interfaces (implicit), packages and visibility (capital letter = public), `go vet`, the linter.
**Analogy:** a Go interface is like a TypeScript type that an object matches by shape, with no `implements`.
**Exercises:**
- A domain rule as a pure function with table tests, from your first slice.
- A storage interface and its in-memory implementation for tests.
**Sources:** Learn Go with Tests: Structs, methods & interfaces. Pointers & errors. Dependency Injection.
**Mastered when:** the learner explains why the exercise needs the interface and writes the rule with tests alone.

## Level 2 — HTTP

**Topics:** `net/http`, handlers, routing, JSON in and out, status codes, middleware, `context.Context`, errors as RFC 9457 Problem Details ([backend standard 02](../dev/backend-standard/02-api.md#errors)).
**Difference:** Go has no built-in framework. Routing and middleware come from the standard library.
**Exercises:**
- `GET /v1/public/<resource>/{slug}` with in-memory storage: 200 with JSON, or 404 as Problem Details.
- A middleware that adds `request_id` to the response and to the context.
**Sources:** Learn Go with Tests: HTTP Server, Context. [net/http](https://pkg.go.dev/net/http).
**Mastered when:** the learner writes a handler with an `httptest` test alone and knows where its `context` comes from.

## Level 3 — Postgres and SQL

**Topics:** tables, primary and foreign keys, composite keys, indexes, transactions. goose migrations, sqlc queries, testcontainers tests. Row-level security and database roles ([ADR 0009](../dev/adr/0009-account-isolation-rls.md)).
**Analogy:** sqlc generates typed Go code from SQL, as openapi-typescript generates a client from the contract. There is no ORM: we write SQL and migrations by hand.
**Exercises:**
- A migration for a table with `account_id` and a composite key.
- An sqlc query and an isolation test: account A does not see the rows of account B.
- Read the `EXPLAIN` of a query and say if it uses the index.
**Sources:** [Use The Index, Luke](https://use-the-index-luke.com/), [sqlc](https://docs.sqlc.dev/), [goose](https://pressly.github.io/goose/), [testcontainers-go](https://golang.testcontainers.org/).
**Mastered when:** the learner writes a migration, a query and an isolation test alone, and explains why `account_id` comes first in the index.

## Level 4 — API contract

**Topics:** OpenAPI, oapi-codegen, implementing the generated interface, validation, compatibility (only add things).
**Analogy:** the same "contract first" as on the frontend ([frontend standard](../dev/frontend-standard/00-overview.md)), from the server side.
**Exercises:** add an endpoint and an optional field to the contract, generate the code, implement it. Explain why you cannot remove a field at once.
**Sources:** [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen), [backend standard 02](../dev/backend-standard/02-api.md).
**Mastered when:** the learner does "contract → code → test" alone.

## Level 5 — background jobs

**Topics:** the River queue, enqueue in the same transaction, idempotency, retries, timeouts and cancellation through `context`.
**Analogy:** a River job is like a Celery or BullMQ job, but the queue is in the same Postgres and the job goes in with the data in one transaction.
**Exercises:** a "send a notification" job. It goes in with its data in one transaction and does not send twice when it runs again. Test: "run it twice".
**Sources:** [River](https://riverqueue.com/docs), [backend standard 04](../dev/backend-standard/04-async-integrations.md).
**Mastered when:** the learner explains what happens when the server stops in the middle of a job, and writes an idempotent handler alone.

## Level 6 — concurrency

**Topics:** goroutines, channels, `sync`, `errgroup`, limited parallelism, the `-race` detector.
**Analogy:** `errgroup` is like `Promise.all` that cancels the rest at the first error. **Difference from Node:** Go code runs in parallel for real, so shared data needs protection.
**Exercises:** process a list of items in parallel, with at most 4 at a time, and cancel on error. Test with `-race`.
**Sources:** Learn Go with Tests: Concurrency, Select, Sync, Context. [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup).
**Mastered when:** the learner writes limited parallel processing and explains where a race could be.

## Level 7 — architecture

**Topics:** modules and balanced coupling, deep modules, outbox and events, adapters for external sources, ADRs.
**Exercises:**
- An adapter for an incoming webhook: from the signature check to a core command.
- Write a draft ADR for a real decision. The agent reviews it.
**Sources:** [backend standard 01 and 04](../dev/backend-standard/00-overview.md#what-to-read), books by Khononov and Ousterhout.
**Mastered when:** the learner prepares a decision card ([AI standard 05](../dev/ai-standard/05-human-in-the-loop.md#decision-card)) with options and consequences alone.

## Level 8 — operations

**Topics:** structured logs, traces, metrics and alerts, Sentry, Docker and healthchecks, deploy and rollback, backups.
**Already in the template:** Docker Compose and the core image, Sentry through `SENTRY_DSN`, healthchecks, the build version, `task db:dump` and `task db:restore`.
**Exercises:** find with logs and a trace why a notification did not arrive. Add a metric and an alert.
**Sources:** [backend standard 06](../dev/backend-standard/06-observability.md), [operations](../dev/ops.md).
**Mastered when:** the learner finds the cause of a failure from logs and a trace with no hint.

## Waiting their turn

Technologies from the [tech stack](../dev/tech-stack.md) that are not in the code yet. When one comes into the project, move it to a level above.

| Technology | When it comes | Where in the map | What to learn |
| --- | --- | --- | --- |
| sqlc | The first module with queries | Level 3 | Queries in SQL files, generated Go code |
| River | The first background job | Level 5 | Jobs in a transaction, retries, idempotency |
| <technology> | <the slice that needs it> | <level> | <topics> |
