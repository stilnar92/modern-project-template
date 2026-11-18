# 0011 — UX rules as a project standard, not a skill

Date: 2026-11-12
Status: accepted

## Context

The frontend standard has only raw points about design, accessibility and forms ([18](../frontend-standard/18-design-principles.md), [22](../frontend-standard/22-accessibility.md), [26](../frontend-standard/26-forms-validation.md)). Agents write most of the UI. Without written UX rules, each agent makes its own choices about states, errors, focus and text.

The `ux-best-practices` skill of [CodeAlive-AI/ai-driven-development](https://github.com/CodeAlive-AI/ai-driven-development/tree/main/skills/ux-best-practices) (MIT) has a good base: WCAG 2.2 AA, laws of UX, design system patterns, verified sources. But a skill works only when an agent calls it. The owner wants the rules to apply to every UI change.

## Decision

- The UX rules are in the [UX standard](../ux-standard/README.md): 35 rules (U1–U35), a verification procedure and a review format.
- The topical reference files of the skill are copied without changes to `docs/dev/ux-standard/reference/`. We did not copy the `audit/` folder.
- [web/CLAUDE.md](../../../web/CLAUDE.md) links to the standard, so every agent that works in `web/` reads it.
- The AI review checklist ([AI standard 03](../ai-standard/03-review-verification.md#ai-review-checklist)) includes the UX standard for UI changes.
- We do not install the skill.

## Trade-off

- **Prefer:** consistency of the UI, accessibility, the same rules for every agent.
- **Give up:** a little speed on each UI task (states, keyboard and width checks).

## Stress round

Not necessary: the decision adds documentation only and is easy to change.

## Consequences

- An agent that changes the UI follows U1–U35 and checks the change in a browser.
- A rule that a tool can check must also get a gate (A10). Next step: an automated accessibility scan in the Playwright tests, for example `@axe-core/playwright`.
- When the source skill changes, we compare the reference copy with it and record the new commit in the UX standard.
