# 0008 — Dashboard on Mantine

Date: 2026-10-28
Status: accepted

## Context

[ADR 0001](0001-dashboard-ui-library.md) put the dashboard on a ready-made UI library and moved the choice between Ant Design and Mantine to the first slice. Slice 1 already needs the dashboard. Requirements from ADR 0001: the theme uses the same CSS variables as the public page tokens (D7 of the frontend standard); the dashboard is easy to use on a phone (D8, D14); shared packages do not depend on the dashboard library.

## Decision

The dashboard (`web/apps/dashboard`) uses **Mantine**: components, forms, notifications; charts come from the Mantine Charts package. This is the owner's decision.

## Consequences

- The Mantine theme uses CSS variables, so we can build it from the shared tokens, as ADR 0001 requires.
- Mantine has fewer ready-made complex tables and CRUD screens than Ant Design. If the standard `Table` is not sufficient for tables with sorting and filters, we add a separate table library.
- Refine for typical CRUD screens: only when necessary. We decide when the first such screen appears.
- We pin the Mantine major version when we create `web/apps/dashboard`, and we use the current version at that time.
