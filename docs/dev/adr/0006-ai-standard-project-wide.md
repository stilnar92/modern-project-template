# 0006 — The AI standard applies to the whole repository

Date: 2026-10-28
Status: accepted

## Context

The [frontend standard](../frontend-standard/00-overview.md) contains group B, "AI Development System": 04 (roles of the human and agents), 05 (specs), 06 (agent setup), 07 (skills), 08 (agent workflows), 37 (task path). But agents write the backend and the documentation too, so the rules for agents must be the same in the whole repository. Two sets of rules about the same thing break the principle "one fact, one place" (P2 of the frontend standard).

## Decision

- The rules for agents in the whole repository are in the [AI standard](../ai-standard/README.md).
- The AI standard replaces sections 04, 06, 07, 08 and 37 of the frontend standard. The points collected there moved to the AI standard or are already covered.
- Section 05 (specs) stays the source of the spec format, with the change from [ADR 0002](0002-specs-openspec-hybrid.md). The AI standard links to it.
- D10 of the frontend standard (a required AI review before merge) applies to the whole repository. This is decision A7 of the AI standard.
- We do not edit the frontend standard files: we copied the standard as-is, and this ADR records the replacement.

## Consequences

- An agent on the backend and on the frontend uses one cycle, the same hooks and the same review.
- When we update the frontend standard from best-project, we compare group B with the AI standard and move new points into the AI standard.
