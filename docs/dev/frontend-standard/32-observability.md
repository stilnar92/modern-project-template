# 32 — Observability

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D13** — observability is wired in at scaffold time, never retrofitted. The day-one kit:
  - **error tracking** — unhandled errors and error-boundary catches are reported;
  - **structured frontend logging** — levels, event naming conventions, PII rules;
  - **real-user metrics** — Core Web Vitals RUM plus custom app metrics;
  - **product analytics**.

### Monitoring interaction performance (React RUM)

Source: [React Performance Monitoring](https://3perf.com/blog/react-monitoring/) (Ivan Akulov). Loading performance has ready-made RUM; post-load interaction performance must be instrumented by hand:

- **Instrument key interactions only** — pick the highest-impact interactions (submits, navigations, critical clicks) and time those; monitoring everything drowns the signal.
- **`performance.mark()` / `performance.measure()` as the foundation** — visible in DevTools, readable app-wide via `PerformanceObserver`.
- **Measure to the next frame** (afterFrame-style), not inside React: `useEffect` timing is inaccurate (fires synchronously on user input in React 18), and the React Profiler API misses `useLayoutEffect` and layout-recalculation costs. Caveat: `useTransition`/`useDeferredValue` break next-frame timing — document per-interaction which method applies.
- **Send interaction timings as custom transactions** to the error/performance tracker (e.g. Sentry) for trend visualization.
- **Watch for degradation via `PerformanceObserver`** on `longtask` (JS > 50 ms) and `event` entries.
- **Alerts + cadence** — notifications on metric degradation, plus a recurring (weekly/bi-weekly) review of trends; metrics nobody looks at don't exist.

### When and what to log

Source: [When and what should I be logging?](https://blog.sentry.io/logging-best-practices/) (Sentry). Intro-level restatement of established practice; fills the "what/when" gap under D13's structured-logging requirement.

#### What to log

- **Runtime decisions** — every branch taken on a condition (`isMobile`, feature flags, experiment buckets) is logged with its resolved value, including the feature-toggle configuration, so the code path of a given session can be reconstructed after the fact.
- **Steps of multi-step flows** — a long sequential process logs each step and its result, not only the final outcome.
- **Audit events** — when a user or the system acts, log who did what: every create, update, and delete of an entity.
- **Context around errors** — the error alone is not enough; the retry attempt, status code, URL, and anything else needed to debug travel with it. Complements the typed errors carrying structured data in [27](27-error-handling.md).

#### How to write logs

- **Structured, not stringly** — log records are objects (JSON), not formatted messages. Each record answers *what happened*, *who did it*, *when it happened*.
- **Context accumulates** — a session may start unauthenticated; once identity is known (`userId` after sign-in), it flows into every subsequent log. Related records are linked across a flow via tracing / correlation IDs.

#### What not to log

- **Excessive logging** — not every function call and not every line; noise buries signal.
- **Secrets and personal data** — see [31 — Security](31-security.md).
- **Large payloads with no defined purpose** — the constraint is storage and processing cost.
