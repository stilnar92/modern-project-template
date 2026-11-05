# 03 — Review and verification

**Status:** DRAFT
**Decisions:** A4, A7, A8

## For each task

| Step | Who | What it checks |
| --- | --- | --- |
| 1. Automation | CI and hooks | Tests, linters, types, generated code against the contract, module boundary rules |
| 2. Scenarios | The agent that did the work | Each spec scenario passes; the PR shows the "scenario → test" links; E2E in a browser for public pages and the dashboard |
| 3. AI review in a fresh context | Another Claude agent, `code-review` | The change against the proposal and the deltas; the checklists below. Work done without a spec is a finding, the same as unfinished work |
| 4. External expert | ChatGPT through the Codex plugin in Claude Code | An independent view of another model on the same branch: bugs, security, what Claude missed. For important tasks, also a review that challenges the approach. The agent applies the comments or explains in the PR why not |
| 5. Human | You | The merge decision |

**How the agent handles the ChatGPT review**

1. The agent starts the review itself, in the task branch, through the helper script of the `codex@openai-codex` plugin:
   ```sh
   C=~/.claude/plugins/cache/openai-codex/codex/<version>/scripts/codex-companion.mjs
   node $C review --wait --base main --scope branch                 # normal review
   node $C adversarial-review --wait --base main --scope branch     # challenges the approach: for important tasks
   ```
   The result comes to the session at once; there is no wait for GitHub. If Codex code review on GitHub is on, the agent also reads its PR comments: `gh api repos/<repo>/pulls/<N>/reviews` and `.../pulls/<N>/comments`.
2. The agent posts the review result as a PR comment, so the owner sees the opinion of both models.
3. The agent evaluates each comment and does not apply it blindly:
   - **correct:** fixes it in the same branch and replies in the thread "fixed in <commit>";
   - **disputed or wrong:** does not change the code; replies in the thread with the reason and a link to the standard or ADR;
   - **needs the owner's decision** (data schema, security, money): marks it for the owner, by the [rules for expensive decisions](05-human-in-the-loop.md#decisions-cheap-and-expensive).
4. The agent resolves the threads, runs the checks again, writes the summary in the PR description ("ChatGPT review: N comments, M fixed, K rejected with reasons") and moves the task to Human review.

**AI review checklist**

