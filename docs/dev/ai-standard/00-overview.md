# 00 — AI standard overview

**Status:** DRAFT

## Fixed decisions

| ID | Decision | Section |
| --- | --- | --- |
| A1 | The human decides; agents do the work and check it. Only the human approves OpenSpec proposals, ADRs and merges. A12 tells which decisions go to the human | 01, 05 |
| A2 | Specs before code: OpenSpec as the format, Pocock's skills as the steps ([ADR 0002](../adr/0002-specs-openspec-hybrid.md)) | 01 |
| A3 | Each task goes through the phases "explore → plan → code → commit". No code without an approved plan | 01 |
| A4 | An agent in a fresh context checks the plan first, then the human | 01, 03 |
| A5 | When possible, hooks enforce the process rules, not only text | 02 |
| A6 | `CLAUDE.md` files are short and layered. The details are in documents, through links | 02 |
| A7 | One agent writes, another agent checks in a fresh context. An AI review before merge is required for the backend and the frontend (D10 of the frontend standard applies to the whole repository) | 03 |
| A8 | "Done" means checks that passed, with evidence: tests, linters, code generation, spec scenarios, E2E | 03 |
| A9 | Parallel work: an orchestrator and subagents by folder. Only the orchestrator changes shared files. No more than 2–3 streams | 01 |
| A10 | Each agent mistake becomes a guard: a hook, a line in `CLAUDE.md`, a skill fix | 02 |
| A11 | Install an external skill or plugin only after you read its code and check its license and last update date: skills run code | 04 |
| A12 | The agent makes cheap decisions and records them in plain words. Expensive decisions go to the human as a card: the point, the options, the consequences, a recommendation | 05 |
| A13 | Learning comes through tasks: a problem that had a real choice becomes a lesson in the same PR. A lesson has the problem, the theory with sources, the options, the edge cases and the decision | 05 |
| A14 | All project text is in English and follows ASD-STE100: documents, issues, PRs, PR comments, commits, code comments, agent replies | 06 |
| A15 | A subagent executes one assignment by a fixed contract and brief. The orchestrator checks it in the first minute and accepts its work only on evidence: the diff, the checks, the deviation journal | 07 |

## Sections

| # | Section | File | Status |
| --- | --- | --- | --- |
| 01 | Workflow | [01-workflow.md](01-workflow.md) | DRAFT |
| 02 | Claude Code setup | [02-harness.md](02-harness.md) | COLLECTING |
| 03 | Review and verification | [03-review-verification.md](03-review-verification.md) | DRAFT |
| 04 | Tools | [04-tools.md](04-tools.md) | DRAFT |
| 05 | Human in the loop: decisions and learning | [05-human-in-the-loop.md](05-human-in-the-loop.md) | DRAFT |
| 06 | Language and writing style | [06-writing-style.md](06-writing-style.md) | DRAFT |
| 07 | Delegation to subagents | [07-delegation.md](07-delegation.md) | DRAFT |

## Sources

Selected from the owner's bookmarks on 28.10.2026 and checked for currency.

| Source | What we take |
| --- | --- |
| [Agentic Design Patterns](https://mokevnin.github.io/agentic-coding-design-patterns/ru/explore-plan-code-commit.html) (book in Russian, 2026) | The phases "explore → plan → code → commit", patterns and anti-patterns, a review of OpenSpec and Pocock's skills |
| [Hook-driven dev workflows with Claude Code](https://nick-tune.me/blog/2026-02-28-hook-driven-dev-workflows-with-claude-code/) (Nick Tune, 2026) | The process as a state machine; hooks block actions outside the current phase |
| [OpenSpec](https://openspec.dev/) | The spec format |
| Matt Pocock's skills | Clarification, specs, slicing, TDD, review ([development process](../dev-process.md#matt-pocock-skills-and-our-additions)) |
