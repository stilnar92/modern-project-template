# 30 — Performance

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D13** — performance regressions are caught in two places: in CI (budgets, lab checks) and in the field via the real-user metrics (Core Web Vitals RUM) that ship with every project from day one.

### React rendering performance

- **React Compiler is enabled** — automatic memoization at build time. Manual `useMemo` / `useCallback` / `React.memo` are no longer written by default; they're reserved for cases the compiler can't cover, and only when a measured problem exists.
- The compiler assumes Rules-of-React-compliant code — the compiler's lint rules (via `eslint-plugin-react-hooks` / oxlint equivalent) are part of the lint config ([16](16-code-style.md), D12: bug-prevention rules on).
- Interlocks with the bulletproof-components rule ([20](20-component-system.md) pt. 10): `useMemo` is a performance hint the compiler/React may discard — never use it for correctness. The compiler makes this assumption real.
- **Re-render diagnostics are dev-time tooling**: why-did-you-render, React Scan, React DevTools Profiler for finding avoidable re-renders. Never shipped in the production bundle.
- Measure before optimizing: re-render fixes follow profiler evidence, not intuition.
- **Bundle size is checked in CI** (e.g. `size-limit`) — the budget from this document is enforced automatically, not reviewed by eye. Source: [React Tricks](https://molefrog.com/notes/react-tricks).

### Images & layout shift (CLS)

Source: [How to Improve Lighthouse Scores by Avoiding img Layout Shifts](https://www.builder.io/blog/how-to-improve-lighthouse-scores-by-avoiding-img-layout-shifts) (Adam Bradley, Builder.io).

- **Every image reserves its space before it loads** — an image with no reserved space is a CLS defect. Two sanctioned methods:
  - CSS `aspect-ratio` + `width: 100%` — responsive, no exact dimensions needed;
  - HTML `width`/`height` attributes + `img { height: auto }` in CSS.
- **`object-fit: cover`** when the reserved box's ratio may differ from the file's — fills the box without distortion.
- **Placeholder background** (`background-color` on the `img`) so the reserved box reads as intentional while loading.
- These rules belong in the base image component/styles, not left to per-usage discipline — ties into the CLS part of the Core Web Vitals budget (D13).

### JS-level optimization (hot paths only)

Source: [Optimizing JavaScript](https://romgrk.com/posts/optimizing-javascript) (romgrk). ⚠️ Scope guard: these techniques trade readability for speed and apply **only to profiler-proven hot paths** — default code stays idiomatic ([16](16-code-style.md), P5); the author himself leads with "benchmark before optimizing."

- **Avoid work first** — memoization, lazy evaluation, incremental processing beat micro-optimizing code that shouldn't run at all.
- **Numeric enums over string enums** in hot comparisons — integer compare vs O(n) character compare.
- **Monomorphic shapes** — create objects of one hot path with identical property order/types; polymorphic/megamorphic call sites lose engine optimization (even React prop order participates in shape).
- **Imperative loops over `.map().filter().reduce()` chains** in hot paths — intermediate arrays cost allocation + GC; same for `Object.keys/values/entries`.
- **Avoid indirection** — Proxies block optimization; cache deep property chains (`a.b.c.d` repeatedly) into locals; dynamic dispatch resists inlining.
- **Cache-friendly data** — sequential access over random; small working sets; `TypedArray` over object arrays where applicable.
- **Avoid indexing into very large objects** (hashmap mode + cache misses) — restructure to arrays beforehand.
- **Strings**: slices/concat are pointer-based and cheap; mutations (`.trim()`, `.replace()`) force copies — defer mutation. Watch: a small slice can retain a huge source string from GC; force-copy when needed.
- **Specialization** — hoist likely-condition checks out of hot loops; branch prediction rewards it.
- **Right data structure beats micro-tuning** — `Set` for membership (O(1) vs O(n) `Array.includes`), linked lists/priority queues/tries where access patterns fit.
- **Benchmark honestly** — top-down (biggest impact first), production mode, doubt surprising results, clean browser profile (React DevTools distorts traces); `eval`/`new Function` tricks exist but are a last resort with CSP/security implications.

### Synthetic performance testing

Source: [React Performance Monitoring](https://3perf.com/blog/react-monitoring/) (Ivan Akulov).

- **INP is measured synthetically** with Lighthouse in timespan mode — it's the metric built for interaction responsiveness.
- **Synthetic runs are automated** via Playwright/Puppeteer (measurements collected in-page, extracted via `page.evaluate()`).
- **Compare against a rolling baseline** — average of the last 4–40 runs, never just the previous run; single-run comparisons drown in environmental noise.
- Field-side interaction monitoring lives in [32 — Observability](32-observability.md).
