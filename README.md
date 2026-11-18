# modern-project-template

A template for a new product: a Go core, two web apps, Postgres with row-level security, background jobs, quality gates, CI, a docs site and a harness for AI agents. It contains the foundation only. There is no business domain in it.

## What is inside

| Part | Where | What it is |
| --- | --- | --- |
| Core | `core/` | Go: `cmd/api`, `cmd/worker`, `cmd/migrate`. Shared infrastructure in `internal/platform/`: config, database, HTTP middleware, health, jobs, cursor pagination, observability, Problem Details errors, test database |
| Core modules | `core/internal/accounts`, `core/internal/events` | Accounts and server sessions with CSRF protection ([ADR 0013](docs/dev/adr/0013-session-lookup-function.md)). The event log and the outbox on River ([ADR 0014](docs/dev/adr/0014-river-and-outbox.md)). No route uses them yet |
| Database | `core/migrations/` | Postgres 16, goose, sqlc. Three roles (`app_owner`, `app_app`, `app_system`) and account isolation with RLS ([ADR 0009](docs/dev/adr/0009-account-isolation-rls.md)) |
| Contract | `api/openapi.yaml` | OpenAPI 3.1. `mise run gen` generates the Go server, the TypeScript client and the zod schemas from it |
| Dashboard | `web/apps/dashboard` | Vite, React, Mantine, TanStack Query, react-router with lazy pages ([ADR 0008](docs/dev/adr/0008-dashboard-mantine.md)) |
| Site | `web/apps/site` | Astro with server rendering and Tailwind, no JavaScript in the browser |
| Shared web code | `web/packages/` | `@app/api-client` (generated client, Problem mapping), `@app/tokens` (design tokens, [ADR 0015](docs/dev/adr/0015-design-system-and-claude-design.md)), `@app/i18n` (Paraglide) |
| End-to-end tests | `web/e2e` | Playwright, with real phone viewports |
| Local environment | `deploy/` | Docker Compose: Postgres, S3 (SeaweedFS), Mailpit, Caddy and the core. Each git worktree gets its own ports ([deploy/README.md](deploy/README.md)) |
| Specs | `openspec/` | OpenSpec: current capability specs and change proposals ([ADR 0002](docs/dev/adr/0002-specs-openspec-hybrid.md)) |
| Docs site | `tools/docs-viewer` | VitePress over the Markdown in the repository, with Mermaid, the API page and an Excalidraw sketch editor ([ADR 0016](docs/dev/adr/0016-docs-viewer.md), [ADR 0017](docs/dev/adr/0017-sketch-editor.md)) |
| Agent harness | `CLAUDE.md`, `.claude/`, `AGENTS.md`, `.codex/`, `.agents/`, `.mcp.json` | Rules, hooks, skills and review agents for Claude Code and Codex ([AI standard](docs/dev/ai-standard/README.md)) |
| CI | `.github/` | Linters, architecture checks, tests, generated-code checks, end-to-end tests, a nightly job and a weekly audit. Issue and PR templates, Dependabot |
| Tools | `mise.toml` | Pinned tool versions and all commands ([ADR 0010](docs/dev/adr/0010-mise-single-toolchain.md)) |

## Start a new project

