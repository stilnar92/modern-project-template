# 01 — Vision & Principles

**Status:** DRAFT
**Depends on:** none (root document)
**Elaborated by:** every other document in this standard

---

## 1. Purpose

This standard exists to make frontend projects buildable, reviewable, and maintainable primarily by AI agents, under human direction, without loss of quality or architectural integrity over time.

It is a reusable, product-agnostic standard: it defines *how* projects are built, not *what* any particular product does. A new project adopts the standard as-is and records any deviation explicitly (see [02 — Documentation Standards](02-documentation-standards.md)).

## 2. What "AI-first" means here

AI-first does **not** mean "AI writes code fast and humans clean up." It means the entire development system — documentation, architecture, tooling, and process — is designed for AI agents as the primary implementers:

1. **Documentation is written for agent consumption.** Explicit, unambiguous, self-contained. An agent reading a document must be able to act on it without asking what was meant. Prose that only an experienced human could interpret correctly is a defect.
2. **The docs are the source of truth.** When code and documentation disagree, the documentation wins until it is deliberately changed. Agents and humans resolve conflicts by fixing one or the other — never by silently ignoring the docs.
3. **Every rule aspires to be enforceable.** A rule that exists only as prose will decay. Wherever possible, rules are backed by an automated check (lint, types, CI) or an AI review checklist (see [13 — Architecture Enforcement](13-architecture-enforcement.md)). When writing a new rule, always ask: *how will this be checked?*
4. **Humans decide, agents execute and verify.** Product decisions, architectural decisions, and approvals belong to humans. Implementation, verification, and enforcement are delegated to agents operating inside guardrails (see section 04 — AI Collaboration Model, not written yet).

## 3. Core principles

Every document, spec, and line of code in a project under this standard must obey these principles. Lower-numbered documents win conflicts (see §5).

### P1 — Spec before code
No feature, change, or fix beyond the trivial is implemented without a written spec that states the problem, the requirements, and the acceptance criteria. Code that cannot be traced back to a spec is unplanned work. *(Elaborated in [05](05-spec-driven-development.md).)*

### P2 — Single source of truth
Each fact — a rule, a decision, a definition — lives in exactly one document, and everything else links to it. Duplication is how documentation rots. The Locked Decisions register in [00 — Overview](00-overview.md) is the source of truth for upfront constraints D1–D10.

### P3 — Verification over trust
Nothing is "done" because an agent or a human says so. Done means the quality gates passed: typecheck, lint, tests, architectural checks, AI review, and spec acceptance criteria — with evidence. *(Elaborated in [17](17-quality-gates.md).)*

### P4 — Separation of concerns is non-negotiable
UI renders; business logic lives in hooks and services, testable without the UI (D3). Layer boundaries are defined once ([12](12-architecture.md)) and enforced mechanically ([13](13-architecture-enforcement.md)). "It was faster this way" is never a valid reason to cross a boundary.

### P5 — Consistency over preference
One way to do each thing. Styling goes through tailwind-variants (D5, D6), design values through CSS variables (D7), API types through codegen (D4). A second way of doing something already done one way is a defect, even if the second way is objectively fine.

### P6 — Explicit deviations only
Deviating from the standard is allowed — silently deviating is not. Every deviation is recorded as a decision with its rationale (ADR process, [02](02-documentation-standards.md)). An unrecorded deviation is treated as a bug regardless of code quality.

### P7 — Modern platform, defined support
Build on the modern web platform by default — modern CSS (D2), evergreen browsers (D1) — against an explicitly defined support target ([10](10-browser-support.md)). Neither nostalgia ("support everything") nor recklessness ("works in Chrome") is acceptable; the browserslist is the contract.

### P8 — Responsive and accessible by default
Every screen works across the full viewport range from its first version (D8), fluid-first (D14): intrinsic layout and container queries before viewport breakpoints, and the in-between zones — split-screen, tablets, resized windows — are first-class, not edge cases. Accessibility is part of building a component, not a cleanup phase ([22](22-accessibility.md)). Cross-browser behavior is considered during development, not discovered in production (D9).

### P9 — Small, reviewable increments
Work is sliced so that every change is small enough for meaningful review — by AI and by humans. A change too large to review is too large to merge.

### P10 — Fast feedback loops
Tooling is chosen for speed (D11): agents and humans run the typecheck–lint–format–test loop constantly, so every second of latency is multiplied hundreds of times a day. Prefer native-speed tools (tsgo, oxlint, oxfmt), run independent checks in parallel, and treat a slow check as a defect to fix or replace — not a cost to accept. Checks emit errors only (D12): a warning that can be ignored will be ignored, by humans and agents alike.

## 4. How the standard is applied to a project

1. The documentation set (this `docs/` directory) is copied into the new project before any code is written.
2. Stack- and product-specific choices left open by the standard are made and recorded as decisions.
3. The project is scaffolded following the documents in order: harness ([06](06-ai-harness.md)), tooling, architecture skeleton, enforcement checks, then features.
4. From that point on, every change flows through the agent workflow ([08](08-agent-workflows.md)): spec → plan → implement → review → verify.

## 5. Precedence

When documents conflict, resolve in this order (highest wins):

1. Locked Decisions (D1–D10) in [00 — Overview](00-overview.md)
2. This document (01)
3. Lower-numbered documents over higher-numbered ones
4. Documents over code, comments, or habit

A discovered conflict is itself a defect: fix the losing document, don't just work around it.
