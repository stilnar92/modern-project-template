# 0017 — A user reports an error: finding one request in logs, traces and error reports

From PR #105 · 12.11.2026 · Topics: observability, logging, tracing

## Problem

A user writes: "I saved the price and got an error." The server handled thousands of requests that hour. We must find the records of that one request in the logs, in the error tracker and in the traces. Without a shared key, we can only search by time and guess.

## Predict first

Two requests run at the same time. Each writes three log records, and the records mix in the output. The records have a timestamp and a message, and nothing else. How do you know which records are from the same request?

## Theory

The usual answer is **correlation**: each signal of one request carries the same ID.

- A **request ID** is a random value that the server makes when the request arrives. The server returns it in a response header, usually `X-Request-ID`. The client can show it on an error screen, and the user can send it to support.
- A **structured log** is a record with named fields, usually JSON, not a line of free text. A tool can then filter by `request_id = "..."`. In Node.js terms: `pino` output, not `console.log`.
- A **trace** is the path of one request through all services. A **span** is one step in it, with a start, an end and a parent. All spans of one request share a **trace ID**. It looks like the network waterfall in DevTools, but across servers. The W3C `traceparent` header carries the trace ID from one service to the next.

The ID must go with the request through each function call. In Go it goes in `context.Context`, a value that each function receives as its first argument. It is like `AsyncLocalStorage` in Node.js, but explicit. A log call that does not get the context cannot add the ID.

A request ID from a client is input from outside. If the server writes it into the log as it is, a client can put fake records or very long text into the log. Accept it only in a strict form.

Sources:

- [Traces](https://opentelemetry.io/docs/concepts/signals/traces/), the OpenTelemetry documentation: traces, spans and span context.
- [Trace Context](https://www.w3.org/TR/trace-context/), W3C Recommendation: the `traceparent` header.
- [Structured Logging with slog](https://go.dev/blog/slog), the Go blog: the standard logger, handlers and the context.
- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html): what not to log, and log injection.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Text logs, search by time | No ID. You read all records near the time of the report | Slow, and it fails when requests run at the same time |
| Request ID in the logs only | One ID in each record and in the response header | The error tracker and the traces are separate searches |
| Trace ID only | The tracing library makes the ID and the spans | The user has no value to send to support, unless the response also returns it |
| Request ID and trace ID in each signal | Each log record has both. The error report has the request ID | A wrapper for the logger, and a rule for each log call |

## Edge cases

| What if | What happens |
| --- | --- |
| The client sends its own `X-Request-ID` | The server keeps it only if it has 1 to 64 letters, digits, `.`, `_` or `-`. A valid ID comes back unchanged |
| The client sends `x\n{"level":"ERROR"}` as the ID | The server makes a new random ID. The client cannot write fake records into the log |
| Code calls `logger.Info`, not `logger.InfoContext` | The record has no IDs, because the logger did not get the context |
| No trace collector runs | The spans stay in the process. The logs still get a trace ID. The export starts only when `OTEL_EXPORTER_OTLP_ENDPOINT` is set |
| The work runs in a background job, not in a request | The record gets `job_id` and `job_kind` from the context |
| The error goes to Sentry | The report gets the tag `request_id` |
| A URL has personal data in the query string | The access log writes the path, not the query string or the body |

## Decision

Each request gets a request ID and a trace ID. The logs are `slog` in JSON. A wrapper around the log handler reads the context and adds `request_id`, `trace_id` and `span_id` to each record. The response has the `X-Request-ID` header, and each Sentry report has the same ID as a tag. OpenTelemetry makes the traces.

This is too much for one process with no background work: a request ID is sufficient there. We already have an API, a worker and external calls, and the standard wants one trace from the HTTP request through the job to the external call.

Not verified yet: the export to a real trace collector, because no collector runs.

Record: [backend standard 06](../../dev/backend-standard/06-observability.md). The libraries are in [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md).

## In our code

1. `core/internal/platform/httpmw/httpmw.go`, `RequestID`: the check of the client ID and the response header.
2. `core/internal/platform/observability/context.go`, `ContextHandler`: the IDs go from the context into each log record.
3. `core/internal/platform/observability/observability.go`: `Setup` creates the logger, the tracer and Sentry. `ReportError` adds the tag.
4. `core/internal/platform/observability/observability_test.go`, `TestContextHandlerAddsIDs`: the proof.

## Check yourself

1. Why does the server not accept each `X-Request-ID` that a client sends?
2. A log record of a request has no `request_id`. What is the most probable cause in the code?
3. A user sends you a request ID from an error screen. Where do you search, and what connects the results?

<details><summary>Answers</summary>

1. The ID goes into each log record. Free text from a client could contain a line break and a fake record, or thousands of characters. The strict form (1 to 64 safe characters) makes this impossible.
2. The code used a log method without the context, for example `Info` and not `InfoContext`. The wrapper reads the IDs from the context, so without the context it has nothing to add.
3. Filter the logs by `request_id`. Search Sentry by the tag `request_id`. Each log record also has the `trace_id`, which finds the trace and its spans when a collector runs.

</details>
