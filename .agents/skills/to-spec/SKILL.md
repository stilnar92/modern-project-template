---
name: to-spec
description: Draft or revise an OpenSpec proposal for this project and capability deltas from a requested behavior change. Use when the owner asks for a new feature or a behavior change, before any code is written.
---

# Draft an OpenSpec change

Resolve the repository root before reading or writing paths below.
Read `docs/dev/adr/0002-specs-openspec-hybrid.md`, `docs/dev/brief.md`, `CONTEXT.md`, the affected
`openspec/specs/<capability>/spec.md`, and relevant implementation files.

Use a lowercase hyphenated change ID. For an existing change, preserve completed
work and explain scope changes. For a new change, copy the templates from
`openspec/templates/` into `openspec/changes/<change-id>/`:

- `proposal.md`: problem, scope, exclusions, impact, questions, and references.
- `design.md`: implementation evidence, approach, decisions, and checks.
- `tasks.md`: initial slices with scenario links.
- `specs/<capability>/spec.md`: deltas for each affected capability.

Keep the proposal at `Status: draft`. Only the owner approves it.
Use `ADDED Requirements`, `MODIFIED Requirements`, and `REMOVED Requirements`
sections as needed. Each requirement has named GIVEN/WHEN/THEN scenarios.
A modified requirement includes its full resulting text and all scenarios.
A removal states its reason and migration. Inside `/v1`, follow the additive API rule.

Record unresolved questions instead of inventing policy or data relationships.
Keep accepted specs and runtime code unchanged during drafting.

Finish when all template fields are replaced, every scope item maps to a requirement,
every requirement has an observable scenario, and each scenario has a planned check.
Report the change path and open owner decisions. Publication is a separate action.
