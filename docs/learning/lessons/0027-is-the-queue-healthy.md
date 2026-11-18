# 0027 — Background work is stuck and no user sees an error: how to know that a queue is healthy

From PR #114 · 12.11.2026 · Topics: observability, background jobs

## Problem

A failed request shows an error to a user at once. A failed background job shows nothing. A message to a client does not go out, and nobody sees an error page. We must learn that the queue is stuck before a user tells us. And when one job fails at night, we must find its log records in the morning.

## Predict first

You have an alert: "the number of failed jobs is above 0". The queue is empty and healthy, so the query behind the metric returns no rows, and the metric reports nothing. Then three jobs fail. Does the alert see the change from 0 to 3? What did it see before the failure?

## Theory

A **metric** is a number that a process reports again and again. A monitoring server (Prometheus) asks each process for its numbers at an interval. This is a **scrape**. Two types of metric answer two different questions:

- **Counter.** A sum that only increases: requests served, jobs done. You read its rate: "40 jobs each minute". It says how much work *happened*.
- **Gauge.** A value now, which goes up and down: memory in use, jobs that wait. It says what the state *is*.

A queue needs gauges. A counter of "jobs done" cannot tell a quiet night from a stuck worker: both show a rate of zero. Two gauges tell them apart:

- **Queue length.** How many jobs wait. It is easy to read, but a long queue can be normal after a large import.
- **Age of the oldest job that waits.** How long the most unlucky job has waited. If this number increases, the workers cannot keep up or do not work at all. This is the signal that is nearest to what a user feels.

Where does a gauge get its value? The process can count in memory, or it can read the real state at each scrape. A count in memory starts from zero after a restart, and it sees only its own process.

An alert rule compares a series of values with a limit. **No data is not zero.** If a metric reports nothing while all is well, the rule has nothing to compare, and some rules never see the first bad value as a change. So a metric that can exist must always report a value, also 0.

A metric says *that* something is wrong. A **log** says *what*. Many jobs write log records at the same time, so each record needs the ID of its job. Without the ID, you cannot put the records of one job together. A request ID does the same for HTTP requests.

Sources:

- [Prometheus: metric types](https://prometheus.io/docs/concepts/metric_types/): the counter and the gauge.
- [Prometheus: instrumentation](https://prometheus.io/docs/practices/instrumentation/), sections "Offline processing" and "Avoid missing metrics": what to measure in background work, and why to export a 0.
- "Site Reliability Engineering" (Google), [chapter 6, "Monitoring Distributed Systems"](https://sre.google/sre-book/monitoring-distributed-systems/): the four golden signals. A queue that fills is "saturation".

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Wait for a user report | No code | We learn about a stuck queue hours later |
| The metrics of the queue library | A counter and a duration for each job that ran | They show the work that happened, not the jobs that wait |
| A count in the memory of the worker | The worker adds and subtracts when it takes a job | Wrong after a restart. It does not see the jobs of other workers |
| A gauge that reads the jobs table at each scrape | One SQL query for each scrape gives the count and the oldest age for each state | A database query on each scrape. It needs a timeout |

## Edge cases

| What if | What happens |
| --- | --- |
| The queue is empty | The first version reported no series at all. Now each state reports a value, also 0, so an alert has data to compare |
| The database is slow during a scrape | The query has a timeout of 2 seconds, so the metrics endpoint stays fast |
| A job waits for a retry that is one hour away | The age counts only jobs whose start time has passed. A job that is planned for later is not "old" |
| Three jobs are in the "discarded" state | Three jobs failed on each attempt. An operator reads their errors in the jobs table, fixes the cause and retries them |
| A job fails at night | Its `job_id` finds all its log records. Its `trace_id` connects them to the request that made the job |
| The aggregate in the SQL has its filter in the wrong place | The first query had a syntax error. The test with a real database found it, and a test with a fake database could not |

## Decision

The worker exports two gauges that read the jobs table at each scrape: the number of jobs in each state, and the age of the oldest job in the states that wait. The state of the queue is in Postgres ([lesson 0025](0025-queue-in-postgres.md)), so the table is the only correct source. The alerts of the backend standard read them: "discarded above 0" and "the age of the oldest job increases". The metrics of the queue library stay for job counts and durations.

Each log record of a job has `job_id` and `job_kind`. A middleware puts them into the context of the job, and the log handler reads them from there. In Go, a context carries values through the call chain of one job, like `AsyncLocalStorage` in Node.js.

The zero rule came from the real stack, not from a unit test: the metrics page of a worker with an empty queue had no queue lines. This selection would be wrong for a jobs table so large that a count on each scrape is slow. Then the gauge needs a cheaper query or a cached value.

Record: [backend standard 06](../../dev/backend-standard/06-observability.md), [ADR 0014](../../dev/adr/0014-river-and-outbox.md) (consequences).

## In our code

1. `core/internal/platform/jobs/jobs.go`, `RegisterQueueMetrics`: the two gauges, the SQL query, and the loop that reports a 0 for each state. Then `LogContext`.
2. `core/internal/platform/jobs/jobs_test.go`, `TestQueueMetrics`: two jobs, one of them 90 seconds old, and a 0 for "discarded".
3. `core/internal/platform/observability/context.go`, `WithJob` and `ContextHandler.Handle`: how the job ID gets into each log record.
4. `core/cmd/worker/main.go`: the middleware list in `newRiver` and the call to `RegisterQueueMetrics`.

## Check yourself

1. The counter "jobs done" shows zero new jobs for the last hour. Why can you not tell if this is good or bad, and which metric tells you?
2. Why does the queue gauge read the database and not count in the worker process?
3. Why can an alert on "discarded jobs above 0" fail when the metric reports nothing for an empty queue?

<details><summary>Answers</summary>

1. Zero finished jobs means "no work came" or "work came and nobody does it". The counter looks the same in both cases. The age of the oldest job that waits tells them apart: it is 0 on a quiet night, and it increases when the workers are stuck.
2. The state of the queue is in the jobs table. A count in memory is lost at a restart, and one worker does not see the jobs of the other workers. A query at each scrape always gives the real state.
3. The alert compares values. With no series it has no value: it sees "no data", which is not 0. When all is well, nobody can tell a healthy queue from a broken metric. A metric that always reports 0 proves that the measurement works.

</details>
