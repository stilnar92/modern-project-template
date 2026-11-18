# 00 — Overview: Areas of the Standard

This is the master map of the AI-First Frontend Standard. Every area below will become its own document. Statuses: `TODO` (not started), `COLLECTING` (file exists, raw points being gathered), `DRAFT` (being written), `DONE` (approved and in force).

Areas are ordered top-down: foundations first, then the AI development system, stack and architecture, code, design, application concerns, verification, and finally process.

---

## Locked Decisions

Binding decisions made upfront. Area documents must be written around them — they are constraints, not proposals. Each will be elaborated (with rationale and exact rules) in the area documents listed.

| ID | Decision | Area(s) |
|----|----------|---------|
| D1 | Target modern evergreen browsers with an explicit browserslist; polyfills only where that policy requires them, never by default. | 10 |
| D2 | Use modern CSS by default (nesting, container queries, `:has()`, logical properties, etc.) within the bounds of the browser support policy. | 10, 21 |
| D3 | Strict separation of UI and business logic is mandatory: components render; logic lives in hooks/services and is testable without the UI. | 12, 13 |
| D4 | Backend types are always generated from the backend contract (e.g. OpenAPI codegen) — never written by hand. | 24 |
| D5 | The styling stack is Tailwind CSS + tailwind-variants. In this project: the public pages only. The dashboard uses Mantine ([ADR 0001](../adr/0001-dashboard-ui-library.md), [ADR 0008](../adr/0008-dashboard-mantine.md)). | 21 |
| D6 | No ad-hoc `className` styling in components — all component styling is expressed through tailwind-variants definitions (`tv()`), including one-off cases. In this project: the public pages only (see D5). | 20, 21 |
| D7 | All design values flow through CSS custom properties; design tokens are CSS variables and Tailwind consumes them. | 19, 21 |
| D8 | Responsive by default: every component and screen is designed and reviewed across all breakpoints — desktop-only is never an acceptable initial state. | 18, 21 |
| D9 | Cross-browser behavior is always considered: differences between engines are accounted for during development and verified during testing. | 10, 29 |
| D10 | All code — especially AI-generated code — must pass architectural verification before merge: automated static checks (boundary/dependency rules in lint and CI) **and** an AI review pass against the architecture and spec documents. | 08, 13, 17 |
| D11 | Tooling is selected for feedback speed — humans and AI agents run the check loop constantly, so latency is a first-class selection criterion. Default toolchain: **tsgo** for typechecking (tsc as fallback where unsupported), **oxlint** for linting, **oxfmt** for formatting, `mise` tasks for parallel checks (ADR 0010). [^1] | 09, 15, 16, 33 |
| D12 | Lint policy: errors only — no warnings; strict, consistent code style; bug-prevention rules on; slow rules off; subjective style rules off. Every diagnostic is either fixed or explicitly suppressed with a reason — a noisy check is a broken check. [^1] | 16, 17 |
| D13 | Observability from day one: error tracking, structured frontend logging, and real-user metrics (Core Web Vitals + key app metrics) are wired in at scaffold time, never retrofitted. | 30, 32 |
| D14 | Responsiveness is fluid-first: layouts must be correct at **every** viewport width, not only at named breakpoints — in-between zones (split-screen, tablets, resized windows) are first-class. Prefer intrinsic layout (Grid/Flexbox, container queries, content-driven sizing) over viewport breakpoints, and never switch to the mobile layout earlier than the content forces it. [^2] | 18, 21 |

---

## A. Foundation

| # | Area | File | Scope | Status |
|---|------|------|-------|--------|
| 01 | Vision & Principles | `01-vision-principles.md` | The AI-first philosophy: spec- and documentation-driven development, docs as the single source of truth, and the core principles every other document must obey. Defines what "AI-first" means in practice for this standard. | DRAFT |
| 02 | Documentation Standards | `02-documentation-standards.md` | How these documents themselves are written: structure, tone, formatting rules, versioning, and the ADR (Architecture Decision Record) process for recording every non-trivial decision. | DRAFT |
| 03 | Glossary & Ubiquitous Language | `03-glossary.md` | Shared vocabulary so humans and AI agents mean the same thing by the same word. Term definitions, naming of domain concepts, banned ambiguous terms. | TODO |

## B. AI Development System

The machinery of AI-first development: how work is specified, how agents are configured, and the repeatable workflows they follow.

