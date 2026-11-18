# 03 — Review and verification

**Status:** DRAFT
**Decisions:** A4, A7, A8

## For each task

| Step | Who | What it checks |
| --- | --- | --- |
| 1. Automation | CI and hooks | Tests, linters, types, generated code against the contract, module boundary rules |
| 2. Scenarios | The agent that did the work | Each spec scenario passes; the PR shows the "scenario → test" links (`mise run check:scenarios` fails without them); E2E in a browser for public pages and the dashboard |
| 3. AI review in a fresh context | The `architecture-reviewer` agent (`.claude/agents/`) and `code-review` | The change against the proposal and the deltas; the checklists below. Work done without a spec is a finding, the same as unfinished work |
| 4. External expert | ChatGPT through the Codex plugin in Claude Code | An independent view of another model on the same branch: bugs, security, what Claude missed. For important tasks, also a review that challenges the approach. The agent applies the comments or explains in the PR why not |
| 5. Human | You | The merge decision |

### How the agent handles the ChatGPT review

1. The agent starts the review itself, in the task branch:

   ```sh
   mise run review:codex                                         # normal review
   mise run review:codex adversarial-review "<what to challenge>"  # challenges the approach: for important tasks
   ```

   The task calls the helper script of the `codex@openai-codex` plugin with `--base origin/main --scope branch`. The result comes to the session at once; there is no wait for GitHub. If Codex code review on GitHub is on, the agent also reads its PR comments: `gh api repos/<repo>/pulls/<N>/reviews` and `.../pulls/<N>/comments`.
