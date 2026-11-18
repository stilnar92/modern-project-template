# 0024 — A job or a message can arrive twice: delivery guarantees and the idempotent receiver

From PR #107 · 12.11.2026 · Topics: messaging, reliability

## Problem

A background job sends an event to the analytics tool. The send succeeds, and then the worker process stops before it records "done". After the restart, the queue has a job that looks unfinished. The queue must either run it again or drop it. Both selections have a price.

## Predict first

A job has two steps: it calls an external tool, and then it marks itself as done. The process stops between the two steps. What does the external tool see if the queue runs the job again? What does it see if the queue never runs a job twice, and the process stops before the first step?

## Theory

A sender and a receiver talk over a network. The sender knows that a message arrived only when a confirmation comes back. If no confirmation comes, the sender cannot know which of two things happened: the message was lost, or the confirmation was lost. So the sender has only two honest strategies:

- **At-most-once.** Do not send again. No duplicates, but a message can be lost.
- **At-least-once.** Send again until a confirmation comes. No message is lost, but a message can arrive twice.

**Exactly-once delivery** does not exist as a property of the network. What exists is an exactly-once *effect*: at-least-once delivery and a receiver that gives the same result for the same message twice. Such a receiver is **idempotent**.

You know this from HTTP. `PUT /price` with the body `500` is idempotent: two requests leave the price at 500. `POST /payments` is not: two requests make two payments. A double click on a submit button is the same problem in the browser. There are two ways to make a receiver idempotent:

- **A naturally idempotent operation.** "Set the price to 500", not "add 100 to the price". "Delete all expired rows", not "delete the next 10 rows".
- **An idempotency key.** The sender puts a unique ID into the message. The receiver saves the ID with a unique constraint, in the same transaction as the effect. A second message with the same ID breaks the constraint, and the receiver skips it. Payment APIs use this: the `Idempotency-Key` header.

Sources:

- Kleppmann, "Designing Data-Intensive Applications", chapter 11, sections "Acknowledgments and redelivery" and "Idempotence".
- [Designing robust and predictable APIs with idempotency](https://stripe.com/blog/idempotency), Stripe: why a client must retry, and how a key makes the retry safe.
- [Implementing Stripe-like Idempotency Keys in Postgres](https://brandur.org/idempotency-keys), Brandur Leach: the key, the unique constraint and the transaction, step by step.
- [River: writing reliable workers](https://riverqueue.com/docs/reliable-workers): the at-least-once rule of our queue.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| At-most-once | Mark the job as done before the work, or never retry | A stop in the middle loses the work, and nobody knows |
| At-least-once, and the operation is naturally idempotent | The queue retries. The handler sets a state and does not add to it | Not each operation has this form: "send an email" does not |
| At-least-once, and an idempotency key | The queue retries. The receiver saves the key with a unique constraint and skips a known key | A key in each message, and a table or a column for the keys |
| Trust the queue to deliver exactly once | No code in the handler | No queue can promise this for an effect outside itself |

## Edge cases

| What if | What happens |
| --- | --- |
| The worker stops in the middle of a send | The job runs again, so the receiver can get the same event twice |
| A deploy stops the worker while jobs run | The worker waits for the running jobs. After a timeout the queue cancels them, and they run again later |
| An event has two receivers, and one of them fails | The whole job fails and retries. The receiver that worked gets the event a second time |
| The job runs again, but its event is gone | The account was deleted with its events. The job logs this and ends with success, so it does not retry forever |
| The hourly clean-up of expired sessions runs twice | The second run finds nothing to delete. This operation is naturally idempotent |

## Decision

All our background work is at-least-once, and each handler must be idempotent. The method in the backend standard is an idempotency key in the job arguments and a unique constraint on the result. For events, the rule is on the receiver: a sink must accept the same event twice. Each event has its own `id`, which a sink can use as the key.

We selected this because a lost event or a lost message to a client is worse for us than a duplicate that the receiver can find. The cost is that each new handler and each new sink needs this check in review. No sink exists yet, so no test proves the rule on a real receiver.

This would be wrong for an effect that cannot be made idempotent, when the receiver takes no key and a duplicate does real damage. Then at-most-once with an alert on failures can be the smaller problem.

Record: [backend standard 04](../../dev/backend-standard/04-async-integrations.md), [ADR 0014](../../dev/adr/0014-river-and-outbox.md) (stress round, "A job runs twice after a crash").

## In our code

1. `core/internal/events/events.go`, the `Sink` interface: its comment is the contract "accept the same event twice". Then `DispatchWorker.Work`: one failed sink fails the whole job.
2. `core/cmd/worker/main.go`, the call to `client.Stop`: the comment says why a deploy can run a job twice.
3. `core/internal/accounts/cleanup.go`: a job that is naturally idempotent.
4. [Lesson 0023](0023-dual-write-outbox.md): why the outbox makes the second delivery possible.

## Check yourself

1. Why can a sender not tell a lost message from a lost confirmation, and what follows from this?
2. A job does "add 100 bonus points to the client". Why is it dangerous under at-least-once, and how do you make it safe?
3. A colleague says: "Our queue is in the same database as the data, so a job runs exactly once." What is wrong?

<details><summary>Answers</summary>

1. In both cases the sender sees the same thing: no answer. It must select a strategy before it knows the cause. If it sends again, a duplicate is possible. If it does not, a loss is possible. No third strategy exists.
2. A second run adds 100 points again. Give the job a unique key, and save the key with a unique constraint in the same transaction as the points. The second run breaks the constraint and changes nothing.
3. The same database makes the *insert* of the job atomic with the data. It does not make the *run* atomic with an effect outside the database. The worker can call the external tool and then stop before it marks the job as done. The queue then runs the job again.

</details>