| # | Area | File | Scope | Status |
|---|------|------|-------|--------|
| 04 | AI Collaboration Model | `04-ai-collaboration.md` | The division of labor between human and AI agents: who decides, who implements, who reviews. Guardrails (what agents may never do), context rules, and how agents keep docs and code in sync. | TODO |
| 05 | Spec-Driven Development | `05-spec-driven-development.md` | No feature without a spec, managed with [OpenSpec](https://openspec.dev/): capability-scoped living specs (requirement–scenario format) evolved via change proposals (proposal/design/tasks/spec deltas), the propose → approve → implement → archive lifecycle, and traceability between spec, code, tests, and AI review. | DRAFT |
| 06 | AI Harness | `06-ai-harness.md` | Configuration of the agent runtime (Claude Code): what goes in CLAUDE.md and what doesn't, settings and permissions, hooks (pre-commit checks, guardrail enforcement), MCP servers, and subagent definitions. | COLLECTING |
| 07 | Skills | `07-skills.md` | Reusable agent skills for recurring procedures: when a workflow deserves to become a skill, skill file structure and naming, project skills vs global skills, and how skills are tested and maintained. | TODO |
| 08 | Agent Workflows | `08-agent-workflows.md` | The standard loops agents follow: brainstorm → spec → plan → implement → review → verify. When to use plan mode, when to spawn subagents, the TDD loop, and the review checkpoints — including the mandatory AI architecture review (D10). | COLLECTING |
| 37 | Task Playbook | `37-task-playbook.md` | The lifecycle of a single task from the GitHub board to done: what to understand first, mapping the affected modules, verifying assumptions, choosing an MVP slice, and the check/test/verify steps before a task counts as complete. (Numbered 37 to avoid renumbering existing files; belongs to this group.) | COLLECTING |

## C. Stack & Architecture

| # | Area | File | Scope | Status |
|---|------|------|-------|--------|
| 09 | Tech Stack | `09-tech-stack.md` | The chosen runtime, framework layer (e.g. Vite SPA vs Next.js), build tool, package manager, and Node version policy — each with the selection criteria that justify it. Toolchain feedback speed is a first-class selection criterion (D11). | COLLECTING |
| 10 | Browser Support & Compatibility | `10-browser-support.md` | The browserslist target, polyfill and transpilation policy (D1), what "modern CSS by default" may rely on (D2), engine-difference checklist (Safari/WebKit quirks, etc.), and how cross-browser correctness is verified (D9). | COLLECTING |
| 11 | Dependency Policy | `11-dependency-policy.md` | Rules for adding, updating, and removing dependencies: evaluation criteria (size, maintenance, types quality), approval process, lockfile discipline, and audit cadence. | TODO |
| 12 | Architecture | `12-architecture.md` | The architectural style (e.g. feature-sliced layers), module boundaries, allowed dependency directions between layers, and rules for what code may import what. Enforces the mandatory UI / business-logic separation (D3): components render, logic lives in hooks/services. | COLLECTING |
| 13 | Architecture Enforcement | `13-architecture-enforcement.md` | How architecture rules become executable checks (D10). Automated layer: oxlint `no-restricted-imports` rules for layer and import violations, architecture tests, forbidden-pattern lint rules (e.g. business logic in components per D3), all wired into CI. AI layer: a review agent that checks each change against the architecture, structure, and spec documents using a defined checklist, with findings that block or warn. | COLLECTING |
| 14 | Project Structure | `14-project-structure.md` | Concrete folder layout, file naming conventions, colocation rules (tests, styles, types next to code), and where each kind of artifact lives. | COLLECTING |

## D. Language & Code Quality

| # | Area | File | Scope | Status |
|---|------|------|-------|--------|
| 15 | TypeScript Standards | `15-typescript-standards.md` | tsconfig strictness settings (no legacy flags), type-definition conventions, `any`/`unknown`/assertion policy, generics guidance, and rules for typing component props and API contracts. Typechecking runs on tsgo (D11), including the editor integration. | COLLECTING |
| 16 | Code Style & Conventions | `16-code-style.md` | Linter and formatter configuration: oxlint + oxfmt (D11), rule selection per the lint policy (D12 — errors only, fast rules only, subjective rules off). Naming conventions (files, variables, components, hooks), import ordering, and comment policy. | COLLECTING |
| 17 | Quality Gates & Definition of Done | `17-quality-gates.md` | The checklist code must pass to be considered complete: typecheck, lint, tests, architectural checks and AI review (D10), human/agent review, docs and spec updated. What blocks a merge and what merely warns. | COLLECTING |

## E. Design & UI

| # | Area | File | Scope | Status |
|---|------|------|-------|--------|
| 18 | Design Principles | `18-design-principles.md` | The visual language: layout rules, density, hierarchy, motion philosophy. Responsive strategy — responsive by default (D8) and fluid-first (D14): every screen is designed for the full width range from the start, breakpoints are driven by content rather than device classes, and in-between viewport zones are first-class. | COLLECTING |
| 19 | Design Tokens | `19-design-tokens.md` | The token system: color, typography, spacing, radii, shadows, motion. Token naming scheme, tiering (primitive → semantic → component), theming (light/dark). Tokens are implemented as CSS custom properties consumed by Tailwind (D7). | COLLECTING |
| 20 | Component System | `20-component-system.md` | Component API design: props conventions, composition patterns, controlled vs uncontrolled, and the policy on UI libraries (headless vs styled, build vs buy). Variants are the only styling surface — components expose tailwind-variants APIs, never raw `className` styling (D6). | COLLECTING |
| 21 | Styling Approach | `21-styling-approach.md` | The styling stack: Tailwind CSS + tailwind-variants (D5). All styling via `tv()` definitions, no ad-hoc `className` strings (D6); tokens as CSS variables (D7); modern CSS by default (D2); responsive utilities required by default (D8), implemented fluid-first — intrinsic layout and container queries before viewport breakpoints (D14). | COLLECTING |
| 22 | Accessibility (a11y) | `22-accessibility.md` | The WCAG conformance target, semantic HTML rules, keyboard and focus management requirements, ARIA usage policy, and how a11y is tested. | COLLECTING |

## F. Application Concerns

| # | Area | File | Scope | Status |
|---|------|------|-------|--------|
| 23 | State Management | `23-state-management.md` | The boundaries between local, global, and server state; which tool owns each kind; rules for when state may be lifted or globalized. | COLLECTING |
| 24 | Data & API Layer | `24-data-api-layer.md` | Data fetching and caching strategy, the API client, request/response error propagation, and mocking for development. API types are always generated from the backend contract — codegen pipeline, regeneration workflow, and the ban on hand-written API types (D4). | COLLECTING |
| 25 | Routing & Navigation | `25-routing.md` | Route structure and naming, route-level code splitting, guards/redirect patterns, and URL-as-state conventions (query params, deep linking). | DRAFT |
| 26 | Forms & Validation | `26-forms-validation.md` | Form library policy, schema-based validation (shared with the API layer where possible), error message UX, and submission/disabled-state patterns. | COLLECTING |
| 27 | Error Handling & Resilience | `27-error-handling.md` | Error boundary placement, fallback UI standards, retry policies, distinguishing expected vs unexpected errors, and what gets reported vs shown to users. | COLLECTING |
| 28 | Internationalization (i18n) | `28-i18n.md` | Whether and how projects support multiple locales: message catalog workflow, key naming, pluralization, and date/number formatting rules. | COLLECTING |

## G. Verification & Operations

| # | Area | File | Scope | Status |
|---|------|------|-------|--------|
| 29 | Testing Strategy | `29-testing-strategy.md` | The test pyramid for frontend: unit, component, and end-to-end layers, chosen tools, what must be tested vs what must not, and coverage policy. Includes cross-browser verification (D9). | COLLECTING |
| 30 | Performance | `30-performance.md` | Performance budgets (bundle size, Core Web Vitals targets), code-splitting and lazy-loading rules, image/font policy, and how regressions are caught — in CI and in the field via RUM metrics (D13). | COLLECTING |
| 31 | Security | `31-security.md` | XSS prevention rules, secrets handling (never in the client bundle), dependency vulnerability auditing, CSP, and auth-token storage policy. | COLLECTING |
| 32 | Observability | `32-observability.md` | The day-one observability kit (D13): error tracking, structured frontend logging (levels, event naming, PII rules), real-user metrics — Core Web Vitals RUM plus custom app metrics — and product analytics. What is captured, where it's sent, and how issues are triaged from reports. | COLLECTING |

## H. Process & Delivery

| # | Area | File | Scope | Status |
|---|------|------|-------|--------|
| 33 | Tooling & Dev Environment | `33-tooling-dev-env.md` | Standard scripts and `mise` tasks (parallel, D11), editor configuration and extensions (tsgo, oxc), git hooks (pre-commit checks), and the local setup path from clone to running app. | COLLECTING |
| 34 | Git Workflow | `34-git-workflow.md` | Branching model, commit message conventions, PR size and description rules, and the code review process (including AI-assisted review). | TODO |
| 35 | CI/CD & Deployment | `35-ci-cd.md` | Pipeline stages (typecheck, lint, architectural checks, test, build, deploy), environment promotion flow, preview deployments, and release/rollback process. | TODO |
| 36 | Environments & Configuration | `36-environments-config.md` | Environment variable conventions and validation, per-environment configuration, and feature flag policy. | TODO |

---

## Suggested writing order

Foundation (01–03) and the AI Development System (04–08) come first — they define *how* everything else gets built, so every later document and every line of code flows through them. Then 09–14 (stack, browser support, architecture and its enforcement) unlock scaffolding the actual project. The rest can be written in parallel with early development, but 17 (Quality Gates) and 29 (Testing) should be in force before the first feature is built.

---

## References

[^1]: [Fastest Frontend Tooling for Humans & AI](https://cpojer.net/posts/fastest-frontend-tooling) — basis for D11 (fast toolchain: tsgo, oxlint, oxfmt, npm-run-all2) and D12 (lint policy principles). Migration guides: [Prettier → oxfmt](https://oxc.rs/docs/guide/usage/formatter/migrate-from-prettier.md), [ESLint → oxlint](https://oxc.rs/docs/guide/usage/linter/migrate-from-eslint.md).

[^2]: [The Too Early Breakpoint](https://ishadeed.com/article/too-early-breakpoint/) (Ahmad Shadeed) — basis for D14: don't switch to mobile layouts prematurely; in-between viewport zones (split-screen, tablets, resized windows, link previews) are real usage; prefer container queries, Grid/Flexbox, and content-driven layout over early viewport breakpoints.
