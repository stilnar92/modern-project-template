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

### Repeated triggers of mutations (double tap / double click)

Source: DantePay incident DPP-4666 (2026-11) — a fast double tap on the "save to address book" star sent two `POST /template/create`; the second one came back `409 Already saved` and the user saw an error for an action that had succeeded.

- **Every user-triggered mutation is guarded against re-entry** — create/delete/pay/submit buttons, toggles, "repeat" actions. Assume the user taps twice; on mobile a double tap is normal, not an edge case.
- **The guard must be synchronous — a `useRef` flag, not `useState`** — `disabled={isBusy}` only takes effect after the re-render, so two taps in the same frame both read `isBusy === false` from the same closure and both fire. Pattern: `if (lockRef.current) return; lockRef.current = true; try { await mutate() } finally { lockRef.current = false }`. Keep the state flag only for the visual `disabled`/spinner.
- **Data-fetching libraries don't save you here by default** — a mutation hook's `isPending` is state too; check it through a ref or use the library's own dedup/`mutationKey` guarantees, and say which one in code review.
- **The server stays the last line of defence** — non-idempotent operations (payments, transfers) also carry an idempotency key; a duplicate/"already exists" response (e.g. 409) is mapped to the final state ("saved"), not shown as an error, when it means the user's intent is already fulfilled.
- **Acceptance check for any such control**: double click / fast double tap → exactly one request in the network log, one success toast, no error.
