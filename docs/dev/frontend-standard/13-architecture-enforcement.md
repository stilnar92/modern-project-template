# 13 — Architecture Enforcement

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D10** — all code (especially AI-generated) passes architectural verification before merge, in two layers:
  - **Automated:** boundary/dependency rules in lint and CI (in this project: `no-restricted-imports` rules in `web/.oxlintrc.json`), architecture tests, forbidden-pattern rules.
  - **AI:** a review pass that checks each change against the architecture, structure, and spec documents using a defined checklist; findings block or warn.
- **D3** — the UI / business-logic separation is one of the rules enforced mechanically here (e.g. forbidden-pattern rules for business logic inside components), not just by convention.
