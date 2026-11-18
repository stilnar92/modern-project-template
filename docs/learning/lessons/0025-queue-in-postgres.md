# 0025 — Work that the user must not wait for and that a restart must not lose: a job queue in the database

From PR #107 · 12.11.2026 · Topics: background jobs, infrastructure choice

## Problem

Some work does not fit into a request: to send a login code, to make a photo preview, to run an AI task, to build a weekly summary. The user must not wait for it. And a restart of the server must not lose it. Some work has no request at all: each hour, something must delete the expired sessions.

## Predict first

In Node.js you can start a promise and not `await` it, so the response goes out at once. The server restarts one second later. What happens to the work? Now you put the work into a Redis queue after the database commit. Which failure is still possible?

## Theory

A **job queue** splits work into two steps. The request only *records* that the work is necessary. A separate **worker** process takes the record later and does the work. Celery in Django and BullMQ in Node.js are job queues. The record must be on a disk, not in the memory of a process, or a restart loses it.

A queue gives three more things:

- **Retries.** A failed job runs again after a pause. The pause grows with each attempt (a back-off), so a receiver that is down does not get a flood.
- **A dead-letter state.** A job that fails on each attempt stops. It stays in the queue with its errors, and a person decides what to do with it. Our queue calls this state "discarded".
- **Scheduled jobs.** The queue adds a job by itself at an interval, like cron or Celery beat.

Where is the queue? A **broker** (Redis, RabbitMQ) is a separate server made for this, and it is fast. Kafka is a different tool: a log of events that many readers can read again, not a list of tasks. The third place is a **table in your own database**. A worker takes a row with `SELECT ... FOR UPDATE SKIP LOCKED`: it locks one free row and skips the rows that other workers hold. So many workers can read one table and do not block each other.

The table is slower than a broker. But it gives one thing that a broker cannot give: the job is in the same transaction as the data. This removes the dual write problem of [lesson 0023](0023-dual-write-outbox.md).

Sources:

- Kleppmann, "Designing Data-Intensive Applications", chapter 11, section "Messaging Systems": brokers, logs, and how they compare to databases.
- [PostgreSQL: SELECT, the locking clause](https://www.postgresql.org/docs/current/sql-select.html#SQL-FOR-UPDATE-SHARE): `SKIP LOCKED` and its use for "a queue-like table".
- [River: job retries](https://riverqueue.com/docs/job-retries): the attempts, the back-off, and when a job is discarded.
- [River: periodic and cron jobs](https://riverqueue.com/docs/periodic-jobs): scheduled jobs and their limits.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Do the work in the request | No queue | The user waits. A slow external tool makes our API slow |
| Start it in the memory of the API process | The response goes out at once | A restart loses the work. No retries |
| A broker (Redis) | A separate fast server keeps the jobs | One more service to run. The data can commit while the job is lost, or the opposite |
| An event log (Kafka) | Many readers read the same events again | A heavy system. It solves a problem that we do not have |
| A table in Postgres (River) | The job is a row. It commits with the data | Slower than a broker. The queue adds load to the main database |

## Edge cases

| What if | What happens |
| --- | --- |
| The action rolls back after it added a job | The job row rolls back too. No job exists for a change that did not happen |
| The worker stops in the middle of a job | The job runs again later, so each job must accept a second run ([lesson 0024](0024-at-least-once-idempotency.md)) |
| A job fails on each attempt | The queue marks it "discarded". It stays in the table with its errors, an alert watches this state, and an operator retries it |
| The scheduled clean-up must touch all accounts | It is the only job that gets the system database role. All other jobs work for one account, which is in the job arguments |
| The worker restarts between two hourly runs | The schedule is in the memory of the worker. Our clean-up also runs at each start, so a restart does not skip it |
| A new version of the queue library changes its tables | The `migrate` step applies the migrations of the library after ours. The worker never changes the schema at startup |

## Decision

The queue is River, a job queue in Postgres. The main reason is the transaction: a job commits with the data that it is about. The second reason is fewer moving parts: no broker to run, to back up and to monitor. ADR 0012 rejected a second queue library next to River for the same reason.

The tables of River come from the same `migrate` command as our tables, through the migrator of River. We do not copy its SQL. The first scheduled job deletes expired sessions each hour.

This would be wrong at a load that a database table cannot carry, or if many independent systems must read the same events again. The first case needs a broker, the second needs a log like Kafka. Both bring the dual write problem back, so both need the sweep of lesson 0023.

Record: [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md), [ADR 0014](../../dev/adr/0014-river-and-outbox.md).

## In our code

1. `core/cmd/worker/main.go`, `newRiver`: each job kind and the scheduled job are registered here.
2. `core/internal/accounts/cleanup.go`: the smallest full job. The arguments, the worker, and `CleanupJob` with the interval.
3. `core/migrations/river.go`: the tables of River, applied by `migrate` and by the tests.
4. `core/internal/events/events.go`, `Record`: `InsertTx` puts a job into the transaction of the caller.

## Check yourself

1. Why is "start the work and do not wait for it" not sufficient for a login code?
2. Ten workers read one jobs table. Why do two of them not take the same job, and why do they not wait for each other?
3. What is a discarded job, and why does the queue not delete it or retry it forever?

<details><summary>Answers</summary>

1. The work is only in the memory of the process. A restart or a crash loses it, and nothing knows that it was necessary. Also nothing retries it if the sender of the code is down for a moment.
2. A worker takes a row with a lock. `SKIP LOCKED` tells the other workers to skip a locked row and take the next free one. So each job has one owner, and no worker waits for a lock.
3. It is a job that failed on each attempt. To delete it hides a lost message. To retry forever hides a problem that needs a person, and it loads the receiver. So the job stops and stays in the table, an alert fires, and an operator fixes the cause and retries it.

</details>
