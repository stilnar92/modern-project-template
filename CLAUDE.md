# modern-project-template

A project template: a Go core, a React dashboard, an Astro site, Postgres with RLS, quality gates and an agent harness. Replace this paragraph with your product description. Product: [docs/dev/brief.md](docs/dev/brief.md). Domain terms: [CONTEXT.md](CONTEXT.md). Architecture: [docs/dev/architecture.md](docs/dev/architecture.md).

This file is for all agents. `core/CLAUDE.md` has the Go rules. `web/CLAUDE.md` has the frontend rules.

## Which documents are rules

- In force: the `CLAUDE.md` files, the ADRs with `Status: accepted`, the UX standard, and each standard or section that a `CLAUDE.md` links. A newer accepted ADR wins over an older one and over a standard.
- A standard marked DRAFT is in force where a `CLAUDE.md` links it. A COLLECTING or TODO section is a collection of ideas, not a rule. If it conflicts with a rule, follow the rule and report the conflict.
- The stack that the repository really uses: [tech stack, "In force now"](docs/dev/tech-stack.md#in-force-now).

## Commands

| Command | What it does |
| --- | --- |
| `mise install` | Install the tool versions from `mise.toml`: Go, Node, pnpm, linters. Change a version only there |
| `mise tasks` | List all commands. Run one with `mise run <task>`. The tasks are in `mise.toml` |
| `mise run setup` | Bring a new worktree or session to a checked green baseline: dependencies, `.env`, the local stack, a smoke test. Run it first |
| `mise run up` / `mise run down` | Start or stop the local environment ([deploy/README.md](deploy/README.md)) |
| `mise run check` | Every check: lint, tests and the generated code. CI runs the same tasks |
| `mise run lint` | All linters and architecture checks. CI runs the same |
| `mise run test` | Go tests (Postgres in Docker) and web tests |
| `mise run gen` | Generate code from `api/openapi.yaml`. Run after each contract change |
| `mise run test:e2e` | End-to-end tests, after `mise run up` |

## Checks

- A change is done only when `mise run check` passes.
- Do not edit generated files by hand: `core/internal/platform/openapi/openapi.gen.go` and `web/packages/api-client/src/schema.d.ts`. Change `api/openapi.yaml` and run `mise run gen`. CI fails when they differ.
- Do not disable a lint rule or lower a check to a warning to make a change pass. An exception covers one path and one rule, has a reason, and needs the owner.

## Process

- Phases: explore → plan → code → commit ([AI standard 01](docs/dev/ai-standard/01-workflow.md)). No code without an approved plan.
- A feature branch is `feat/<change-id>`. Its plan is `openspec/changes/<change-id>/proposal.md` with the line `Status: approved`. A change with a new screen or flow also needs `Design: approved` after the owner sees a prototype. Only the owner approves. Tooling and fixes use `chore/`, `fix/` or `docs/` branches.
- Do not push to `main`. Push with an explicit branch name (`git push -u origin <branch>`) and open a PR. A push without a branch name is blocked. Hooks enforce this ([AI standard 02](docs/dev/ai-standard/02-harness.md#hooks)).
- Only the orchestrator changes shared files: `api/openapi.yaml`, `core/migrations/`, `docs/dev/events.md`, `CONTEXT.md`, `docs/dev/THESAURUS.md`, `openspec/specs/`.
- Expensive decisions go to the owner as a decision card ([AI standard 05](docs/dev/ai-standard/05-human-in-the-loop.md#decision-card)). Record cheap decisions in the PR.
- The owner learns from the real problems of the tasks. When a change solved a problem that had a real choice, write a lesson in `docs/learning/lessons/` in the same PR ([format and rules](docs/learning/README.md)).
- No AI attribution in commits and PRs: no `Co-Authored-By`, no "Generated with".

## Language

All project text is in English and follows ASD-STE100: docs, issues, PRs, commits, code comments ([AI standard 06](docs/dev/ai-standard/06-writing-style.md)). Short sentences, active voice, no semicolons. Text for users of the product (UI, emails, messages) stays in the user's language.

## Engineering rules

- Select the simplest implementation that fully meets the current requirements.
- Use established, well-maintained libraries instead of custom code.
- Fix the cause, not the symptom.
- Suggest good practices, even if they need a refactoring.
- Inside `/v1`, API changes only add things (backend standard rule B4).
- A thing works only when it is installed and checked in the real environment. If you did not run it, say "not verified".
