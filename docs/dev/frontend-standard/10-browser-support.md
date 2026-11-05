# 10 — Browser Support & Compatibility

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D1** — target modern evergreen browsers with an explicit browserslist; polyfills only where that policy requires them, never by default.
- **D2** — modern CSS by default (nesting, container queries, `:has()`, logical properties, etc.) within the bounds of the browser support policy. This document defines exactly what "modern CSS" may rely on.
- **D9** — cross-browser behavior is always considered: engine differences (Safari/WebKit quirks, etc.) are accounted for during development and verified during testing.
