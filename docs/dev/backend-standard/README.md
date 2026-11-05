# Backend standard

28.10.2026 · @ilnar

The rules for the Go core: modules, API, data, background jobs and integrations, security, observability and tests. The standard is for agents and people. Each rule must be clear without the question "what did they mean?".

**How to use it**

1. Start with [00-overview.md](00-overview.md): the fixed decisions B1–B12 and the map of sections.
2. The [tech stack](../tech-stack.md#backend-and-data) selects the technologies and libraries. This standard gives the rules for using them. The core boundaries are in the [architecture](../architecture.md).
3. A deviation from the standard needs an ADR in [`docs/dev/adr/`](../adr/). A silent deviation is a defect, even if the code is good.

Section statuses are the same as in the [frontend standard](../frontend-standard/02-documentation-standards.md#3-area-lifecycle): `COLLECTING` — points are collected; `DRAFT` — the rules are written and wait for practice; `DONE` — the rules apply.
