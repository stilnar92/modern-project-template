# 06 — AI Harness

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

- **OpenSpec integration** — the harness setup includes [OpenSpec](https://openspec.dev/) wiring for Claude Code (the `/openspec:proposal` command and the `openspec/` directory conventions), since spec-driven development ([05](05-spec-driven-development.md)) runs through it.
- **Browser automation for agent verification** — agents need a way to drive the running app for end-to-end verification ([37](37-task-playbook.md) step 5): Playwright MCP as the baseline; [Stagehand](https://github.com/browserbase/stagehand) as a candidate where AI-resilient interaction helps (see [29](29-testing-strategy.md)).

### Automation candidates (exploration — where/how not yet decided)

Common shape: *docs define the rule → harness drives the app → AI judges against the rule* (D10 extended to runtime). In rough value order:

1. **Spec scenarios → executable verification** — OpenSpec scenarios are already GIVEN/WHEN/THEN ([05](05-spec-driven-development.md)); an agent walks them against the running app after implementation.
2. **Responsive sweep** — screenshots of key screens across many widths (including in-between zones per D14), AI-reviewed against the fluid-first rules.
3. **A11y walk** — axe checks + agent-driven keyboard-only navigation of APG patterns ([22](22-accessibility.md)).
4. **Cross-browser smoke** — same flows on WebKit/Firefox in CI (D9).
5. **Regression triage** — observability alerts ([32](32-observability.md)) wake an agent to investigate and file a reproduction.