- Account isolation: the server finds the account; the query has an `account_id` filter; a new table has an RLS rule; access goes through the entry point; the system role is only in allowed places; there is an isolation test ([backend standard 05](../backend-standard/05-security.md), [ADR 0009](../adr/0009-account-isolation-rls.md)).
- Personal data: not in logs and not in sinks; a new storage place is in `data-map.md`.
- Migrations are backward compatible ([backend standard 03](../backend-standard/03-data.md#migrations)).
- Contract: changes only add things; errors follow RFC 9457.
- New event field: there is a decision whether it goes to the analytics sink ([ADR 0003](../adr/0003-analytics-outbox-and-sinks.md)).
- Documents are updated together with the code.
- All project text is in English and follows ASD-STE100 ([06](06-writing-style.md)).

For large diffs from agents, you can use [reslop](https://github.com/tshemsedinov/reslop): review, a fix plan, handoff to an agent.

## Quality gates

A quality gate is an automatic check in the agent loop: the agent does the work, runs the check, reads the error and fixes it. A gate is for the agent first, not for a person. If a tool can check a rule, the rule does not live only in text ([02](02-harness.md#hooks), A10). Some of the tools and rules here come from a podcast on quality gates by Kirill Makevnin, which we reviewed on 03.11.2026.

| Gate | What it checks | After an edit | Before a commit | Local command | CI job |
| --- | --- | --- | --- | --- | --- |
| gofmt, go vet | Go format, likely Go bugs | Yes | Yes | `task lint`, `task test` | core |
| oxlint | TypeScript bugs, React hooks rules, accessibility, import boundaries | Yes | Yes | `task lint` | web |
| oxfmt | TypeScript format | Yes | Yes | `task lint` | web |
| TypeScript 7 (`tsc`), `astro check` | TypeScript types | — | — | `task lint` | web |
| golangci-lint | Go bugs, errors only | — | — | `task lint` | core |
| archcheck | Go import rules, table ownership, coupling report | — | — | `task lint` | core |
| `gen:check` | The generated code matches `api/openapi.yaml` | — | — | `task gen:check` | core, web |
| gitleaks | Secrets | — | Yes: staged, changed and new files | `task lint:secrets`: all history | repo (always) |
| govulncheck | Known vulnerabilities that the Go code calls | — | — | `task audit:go` (network) | core |
| osv-scanner | Known vulnerabilities in the web lockfile. Accepted findings with a reason and an end date in `osv-scanner.toml` | — | — | `task audit:deps` (network) | repo (always) |
| RLS check | Each table with `account_id`: RLS enabled and forced, each rule filters by the account. Each view has `security_invoker` | — | — | `task test` | core |
| lychee | Links and headings in Markdown, offline | — | — | `task lint:docs` | repo (always) |
| `lint:versions` | `mise.toml` matches `go.mod`, `core/Dockerfile` and `packageManager`. The test Postgres matches compose | — | — | `task lint` | repo (always) |
| Go and web tests | Behavior. Go tests use a real Postgres. API errors are Problem Details without internal text | — | — | `task test` | core, web |
| Migration round trip | Each migration goes down and up again, and the schema stays the same | — | — | `task test` | core |
| Hook tests, `test:mirrors` | The agent hooks. The Codex copies match the Claude files | — | — | `task test:hooks` | hooks |
| Build | All web apps build | — | — | `task build:web` | web |
| Healthchecks | Each service of the local stack is healthy before `task up` ends | — | — | `task up` | Not yet |
| Agent permissions | Claude Code cannot read `.env` files with real values or `~/.ssh` | Yes, on each read | — | — | — |
| No merge by a subagent | Only the orchestrator merges | Yes, on each command | — | `task test:hooks` | hooks |
| End-to-end (Playwright) | The API and both apps in Chromium and WebKit | — | — | `task test:e2e` after `task up` | Not yet |
| Dependabot | Vulnerable and old dependencies | — | — | — | GitHub. Security updates are on |

**Rules**

- **Fast checks first.** After an edit and before a commit, only checks that take seconds run (D11 of the frontend standard). Slow checks run in `task lint`, `task test` and CI.
- **Errors only.** Each finding fails the gate. There is no level of warnings that you can merge (D12).
- **A new gate starts at zero findings.** Fix the findings, or suppress one in the code with a reason (D12). If there are too many to fix at once, commit a baseline file with the old findings. Then CI fails only on new findings, and the baseline only gets smaller.
- **A slow gate goes to a schedule, not out of the project.** If a check makes the agent loop slow, run it in a nightly job. There is no nightly job yet. The online link check is the first candidate.
- **Each gate fails once on purpose.** Make a deliberate violation and see the gate fail before you trust it. Each PR that added a gate shows this check.
- **One tool for one task.** Before you add a tool, check that no gate already does the job. The candidates are in [04](04-tools.md).

## Periodically

| When | What | With |
| --- | --- | --- |
| Once a month, when there is code | A codebase audit with fix plans | [shadcn/improve](https://github.com/shadcn/improve): a strong model finds and plans, a cheap model executes in a worktree |
| A new module, a large boundary change | Coupling balance | `/modularity:review` or `im-validate-*` ([backend standard 01](../backend-standard/01-modules.md#how-it-is-checked)) |
| Before the pilot and before payments | Security audit | [cloudflare/security-audit-skill](https://github.com/cloudflare/security-audit-skill) in an isolated sandbox |
| Before a release of public pages | Core Web Vitals, WCAG 2.2 accessibility, SEO | [addyosmani/web-quality-skills](https://github.com/addyosmani/web-quality-skills) |
| After each change to `CLAUDE.md`, hooks, skills or the model | Agent workflow evals: does the change make agents better or worse | A set of 5–10 real tasks with automatic graders (planned: issue #37) |
| Once a week | Small agent frictions | The [papercuts](https://github.com/treygoff24/papercuts) log → edits to `CLAUDE.md`, hooks, skills |
