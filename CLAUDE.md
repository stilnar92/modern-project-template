# modern-project-template

A project template: a Go core, a React dashboard, an Astro site, Postgres with RLS, quality gates and an agent harness. Replace this paragraph with your product description. Product: [docs/dev/brief.md](docs/dev/brief.md). Domain terms: [CONTEXT.md](CONTEXT.md). Architecture: [docs/dev/architecture.md](docs/dev/architecture.md).

This file is for all agents. `core/CLAUDE.md` has the Go rules. `web/CLAUDE.md` has the frontend rules.

## Commands

| Command | What it does |
| --- | --- |
| `mise install` | Install the tool versions from `mise.toml`: Go, Node, pnpm, Task, linters. Change a version only there |
| `task up` / `task down` | Start or stop the local environment ([deploy/README.md](deploy/README.md)) |
| `task lint` | All linters and architecture checks. CI runs the same |
| `task test` | Go tests (Postgres in Docker) and web tests |
| `task gen` | Generate code from `api/openapi.yaml`. Run after each contract change |
| `task test:e2e` | End-to-end tests, after `task up` |

Run `task lint` and `task test` before each commit.

## Process

- Phases: explore → plan → code → commit ([AI standard 01](docs/dev/ai-standard/01-workflow.md)). No code without an approved plan.
- A feature branch is `feat/<change-id>`. Its plan is `openspec/changes/<change-id>/proposal.md` with the line `Status: approved`. Only the owner approves. Tooling and fixes use `chore/`, `fix/` or `docs/` branches.
- Do not push to `main`. Push with an explicit branch name (`git push -u origin <branch>`) and open a PR. A push without a branch name is blocked. Hooks enforce this ([AI standard 02](docs/dev/ai-standard/02-harness.md#hooks)).
- Only the orchestrator changes shared files: `api/openapi.yaml`, `core/migrations/`, `CONTEXT.md`, `openspec/specs/`.
- Expensive decisions go to the owner as a decision card ([AI standard 05](docs/dev/ai-standard/05-human-in-the-loop.md#decision-card)). Record cheap decisions in the PR.
- If the team keeps learning notes, write a debrief in `docs/learning/debriefs/` after each backend task and update `docs/learning/concepts.md` ([docs/learning](docs/learning/README.md)).
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
