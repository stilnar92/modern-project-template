# 09 — Tech Stack

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.
**Fixed:** React + TypeScript.

---

## Collected points

### From Locked Decisions

- **D11** — toolchain feedback speed is a first-class selection criterion: humans and AI agents run the check loop constantly, so latency multiplies. Defaults: **tsgo** for typechecking (tsc fallback where unsupported), **oxlint** for linting, **oxfmt** for formatting, `mise` tasks for parallel checks (ADR 0010; the source suggests npm-run-all2). Source: [Fastest Frontend Tooling for Humans & AI](https://cpojer.net/posts/fastest-frontend-tooling).

### Other collected points

- **React Compiler** is part of the build pipeline — automatic memoization; see [30 — Performance](30-performance.md) for the usage rules it implies.