2. When the branch has a PR, the task posts the result as a PR comment, so the owner sees the opinion of both models. Without a PR, the agent posts it after it opens the PR.
3. The agent evaluates each comment by the [rules below](#how-to-handle-a-finding) and does not apply it blindly:
   - **correct:** fixes it in the same branch and replies in the thread "fixed in <commit>";
   - **disputed or wrong:** does not change the code; replies in the thread with the reason and a link to the standard or ADR;
   - **needs the owner's decision** (data schema, security, money): marks it for the owner, by the [rules for expensive decisions](05-human-in-the-loop.md#decisions-cheap-and-expensive).
4. The agent resolves the threads, runs the checks again, writes the summary in the PR description ("ChatGPT review: N comments, M fixed, K rejected with reasons") and moves the task to Human review.

### How to handle a finding

These rules apply to each finding: from a review agent, ChatGPT, a subagent's report or a deviation journal ([07](07-delegation.md#acceptance)). Source: `references/review-acceptance.md` of [pragmatic-orchestration](https://github.com/CodeAlive-AI/pragmatic-orchestration).

Decide by the evidence and by the effect on the task. A severity label, the confidence of the reviewer and the agreement of two models are inputs, not proof.

| Decision | When | What the agent does |
| --- | --- | --- |
| **Fix** | A confirmed violation of the spec or a real risk in normal use | Fixes it in the branch |
| **Verify** | The concern is possible and can change the decision, but there is not enough evidence | One targeted check: a test, a trace of the code path. Not a rewrite "just in case" |
| **Accept or defer** | A real limitation that the task can accept | Writes the reason, the effect and when to come back to it in the PR |
| **Dismiss** | Proved wrong, a duplicate, a guess without evidence, a matter of taste | Replies in the thread with the reason |
| **Owner** | Needs a scope change or an expensive decision | A decision card ([05](05-human-in-the-loop.md#decision-card)) |

- A trace of a concrete code path can prove a defect without a reproduction. No reproduction alone is not a reason to dismiss.
- An expensive fix does not make a blocker acceptable. An MVP does not excuse a data leak between accounts, a data loss or a broken required behavior.
- Start another review round only for a concrete open risk, new evidence or a required check. Optional suggestions are not a reason.
- The work is done when the criteria and the checks pass and no blocker stays. Zero comments and full agreement of the reviewers are not necessary.
- If the fixes repeat without progress, change the approach or report the blocker. A round limit does not give permission to accept a defect.

### AI review checklist

- Account isolation: the server finds the account; the query has an `account_id` filter; a new table has an RLS rule; access goes through the entry point; the system role is only in allowed places; there is an isolation test ([backend standard 05](../backend-standard/05-security.md), [ADR 0009](../adr/0009-account-isolation-rls.md)).
- Personal data: not in logs and not in sinks; a new storage place is in `data-map.md`.
- Migrations are backward compatible ([backend standard 03](../backend-standard/03-data.md#migrations)).
- Contract: changes only add things; errors follow RFC 9457.
- New event field: there is a decision whether it goes to the analytics sink ([ADR 0003](../adr/0003-analytics-outbox-and-sinks.md)).
- UI change: it follows the [UX standard](../ux-standard/README.md). The spec lists the states, the code shows them, and the PR says which browser checks were done.
- New screen or flow: the proposal has `Design: approved`, written by the owner. Each state matches its screenshot in `openspec/changes/<id>/design/`. The "Screens and contract" table has no MISSING field, and the contract has each field from it.
- Documents are updated together with the code.
- All project text is in English and follows ASD-STE100 ([06](06-writing-style.md)).

For large diffs from agents, you can use [reslop](https://github.com/tshemsedinov/reslop): review, a fix plan, handoff to an agent.

## Quality gates

A quality gate is an automatic check in the agent loop: the agent does the work, runs the check, reads the error and fixes it. A gate is for the agent first, not for a person. If a tool can check a rule, the rule does not live only in text ([02](02-harness.md#hooks), A10). Some of the tools and rules here come from a podcast on quality gates by Kirill Makevnin, which we reviewed on 03.11.2026.

| Gate | What it checks | After an edit | Before a commit | Local command | CI job |
| --- | --- | --- | --- | --- | --- |
| gofmt, go vet | Go format, likely Go bugs | Yes | Yes | `mise run lint`, `mise run test` | core |
| oxlint | TypeScript bugs, React hooks rules, accessibility, import boundaries | Yes | Yes | `mise run lint` | web |
| oxfmt | TypeScript format | Yes | Yes | `mise run lint` | web |
| TypeScript 7 (`tsc`), `astro check` | TypeScript types | — | — | `mise run lint` | web |
| golangci-lint | Go bugs, errors only | — | — | `mise run lint` | core |
| archcheck | Go import rules, table ownership, coupling report | — | — | `mise run lint` | core |
| `gen:check` | The generated code matches `api/openapi.yaml` | — | — | `mise run gen:check` | core, web |
| gitleaks | Secrets | — | Yes: staged, changed and new files | `mise run lint:secrets`: all history | repo (always) |
| govulncheck | Known vulnerabilities that the Go code calls | — | — | `mise run audit:go` (network) | core, and weekly (`audit.yml`) |
| osv-scanner | Known vulnerabilities in the web and docs viewer lockfiles. Accepted findings with a reason and an end date in `osv-scanner.toml` | — | — | `mise run audit:deps` (network) | repo (always), and weekly (`audit.yml`) |
| RLS check | Each table with `account_id`: RLS enabled and forced, each rule filters by the account. Each view has `security_invoker` | — | — | `mise run test` | core |
| lychee | Links and headings in Markdown, offline | — | — | `mise run lint:docs` | repo (always) |
| `docs:check` | The docs viewer build: each page, internal link, anchor, image and title in the HTML ([ADR 0016](../adr/0016-docs-viewer.md)) | — | — | `mise run docs:check` | docs (when a document or the viewer changes) |
| rumdl | Markdown style. Config in `.rumdl.toml` | — | — | `mise run lint:docs` | repo (always) |
| typos | Spelling in code and docs. Real names in `_typos.toml` | — | — | `mise run lint:docs` | repo (always) |
| jactionlint | Workflows and `dependabot.yml`: actions pinned by SHA, job timeouts, Dependabot cooldown. Warnings fail too | — | — | `mise run lint:ci` | repo (always) |
| agnix | `CLAUDE.md`, `AGENTS.md`, skills and hook configs. Warnings fail too | — | — | `mise run lint:agents` | repo (always) |
| hadolint, shellcheck | Dockerfiles and shell scripts | — | — | `mise run lint:docker`, `mise run lint:shell` | repo (always) |
| vacuum, oasdiff | `api/openapi.yaml`: style, and no breaking change against `main` (backend rule B4) | — | — | `mise run lint:api` | repo (always) |
| squawk | Migrations that lock tables or have no lock and statement timeouts | — | — | `mise run lint:core` | core |
| Knip | Unused files, exports and dependencies in `web/` | — | — | `mise run lint:web` | web |
| `lint:versions` | `mise.toml` matches `go.mod`, `core/Dockerfile` and `packageManager` in `web/` and `tools/docs-viewer/`. The test Postgres matches compose | — | — | `mise run lint` | repo (always) |
| Go and web tests | Behavior. Go tests use a real Postgres. API errors are Problem Details without internal text | — | — | `mise run test` | core, web |
| Migration round trip | Each migration goes down and up again, and the schema stays the same | — | — | `mise run test` | core |
| Scenario links | Each spec scenario of a `feat/` branch has a test or check in the "Scenarios → checks" table of the PR description | — | — | `mise run check:scenarios` | scenarios (on each PR change, also an edit of the description) |
| Hook tests, `test:mirrors` | The agent hooks. The Codex copies match the Claude files | — | — | `mise run test:hooks` | hooks |
| Build | All web apps build | — | — | `mise run build:web` | web |
| Healthchecks, bootstrap | Each service of the local stack is healthy before `mise run up` ends. `mise run setup` works on a clean checkout and `/healthz` answers | — | — | `mise run setup` | e2e |
| Agent permissions | Claude Code cannot read `.env` files with real values or `~/.ssh` | Yes, on each read | — | — | — |
| No merge by a subagent | Only the orchestrator merges | Yes, on each command | — | `mise run test:hooks` | hooks |
| End-to-end (Playwright) | The API and both apps in Chromium and WebKit | — | — | `mise run test:e2e` after `mise run setup` | e2e |
| Dependabot | Vulnerable and old dependencies. Each update waits 7 days after a release | — | — | — | GitHub. Security updates are on |

### Rules

- **Fast checks first.** After an edit and before a commit, only checks that take seconds run (D11 of the frontend standard). Slow checks run in `mise run lint`, `mise run test` and CI.
- **Errors only.** Each finding fails the gate. There is no level of warnings that you can merge (D12).
- **A new gate starts at zero findings.** Fix the findings, or suppress one in the code with a reason (D12). If there are too many to fix at once, commit a baseline file with the old findings. Then CI fails only on new findings, and the baseline only gets smaller.
- **A slow gate goes to a schedule, not out of the project.** If a check makes the agent loop slow, run it in the nightly job (`.github/workflows/nightly.yml`). It already runs `go test -race` and the online link check.
- **Each gate fails once on purpose.** Make a deliberate violation and see the gate fail before you trust it. Each PR that added a gate shows this check.
- **One tool for one task.** Before you add a tool, check that no gate already does the job. The candidates are in [04](04-tools.md).

## Periodically

| When | What | With |
| --- | --- | --- |
| Once a month, when there is code | A codebase audit with fix plans | [shadcn/improve](https://github.com/shadcn/improve): a strong model finds and plans, a cheap model executes in a worktree |
| A new module, a large boundary change | Coupling balance | `/modularity:review` or `im-validate-*` ([backend standard 01](../backend-standard/01-modules.md#how-it-is-checked)) |
| Before the pilot and before payments | Security audit | [cloudflare/security-audit-skill](https://github.com/cloudflare/security-audit-skill) in an isolated sandbox |
| Before a release of public pages | Core Web Vitals, WCAG 2.2 accessibility, SEO | [addyosmani/web-quality-skills](https://github.com/addyosmani/web-quality-skills) |
| After each change to `CLAUDE.md`, hooks, skills or the model | Agent workflow evals: does the change make agents better or worse | A set of 5–10 real tasks with automatic graders (planned) |
| Once a week | Small agent frictions | The [papercuts](https://github.com/treygoff24/papercuts) log → edits to `CLAUDE.md`, hooks, skills |
