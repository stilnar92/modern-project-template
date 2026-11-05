# 04 — Tools

**Status:** DRAFT
**Decisions:** A11

External skills, plugins and utilities for development with agents. Checked on 28.10.2026. Before you install one, read its code and check its license and last update date (A11). The statuses are like a tech radar: **adopt** — we use it; **trial** — on a real task; **assess** — study it; **hold** — the reason is given.

| Tool | What it does | Status | Where we use it | Note |
| --- | --- | --- | --- | --- |
| [OpenSpec](https://openspec.dev/) | Spec format: capabilities and changes | Adopt | Each task | [ADR 0002](../adr/0002-specs-openspec-hybrid.md) |
| Matt Pocock's skills | Clarification, specs, TDD, review | Adopt | Each task | Spec templates adapted to OpenSpec |
| Plugin `codex@openai-codex` for Claude Code | Branch review by the ChatGPT model, a review that challenges the approach, handing a task to Codex | Adopt | Each PR, the "AI review" step | Needs the Codex CLI (`npm i -g @openai/codex`) and a ChatGPT login; the agent calls the plugin script itself |
| Codex code review on GitHub | ChatGPT comments directly in the PR | Assess | If we want to see the comments in the PR | Turn it on in the Codex settings in ChatGPT; needs access to the private repository |
| [Playwright MCP](https://github.com/microsoft/playwright-mcp) | The agent controls a browser | Adopt | E2E checks of public pages and the dashboard | — |
| [cloudflare/security-audit-skill](https://github.com/cloudflare/security-audit-skill) | A multi-phase security audit; separate agents check the findings | Trial | Before the pilot and payments | Needs an isolated sandbox |
| [addyosmani/web-quality-skills](https://github.com/addyosmani/web-quality-skills) | Audit of Core Web Vitals, accessibility, SEO | Trial | Public page and landing page templates | Matches D13, 22, 30 of the frontend standard |
| [shadcn/improve](https://github.com/shadcn/improve) | A code audit and plans for a cheaper model | Trial | Once a month, when there is code | — |
| [papercuts](https://github.com/treygoff24/papercuts) | Agents record small frictions | Trial | Improving `CLAUDE.md` and skills | — |
| [vladikk/modularity](https://github.com/vladikk/modularity) | Review and design of modules by the balanced coupling model | Assess | Core module boundaries | ⚠️ CC BY-NC-SA: for a commercial project, ask the author |
| [reslop](https://github.com/tshemsedinov/reslop) | A CLI to review agent diffs and plan fixes | Assess | Large diffs from parallel agents | Has no code checks of its own yet: only review and test summaries |
| [Database Skills](https://database-skills.com/) (PlanetScale) | A skill for agent work with Postgres | Assess | Schema and queries | We did not check the content of the postgres skill |
| [Stagehand](https://github.com/browserbase/stagehand) | AI browser automation | Assess | Stable E2E tests | Already in [29 of the frontend standard](../frontend-standard/29-testing-strategy.md); assertions stay deterministic |
| [Archiplan](https://github.com/archiplan-ai/Archiplan) | Specs with "stress rounds" | Hold | — | Beta, few users; we already selected OpenSpec |
| [mdtask](https://github.com/syabro/mdtask) | Specs and tasks in the repository | Hold | — | An alternative to OpenSpec |
| Superset, Orca, Herdr | Run many agents in parallel | Hold | — | The process limits work to 2–3 streams; worktrees are sufficient |

**Quality gates and the toolchain.** Checked on 03.11.2026, after a podcast on quality gates by Kirill Makevnin. Where each gate runs: [03, Quality gates](03-review-verification.md#quality-gates).

| Tool | What it does | Status | Where we use it | Note |
| --- | --- | --- | --- | --- |
| [mise](https://mise.jdx.dev) | One file with the versions of all tools: Go, Node, pnpm, Task, linters | Adopt | `mise.toml`, local work and CI (`jdx/mise-action`) | Only `[tools]`, so a new worktree needs no `mise trust`. Taskfile and Docker Compose stay. Dependabot does not update `mise.toml`: change versions by hand |
| [gitleaks](https://github.com/gitleaks/gitleaks) | Finds secrets in files and in the git history | Adopt | `task lint:secrets`, the commit hook, CI | No allowlist: the history has no findings |
| [lychee](https://github.com/lycheeverse/lychee) | Checks links and headings in Markdown | Adopt | `task lint:docs`, CI | Offline only. It found 12 broken links on the first run |
| heavy | A lock for heavy commands (tests, builds) from parallel worktrees, with slots for each type of task | Assess | When 2–3 streams block each other on ports, Postgres or CPU | A Rust CLI from the podcast. Find its repository and check it by A11 before use |
| [oasdiff](https://github.com/oasdiff/oasdiff) | Finds breaking changes between two versions of an OpenAPI contract | Assess | CI, when `/v1` has real endpoints | A tool for the rule "inside `/v1`, changes only add things" (ADR 0005). Now only the review checks it |
| A transaction for each test | Each test runs in a transaction and rolls it back, instead of a new database for each test | Assess | `internal/platform/testdb`, when the tests get slow | The podcast measured about 3 ms for each test, against 50–80 ms with database templates or containers |
| A database-against-code check | Compares database constraints with the code: unique keys, foreign keys, nullable columns | Assess | When domain tables exist | Like `database_consistency` in Ruby. The backend standard selects sqlc and no ORM models, so first check what such a tool can compare |
| [TypeSpec](https://typespec.io) | A short language that generates OpenAPI and clients | Hold | — | OpenAPI and oapi-codegen work now. A change adds risk and gives no gain yet |
| A SCIP-based architecture linter | Layer and dependency rules for many languages, with type data | Hold | — | `core/tools/archcheck` covers Go. oxlint covers the web boundaries |
