# 24 — Data & API Layer

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D4** — API types are always **generated** from the backend contract (e.g. OpenAPI codegen); hand-written API types are banned. The codegen pipeline and regeneration workflow are part of the project setup, and generated types are the single source of truth for request/response shapes.

### Promise concurrency

Source: [Effective Promise Concurrency](https://www.builder.io/blog/promises) (Steve Sewell, Builder.io).

- **Independent requests are awaited concurrently** (`Promise.all` with destructuring, or instantiate-then-await) — sequential `await`s on independent requests are wasted latency.
- **But sequential is the default until concurrency is needed** — it's easier to reason about; concurrency is an optimization, and over-parallelizing causes network/resource thrashing.
- **Timeouts via `Promise.race` must pair with `AbortController`** — losing the race must cancel the pending request, not leave it hanging.
- **`Promise.any`** for redundant/unpredictable sources (first success wins) — used sparingly; don't hammer multiple sources by default.
- Mental model: JS awaits promises concurrently, it doesn't execute code in parallel — the wins come from overlapping I/O waits.
