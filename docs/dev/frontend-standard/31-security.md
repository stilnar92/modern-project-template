# 31 — Security

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### XSS & injection

- React's escaping is the baseline; **`dangerouslySetInnerHTML` is banned by default** — where rendering external HTML is unavoidable, it passes through a sanitizer (DOMPurify) and the usage carries a lint suppression with reason (D12 style).
- Never build HTML/DOM from string concatenation.
- **User-supplied URLs are validated** — `javascript:` and other dangerous schemes rejected before hitting `href`/`src`.
- **Trusted Types** where supported (modern-platform stance, D2) to make DOM-injection sinks compile-time-visible.
- The `eval` / `new Function` hot-path trick ([30](30-performance.md)) is CSP-hostile — security wins over micro-optimization by default.

### CSP & headers

- **Strict Content-Security-Policy** — nonce/hash-based, no `unsafe-inline` / `unsafe-eval`; `frame-ancestors` against clickjacking. Delivery of headers is a deployment concern (section 35, not written yet / hosting config).
- Third-party scripts: minimized, and pinned with **Subresource Integrity** when unavoidable.

### Secrets & tokens

- **Nothing secret ships in the client bundle** — any env var exposed to the client (e.g. `VITE_*`) is public by definition; CI SHOULD scan bundles for secret-looking strings.
- **Auth tokens**: prefer `httpOnly` + `Secure` + `SameSite` cookies over `localStorage` (XSS can't read `httpOnly`); if a token must live in JS, keep it in memory, never persisted.
- CSRF: covered by `SameSite` cookies + backend tokens where state-changing endpoints demand it.

### Dependencies (supply chain)

- Vulnerability audit in CI (`npm audit` / OSV) with a triage policy — not just a red light everyone ignores; lockfile discipline and install scripts policy belong to section 11 — Dependency Policy, not written yet.

### Boundaries

- **Client-side validation is UX, not security** — the server is the authority; nothing security-relevant is decided in the browser.
- PII rules for logging/analytics live in [32 — Observability](32-observability.md); sensitive server values tainting in [20](20-component-system.md) (bulletproof pt. 9).
