# 29 — Testing Strategy

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D9** — cross-browser verification is part of the testing strategy: engine differences are not just considered during development but verified by tests (e.g. e2e runs against multiple engines, WebKit included).

### AI-resilient e2e automation

Source: [Stagehand](https://github.com/browserbase/stagehand) (Browserbase) — candidate tool, decision at elaboration.

- **Hybrid AI + deterministic automation** on Playwright-style/CDP primitives: `act()` (natural-language single actions), `agent()` (multi-step), `extract()` (schema-typed data from pages), `observe()` — code where flows are predictable, AI where they aren't.
- **Self-healing selectors** — AI adapts when markup changes, attacking the main source of e2e brittleness; **auto-caching** means repeat runs replay cached actions with no LLM inference (cost + determinism), AI re-engages only when the page actually changed.
- ⚠️ Guard for elaboration: AI belongs in *navigation/interaction resilience*, not in assertions — assertions stay deterministic, or tests become flaky in a new way. Also weigh LLM cost/nondeterminism in CI against classic Playwright, which remains the baseline ([30](30-performance.md) synthetic runs already assume it).
