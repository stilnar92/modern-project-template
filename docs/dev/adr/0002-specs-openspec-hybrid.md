# 0002 — Specs in OpenSpec format, steps done by Pocock skills

Date: 2026-10-28
Status: accepted

## Context

In the [development process](../dev-process.md), we had one spec per task (`docs/specs/NNNN-*.md`). Matt Pocock's skills (`to-spec`, `to-tickets`) wrote them. The [frontend standard](../frontend-standard/05-spec-driven-development.md) requires [OpenSpec](https://openspec.dev/): one living spec per capability, and changes to it as proposals.

Per-task specs pile up. The current behavior of a capability then spreads across tens of files, and an agent can easily rely on an old spec. A solo developer with parallel agents needs one current description of each capability most of all. Pocock's skills already cover clarification, slicing and TDD, so we keep them.

## Decision

- **Storage format: OpenSpec.** The current truth is `openspec/specs/<capability>/spec.md`: requirements and GIVEN/WHEN/THEN scenarios. Work goes into `openspec/changes/<id>/`: `proposal.md`, `design.md`, `tasks.md` and spec deltas.
- **Capabilities match the core modules** (`core/internal/<module>`), for example accounts or billing. Product modules get their capabilities when their turn comes.
- **Pocock's skills do the steps:** `grill-with-docs` clarifies the task, `to-spec` writes the proposal and the deltas, `to-tickets` writes `tasks.md`, `tdd` implements. We adapt the skill templates to OpenSpec files.
- **Order:** proposal → human approval → contract → code → gates → deltas merged into `openspec/specs/` and the change archived.
- **No spec is needed** for typos, dependency updates and a bug fix that restores behavior that is already described. The reproduction and a failing test are the spec ([05 §2](../frontend-standard/05-spec-driven-development.md)).
- We do not create `docs/specs/`.

## Consequences

- Each capability has one current description. Agents read it, not the task history.
- Each spec scenario is a test or an E2E check. The review compares the code with the proposal and the deltas.
- A task needs more files: proposal, tasks, deltas, and a merge of the deltas after the PR merge. We check this on the first core slice. If short proposals still get in the way, a new ADR reviews this decision.
- Pocock's skills and OpenSpec overlap in `to-spec` and `to-tickets`. We must rewrite the skill templates for the OpenSpec format once.
