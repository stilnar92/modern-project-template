# 17 — Quality Gates & Definition of Done

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

The project-wide gates and where each one runs: [AI standard 03, Quality gates](../ai-standard/03-review-verification.md#quality-gates).

---

## Collected points

### From Locked Decisions

- **D10** — the gate includes architectural verification: automated boundary/dependency checks **and** the AI review pass against architecture and spec documents. Both are blocking.
- **D12** — gates operate in errors-only mode: any diagnostic fails the gate; there is no "warning" tier that can be merged past.
