# 0023 — Two systems must change together: the dual write and the outbox

From PR #107 · 12.11.2026 · Topics: data consistency, messaging

## Problem

A user changes a price. We must save the new price in our database and also tell another system about it: analytics now, notifications later. These are two different systems. One of the two writes can fail.

## Predict first

We save the price, the commit succeeds, and then the call to analytics fails. What is the state of each system? Now reverse the order: we call analytics first, and then the commit fails. What is the state now?

## Theory

This is the **dual write problem**. No transaction covers two systems, so each order of the two writes has a failure that leaves them different:

- **Commit, then send.** The send can fail. The change happened, and nobody knows about it.
- **Send, then commit.** The commit can fail. Others know about a change that did not happen.

The **transactional outbox** removes the second system from the write. You write the message into a table of your own database, in the same transaction as the change. Both rows commit or neither does. A separate process reads the table later and sends the message. If the send fails, the process tries again, because the row is still there.

The price: the message arrives later, not immediately, and it can arrive more than once. The receiver must accept the same message twice ([lesson 0024](0024-at-least-once-idempotency.md)).

Sources:

- Kleppmann, "Designing Data-Intensive Applications", chapter 11, section "Keeping Systems in Sync": why dual writes fail.
- [Transactional outbox](https://microservices.io/patterns/data/transactional-outbox.html), Chris Richardson: the pattern on one page.
- [Transactionally Staged Job Drains in Postgres](https://brandur.org/job-drain), Brandur Leach: the same idea with a job queue in Postgres.
- [River: transactional enqueueing](https://riverqueue.com/docs/transactional-enqueueing): how our queue does it.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Call the other system after the commit | Simple code | A failed call loses the event |
| Call it before the commit | Simple code | A failed commit sends a false event |
| Outbox with a sweep | The event row is in the transaction. A periodic process finds the rows that are not sent | A `dispatched_at` column and updates in a log table |
| Outbox with a job in the same transaction | The event row and a queue job commit together. The queue retries | It works only when the queue is in the same database |

## Edge cases

| What if | What happens |
| --- | --- |
| The action fails after the event is recorded | The transaction rolls back. No event and no job exist |
| The receiver is down for a day | The job retries with a growing pause. After the last attempt it is "discarded", an alert fires, and an operator retries it. The event stays in the table |
| The worker stops in the middle of a send | The job runs again, so the receiver can get the event twice |
| The events table becomes large | It is a plain table now. We partition it by month after 10 million rows, a slow report, or a retention rule |

## Decision

We use the outbox with a job in the same transaction, and no sweep. Our queue (River) is in Postgres ([lesson 0025](0025-queue-in-postgres.md)), so one transaction already gives the guarantee. A sweep would only repeat it.

This would be wrong if the queue moves to a separate broker, for example Redis. Then the job is a second system again, and we need the sweep.

Record: [ADR 0014](../../dev/adr/0014-river-and-outbox.md).

## In our code

1. `core/internal/events/events.go`, `Record`: the insert and `InsertTx` use the same `tx`.
2. `core/internal/events/events_test.go`, `TestRecordIsAtomic`: the proof that a rollback leaves nothing.
3. `core/migrations/00003_events.sql`: the table.
4. [ADR 0014](../../dev/adr/0014-river-and-outbox.md): the decision and its stress round.

## Check yourself

1. A colleague wants to send the event "after the commit, because it is simpler". Which failure do you describe to them?
2. Why does the outbox make "the same message twice" possible, and who must handle it?
3. We move the queue to Redis. What breaks in this design?

<details><summary>Answers</summary>

1. The commit succeeds and the send fails, for example because the process stops between the two. The price changed, and the other system never learns about it. Nothing retries the send, because nothing recorded that it was necessary.
2. The sender retries until it gets a confirmation. If the send worked and the confirmation was lost, the sender sends again. The receiver must handle it: the same event twice must give the same result as once.
3. The job is no longer in the transaction of the change. The change can commit while the job is lost, which is the dual write problem again. The design then needs a sweep that finds events with no sent mark.

</details>
