# 26 — Forms & Validation

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### Native validation first

Source: [You Don't Need JS](https://lyra.horse/blog/2025/08/you-dont-need-js/) (lyra.horse).

- Use semantic input types (`email`, `tel`, `url`, …) — built-in validation plus the right mobile keyboard, no custom logic.
- `pattern` attribute for regex-level constraints.
- Style validation state with `:user-valid` / `:user-invalid` (not `:valid` / `:invalid`) so errors appear only after user interaction, never prematurely.
- Accessibility caveat: never communicate validation state by color alone.
- Native validation is the first layer, not the whole story — schema validation (shared with the API layer, see [24](24-data-api-layer.md)) remains the source of truth.