1. On GitHub, click **Use this template** and create your repository. Clone it.
2. Install [mise](https://mise.jdx.dev) and Docker Desktop. Run `mise install` in the repository root. mise installs Go, Node, pnpm and the linters with the versions from `mise.toml`.
3. Rename the Go module. Replace `github.com/stilnar92/modern-project-template` in each file that `git grep -l 'github.com/stilnar92/modern-project-template'` lists: `core/go.mod`, the Go imports, `core/Dockerfile`, `.github/pull_request_template.md` and `tools/docs-viewer/scripts/roots.mjs`. Then run `mise run gen` and `mise run build:core`.
4. Optional: rename the npm scope `@app/`. Replace it in each file that `git grep -l '@app/'` lists: the `package.json` files, the imports, `web/.oxlintrc.json`, `.github/workflows/ci.yml` and the docs. Then run `pnpm install` in `web/` and `mise run gen`.
5. Optional: rename the compose project and the database. `COMPOSE_PROJECT_NAME` is in `.env.example`. The database `app` and the roles `app_owner`, `app_app`, `app_system` are in `deploy/postgres/init/01-roles.sh`, `deploy/compose.yaml`, `.env.example`, `deploy/worktree-env.sh`, `mise.toml` (`db:dump`, `db:restore`) and the migrations.
6. Describe your product. Fill [docs/dev/brief.md](docs/dev/brief.md), the domain model [CONTEXT.md](CONTEXT.md), the glossary [docs/dev/THESAURUS.md](docs/dev/THESAURUS.md), [docs/dev/nfr.md](docs/dev/nfr.md) and [docs/dev/data-map.md](docs/dev/data-map.md). Replace the first lines of `CLAUDE.md` with your product description, then copy it to `AGENTS.md`.
7. The UI texts are in Russian (`web/packages/i18n/messages/ru.json`). To change the language, change the base locale in `web/packages/i18n/project.inlang/settings.json`, rename the message file, and update the tests that check the texts ([frontend standard 28](docs/dev/frontend-standard/28-i18n.md)).
8. Run `mise run setup` and `mise run check`. Both must pass before your first feature.

## Commands

| Command | What it does |
| --- | --- |
| `mise install` | Install the tool versions from `mise.toml` |
| `mise tasks` | Show all commands. Run one with `mise run <task>` |
| `mise run setup` | Bring a new clone or worktree to a checked green baseline: dependencies, `.env`, the local stack, a smoke test |
| `mise run up` / `mise run down` | Start or stop the local environment. `mise run up` creates `.env`, applies the migrations and starts the core on <http://api.localhost:8088> |
| `mise run check` | Every check: lint, tests and the generated code. CI runs the same tasks |
| `mise run lint` | All linters and architecture checks |
| `mise run test` | Go tests (Postgres in Docker), web tests and agent hook tests |
| `mise run gen` | Generate code from `api/openapi.yaml` and the SQL code with sqlc |
| `mise run test:e2e` | Playwright tests. Run `mise run up` first |
| `mise run dev` | The web dev servers: the dashboard on <http://localhost:5173> and the site on <http://localhost:4330> |
| `mise run docs` | Build the docs site and serve it on <http://127.0.0.1:4173> |
| `mise run mock` | A mock API from the contract on <http://127.0.0.1:4010> (Prism) |
| `mise run audit` | Known vulnerabilities in the dependencies. It needs the network |

The ports are for the main checkout. Another git worktree gets its own ports in `.env` ([ops](docs/dev/ops.md)).

## Layout

```text
api/openapi.yaml   API contract
core/              Go core: cmd/, internal/, migrations/, tools/archcheck
web/               pnpm workspace: apps/, packages/, e2e/
deploy/            Docker Compose and Caddy for the local environment
openspec/          capability specs and changes
tools/docs-viewer/ the docs site
docs/dev/          development docs, standards, ADRs
docs/learning/     lessons about the decisions in this foundation
mise.toml          tool versions and commands
```

The layout is in [ADR 0007](docs/dev/adr/0007-monorepo-layout.md).

## Documents

| Document | What it contains |
| --- | --- |
| [How it works](docs/dev/how-it-works.md) | The whole project on one page: parts, the path of a change, the gates |
| [Brief](docs/dev/brief.md) | The product on one page for agents. Fill it for your project |
| [Architecture](docs/dev/architecture.md) | C4 diagrams of the system |
| [Tech stack](docs/dev/tech-stack.md) | Technologies and the reasons for them. "In force now" lists what the code really uses |
| [Development process](docs/dev/dev-process.md), [how to run a task](docs/dev/how-to-run-a-task.md) | The task cycle and how the owner and the agents work |
| [Data map](docs/dev/data-map.md), [NFR](docs/dev/nfr.md) | Data classes and non-functional requirements. Fill them for your project |
| [Operations](docs/dev/ops.md), [Open questions](docs/dev/questions.md) | Runbook and open decisions |
| [Backend standard](docs/dev/backend-standard/README.md) | Rules for the Go core |
| [Frontend standard](docs/dev/frontend-standard/README.md) | Rules for the web apps |
| [UX standard](docs/dev/ux-standard/README.md) | Rules for each change that a person sees |
| [AI standard](docs/dev/ai-standard/README.md) | How agents work: cycle, hooks, review, delegation, writing style |
| [ADRs](docs/dev/adr/) | Architecture decisions. Template: [TEMPLATE.md](docs/dev/adr/TEMPLATE.md) |
| [Learning](docs/learning/README.md) | Lessons: one real problem of this foundation in each, with the theory, the options and the decision. Study one with `/lesson` |
