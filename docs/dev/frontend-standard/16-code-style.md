# 16 — Code Style & Conventions

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D11** — linter is **oxlint** (runs ESLint plugins internally, so the ecosystem is preserved), formatter is **oxfmt** (Prettier-compatible, falls back to Prettier for non-JS/TS files). Migration guides: [ESLint → oxlint](https://oxc.rs/docs/guide/usage/linter/migrate-from-eslint.md), [Prettier → oxfmt](https://oxc.rs/docs/guide/usage/formatter/migrate-from-prettier.md). Editor: `oxc.oxc-vscode` extension.
- **D12** — lint configuration principles:
  1. errors only — no warnings;
  2. strict, consistent code style;
  3. bug-prevention rules on (ban constructions that lead to bugs);
  4. only fast rules — slow rules off;
  5. subjective style rules off.
  Every diagnostic is either fixed or explicitly suppressed with a written reason.

### From errors-as-values ([27](27-error-handling.md))

- If the errors-as-values convention is adopted, add a `no-unhandled-error` lint rule (e.g. from `lintcn`) to flag discarded `Error`-typed return values — the one case the type system alone can't catch. Deliberate discards use the `void` keyword. Source: [errore](https://errore.org/).
