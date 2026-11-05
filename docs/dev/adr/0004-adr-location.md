# 0004 — ADRs are in `docs/dev/adr/`

Date: 2026-10-28
Status: accepted

## Context

The [frontend standard](../frontend-standard/02-documentation-standards.md) keeps ADRs in `docs/adr/`. It was written for a repository where `docs/` contains only development documentation. Our `docs/` has development documentation (`docs/dev/`) next to other folders, for example `docs/learning/`, and architecture decisions belong to development.

## Decision

ADRs are in `docs/dev/adr/NNNN-<slug>.md`. The format, the numbering and the rules follow the standard: nobody edits an accepted ADR; a new ADR replaces it.

## Consequences

- All development documentation is in one folder. An agent that works on code reads `docs/dev/` and the specs.
- Skills and templates that write ADRs to `docs/adr/` by default need the new path.
