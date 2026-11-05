# Capability specifications

This layout implements [ADR 0002](../docs/dev/adr/0002-specs-openspec-hybrid.md).

- `specs/<capability>/spec.md` describes accepted current behavior.
- `changes/<change-id>/` contains a proposal, design, tasks, and capability deltas.
- `changes/archive/` holds completed changes after their deltas are applied.
- `templates/` contains source templates. Replace angle-bracket fields when copying them.

Use `/to-spec <change-id> <request>` to draft a change and `/to-tickets <change-id>`
to divide it into verifiable slices. Both skills are in `.claude/skills/`.
`to-tickets` prepares ticket bodies locally by default. Publishing requires authorization.
An instruction to create no new issues keeps all work local, even for approved proposals.

The owner approves proposals. Only the orchestrator changes accepted specs.
After a merged implementation passes its scenarios, apply the deltas to the current
specs and move the completed change to `changes/archive/<date>-<change-id>/`.
Keep links to the implementation PR and scenario checks in the archived tasks.

No capability specs exist yet. Empty folders do not claim implemented behavior.
