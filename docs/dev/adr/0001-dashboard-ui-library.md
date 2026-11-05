# 0001 — Dashboard on a ready-made UI library, public pages follow the frontend standard

Date: 2026-10-28
Status: accepted

## Context

The project frontend follows the [frontend standard](../frontend-standard/README.md). The standard applies in full, and each deviation needs an ADR ([02 §5](../frontend-standard/02-documentation-standards.md)). The standard requires its own UI kit: Tailwind CSS + tailwind-variants (D5), styles only through `tv()` without `className` (D6), behavior from a headless library (Ariakit or React Aria, [20](../frontend-standard/20-component-system.md)).

The dashboard is mostly tables, forms, filters and charts. The [tech stack](../tech-stack.md) selects a ready-made UI library for it (Ant Design or Mantine). This makes the dashboard faster to build, and SEO and bundle size are not critical for it. Public pages must be light and can carry a custom brand, so a custom kit by the standard is correct there.

## Decision

- **The dashboard** (`web/dashboard`) uses Ant Design or Mantine. We select the library in the first slice. D5, D6 and the headless component policy from [20](../frontend-standard/20-component-system.md) do not apply to the dashboard.
- **Public pages and shared packages** follow the standard in full: Tailwind + tailwind-variants, headless behavior, tokens as CSS variables.
- **All other decisions of the standard apply to the dashboard too:** D1 (browsers), D3 (UI separate from logic), D4 (API types only from code generation), D7 (tokens as CSS variables; the library theme uses the same variables), D8 and D14 (responsive layout: users work from a phone), D10–D13 (checks, tooling, lint, observability).
- **Shared packages do not depend on the dashboard UI library.** The logic of shared modules (API client, session, media upload) is hooks and services without markup (D3). If both the dashboard and the public page need a component, the hook is shared, and each app renders it in its own way.

## Consequences

- The dashboard is faster to build from ready-made tables, forms and filters.
- The monorepo has two style systems. Imports from `web/dashboard` into public pages and shared packages are forbidden. Boundary rules check this ([13](../frontend-standard/13-architecture-enforcement.md)).
- The dashboard and the public pages share a theme through common tokens, not through common components.
- If the library causes problems (size, customization, phone layout), a new ADR reviews this decision.
