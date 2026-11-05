# 33 — Tooling & Dev Environment

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D11** — the dev loop is built on the fast toolchain:
  - independent checks run in parallel via **npm-run-all2**;
  - editor setup includes tsgo (`"typescript.experimental.useTsgo": true`) and the oxc extension (`oxc.oxc-vscode`);
  - git hooks run only fast checks — a slow pre-commit hook is a defect.
