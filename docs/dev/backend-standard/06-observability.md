# 06 — Observability

**Status:** DRAFT
**Decisions:** B10, B11

Connect observability when the repository is created, not later. The frontend part is in [32 of the frontend standard](../frontend-standard/32-observability.md). Its rules on "what and when to log" are shared and apply to the backend too.

## Logs

- `slog` in JSON. Each record has: `request_id`, `trace_id`, the account hash, and the job ID for background work.
- Log: decisions at branches (which adapter, which LLM provider), the steps of multi-step processes (import, data export), audit events, the context of errors (attempt, response code, source).
- Do not log: the personal data of users, secrets, large request bodies.

## Traces

- OpenTelemetry: a trace goes from the HTTP request through the River job to the external call. The job arguments carry the trace ID.

## Errors and metrics

- Errors go to Sentry (our own server or GlitchTip) with `request_id`.
- Prometheus metrics: queue length and the age of the oldest job, jobs in the dead-letter queue, undelivered webhooks, adapter errors, the time and cost of AI tasks, the response time of the public API.

## In the code

- `internal/platform/observability`: the logger with `request_id` and `trace_id` from the context, the tracer, the meter and Sentry ([ADR 0012](../adr/0012-infrastructure-libraries.md)). Traces go out over OTLP only when `OTEL_EXPORTER_OTLP_ENDPOINT` is set.
- `internal/platform/httpmw`: the request ID (`X-Request-ID`), the access log and panic recovery.
- `internal/platform/jobs`: each log record of a River job has `job_id` and `job_kind`. The worker exports `river_queue_jobs{state}` and `river_queue_oldest_age_seconds{state}` for the alerts "jobs in the dead-letter queue" (`state="discarded"` above 0) and "the age of the oldest job increases" (`state="available"`). otelriver adds job counts and durations.
- `/healthz` is liveness, `/readyz` checks the database. `/metrics` is on the internal `METRICS_PORT` of the api and the worker. The proxy does not route it.

## Alerts

| Signal | Why it is important |
| --- | --- |
| Jobs in the dead-letter queue | A message to a user or an event is lost |
| The age of the oldest job increases | The queue cannot keep up |
| Webhooks from an external provider fail | Users do not get their data or messages |
| A certificate expires in less than 14 days | Public pages will stop opening |
| A sudden increase in the cost of AI tasks | The limits did not work |
