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

## Alerts

| Signal | Why it is important |
| --- | --- |
| Jobs in the dead-letter queue | A message to a user or an event is lost |
| The age of the oldest job increases | The queue cannot keep up |
| Webhooks from an external provider fail | Users do not get their data or messages |
| A certificate expires in less than 14 days | Public pages will stop opening |
| A sudden increase in the cost of AI tasks | The limits did not work |
