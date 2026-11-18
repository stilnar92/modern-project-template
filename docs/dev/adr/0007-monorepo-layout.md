# 0007 — Monorepo: folders by stack

Date: 2026-10-28
Status: accepted. The Taskfile part is superseded by [0010](0010-mise-single-toolchain.md)

## Context

Everything is in one repository: the Go core, the TypeScript frontends, the contract, specs and documents. Go and Node have different tools, and each must work without the other. Agents need short, local rules for the folder they work in.

For now, development is only local. We have not selected a server, but the first deploy must be fast.

## Decision

**Folders by stack:**

```text
api/openapi.yaml        contract
openspec/               specs (ADR 0002)
core/                   Go: one go.mod
  cmd/api, cmd/worker
  internal/<module>/    core and product modules
  migrations/
web/                    TypeScript: pnpm workspace
  apps/dashboard, apps/site
  packages/api-client …
  e2e/                  Playwright tests
deploy/                 docker-compose and Caddy: a local environment like production
docs/  .github/  .claude/
Taskfile.yml            shared commands: up, gen, test, lint, migrate
```

- **Tools do not mix:** Go is in `core/` with its own `go.mod`; Node is in `web/` with its own pnpm workspace. Shared commands are in the root `Taskfile.yml`.
- **`CLAUDE.md` in layers:** root, `core/`, `web/`. An agent in a folder sees only the rules for that folder.
- **CI by path:** changes in `core/` do not start the `web/` checks, and the opposite. A change in `api/openapi.yaml` starts both.
- Read the paths from ADR 0001 (`web/dashboard`) as `web/apps/dashboard`.

## Consequences

- The layout matches the [development process](../dev-process.md) and the [tech stack](../tech-stack.md).
- The core and the frontends are at different depths (`core/` and `web/apps/...`). This is the cost of separate tools.
- A new app goes into `web/apps/`, and a new shared package into `web/packages/`.
