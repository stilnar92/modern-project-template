# 04 — Background jobs, events and integrations

**Status:** DRAFT
**Decisions:** B7, B8, B9

The directions that connect the core to the outside world are in the [architecture](../architecture.md#rules-for-the-code-structure). This section gives the implementation rules. The basis is the topics of the course ["Asynchronous architecture"](https://tough-dev.school/architecture): idempotency, retries, outbox, dead-letter queue, event evolution.

## River jobs

- Each side effect outside the database is a job, added in the same transaction as the action (`InsertTx`). Examples: sending a message, an LLM call, an email, an outgoing webhook.
- Delivery is "at least once", so each handler is idempotent: a second run does not create a duplicate and does not send a second message. Method: an idempotency key in the job arguments and a unique constraint on the result.
- Retries use an exponential delay and a limit on attempts. Jobs that use all attempts go to the dead-letter queue with an alert. They do not go to the archive silently.
- Queues are separate by type of work (import, notifications, integrations, AI), so heavy work does not delay urgent work ([tech stack](../tech-stack.md#platform-services)).
- Each external call has a timeout.

## Events

- The event goes to the `events` table in the transaction of the action (outbox, [ADR 0003](../adr/0003-analytics-outbox-and-sinks.md)). A job publishes it.
- The name is in the past tense: `order.created`, `user.email_changed`. The event dictionary is `docs/dev/events.md`.
- An event contains: `id`, name, schema version, time, `account_id` and the data.
- Internal events carry IDs and the changed fields. The sink adapter builds the full state for its sink.
- Evolution: only add fields to an event. A breaking change is a new event version. Publish the old version while somebody reads it.
- Event order is guaranteed only inside one account. A handler must not rely on more than that.

## Adapters

- An adapter translates the source format into a core command and back. It contains no business rules.
- Incoming webhook: check the signature → save the raw request with an idempotency key → return `200` → a job processes it.
- Each input is linked to an account. If this fails, the input goes to the dead-letter queue.
- Account secrets (API tokens, keys) are encrypted in the database. The encryption key is outside the database.
- The sink adapter decides which fields go outside. The personal data of users never goes outside (B10).
