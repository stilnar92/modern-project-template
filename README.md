# modern-project-template

A template for a new product: a Go core, two web apps, Postgres with row-level security, quality gates, CI and a harness for AI agents. It contains the foundation only. There is no business domain in it.

## What is inside

| Part | Where | What it is |
| --- | --- | --- |
| Core | `core/` | Go: `cmd/api`, `cmd/worker`, `cmd/migrate`. Shared infrastructure in `internal/platform/` (config, database, health, observability, Problem Details errors, test database) |
| Database | `core/migrations/` | Postgres 16. Three roles (`app_owner`, `app_app`, `app_system`) and account isolation with RLS ([ADR 0009](docs/dev/adr/0009-account-isolation-rls.md)) |
| Contract | `api/openapi.yaml` | OpenAPI 3.1. `task gen` generates the Go server and the TypeScript client from it |
| Dashboard | `web/apps/dashboard` | Vite, React, Mantine, TanStack Query ([ADR 0008](docs/dev/adr/0008-dashboard-mantine.md)) |
| Site | `web/apps/site` | Astro with server rendering, no JavaScript in the browser |
| Shared web code | `web/packages/` | `@app/api-client`: the generated API client |
| End-to-end tests | `web/e2e` | Playwright |
| Local environment | `deploy/` | Docker Compose: Postgres, S3 (SeaweedFS), Mailpit, Caddy and the core ([deploy/README.md](deploy/README.md)) |
| Specs | `openspec/` | OpenSpec: current capability specs and change proposals ([ADR 0002](docs/dev/adr/0002-specs-openspec-hybrid.md)) |
| Agent harness | `CLAUDE.md`, `.claude/`, `AGENTS.md`, `.codex/`, `.agents/` | Rules, hooks and skills for Claude Code and Codex ([AI standard](docs/dev/ai-standard/README.md)) |
| CI | `.github/` | Linters, architecture checks, tests and generated-code checks. Issue and PR templates, Dependabot |
| Tools | `mise.toml`, `Taskfile.yml` | Pinned tool versions and the shared commands |

## Start a new project

1. On GitHub, click **Use this template** and create your repository. Clone it.
2. Install [mise](https://mise.jdx.dev) and Docker Desktop. Run `mise install` in the repository root. mise installs Go, Node, pnpm, Task and the linters with the versions from `mise.toml`.
3. Rename the Go module. Change the `module` line in `core/go.mod` and replace `github.com/stilnar92/modern-project-template/core` in all Go imports. Replace the repository URL in `.github/pull_request_template.md`. Then run `task build:core`.
4. Rename the npm scope `@app/` in `web/**/package.json`, in the imports, and in the oxlint boundary rules. Then run `pnpm install` in `web/`.
5. Optional: rename the compose project and the database. `COMPOSE_PROJECT_NAME` is in `.env.example`. The database `app` and the roles `app_owner`, `app_app`, `app_system` are in `deploy/postgres/init/01-roles.sh`, `deploy/compose.yaml`, `.env.example`, `Taskfile.yml` (`db:dump`, `db:restore`) and the migrations.
6. Describe your product. Fill [docs/dev/brief.md](docs/dev/brief.md), the domain glossary [CONTEXT.md](CONTEXT.md), [docs/dev/nfr.md](docs/dev/nfr.md) and [docs/dev/data-map.md](docs/dev/data-map.md). Replace the first lines of `CLAUDE.md` with your product description, then copy it to `AGENTS.md`.
7. Run `task up`, `task lint` and `task test`. All must pass before your first feature.

## Commands

| Command | What it does |
| --- | --- |
| `mise install` | Install the tool versions from `mise.toml` |
| `task --list` | Show all commands |
| `task up` / `task down` | Start or stop the local environment. `task up` creates `.env`, applies the migrations and starts the core on http://api.localhost:8088 |
| `task lint` | Version, secret, link, Go and web checks. CI runs the same |
| `task test` | Go tests (Postgres in Docker), web tests and agent hook tests |
| `task gen` | Generate the Go server and the TypeScript client from `api/openapi.yaml` |
| `task test:e2e` | Playwright tests. Run `task up` first |
| `task mock` | A mock API from the contract on http://127.0.0.1:4010 (Prism) |
| `task run:api` | Start the core API on the host, on http://localhost:8080 |

`pnpm dev` in `web/` starts the dashboard on http://localhost:5173 and the site on http://localhost:4330.

## Layout

```
api/openapi.yaml   API contract
core/              Go core: cmd/, internal/, migrations/, tools/archcheck
web/               pnpm workspace: apps/, packages/, e2e/
deploy/            Docker Compose and Caddy for the local environment
openspec/          capability specs and changes
docs/dev/          development docs, standards, ADRs
docs/learning/     optional learning notes
Taskfile.yml       shared commands
mise.toml          tool versions
```

The layout is in [ADR 0007](docs/dev/adr/0007-monorepo-layout.md).

## Documents

| Document | What it contains |
| --- | --- |
| [Brief](docs/dev/brief.md) | The product on one page for agents. Fill it for your project |
| [Architecture](docs/dev/architecture.md) | C4 diagrams of the system |
| [Tech stack](docs/dev/tech-stack.md) | Technologies and the reasons for them |
| [Development process](docs/dev/dev-process.md) | The task cycle and how agents work |
| [Data map](docs/dev/data-map.md), [NFR](docs/dev/nfr.md) | Data classes and non-functional requirements. Fill them for your project |
| [Operations](docs/dev/ops.md), [Open questions](docs/dev/questions.md) | Runbook and open decisions |
| [Backend standard](docs/dev/backend-standard/README.md) | Rules for the Go core |
| [Frontend standard](docs/dev/frontend-standard/README.md) | Rules for the web apps |
| [AI standard](docs/dev/ai-standard/README.md) | How agents work: cycle, hooks, review, writing style |
| [ADRs](docs/dev/adr/) | Architecture decisions. Template: [TEMPLATE.md](docs/dev/adr/TEMPLATE.md) |
| [Learning](docs/learning/README.md) | Optional: debriefs and a concept map for a team that learns the stack |
