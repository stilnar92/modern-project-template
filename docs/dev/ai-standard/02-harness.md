# 02 — Claude Code setup

**Status:** COLLECTING — the points are collected; we refine the rules when we set up the repository.
**Decisions:** A5, A6, A10

## `CLAUDE.md`

- Layers by folder ([development process](../dev-process.md#documentation-in-layers)): the root has rules for everybody; `/core` is for Go; `/web` is for the frontend. An agent in a folder sees only its layers.
- `CLAUDE.md` contains what each task needs: build and check commands, the main prohibitions, links to the standards. The details stay in the standards, through a link, not as a copy.
- Add a line to `CLAUDE.md` when an agent makes the same mistake twice. Remove lines that are not necessary.

## Hooks

Source: [Hook-driven dev workflows](https://nick-tune.me/blog/2026-02-28-hook-driven-dev-workflows-with-claude-code/) (Nick Tune). The idea: the process is a state machine with its state in a file. Hooks stop the agent from doing what the current phase forbids. If a hook can check a rule, the rule must not live only in text.

| Hook | Event | What it does |
| --- | --- | --- |
| Shared files for the orchestrator only | `PreToolUse` on an edit | A subagent cannot change `openapi.yaml`, migrations, `events.md`, `CONTEXT.md`, `openspec/specs/` |
| No code without an approved plan | `PreToolUse` on a code edit | Blocks the edit if the current OpenSpec change has no approval mark |
| No commit without checks | `PreToolUse` on `git commit` | Runs the fast checks on each file that the commit can contain, and gitleaks for secrets. Blocks the commit if a check fails |
| No push to `main` | `PreToolUse` on `git push` | The agent does not push to `main`: the free plan has no branch protection |
| No merge by a subagent | `PreToolUse` on `gh pr merge` and merge calls of `gh api` | A subagent reports its result. Only the orchestrator merges, after the review |
| Fast checks after an edit | `PostToolUse` | `gofmt` and `go vet` for Go, `oxlint` and `oxfmt --check` for TS: only fast checks (D11 of the frontend standard) |
| Task state into the context | `SessionStart`, `SubagentStart` | Loads the phase and the phase instructions, not the whole process |

One entry for all hooks: [`.claude/hooks/guard.mjs`](../../../.claude/hooks/guard.mjs), registered in `.claude/settings.json`. Each new agent failure is a reason to add a check (A10). Add a case to `.claude/hooks/guard.test.mjs` for each new check: `task test:hooks` runs them, and CI too.

**Codex copies.** The Codex CLI reads its own paths: `.codex/hooks/` (registered in `.codex/hooks.json`), `.agents/skills/`, and `AGENTS.md` in the root, `core/` and `web/`. These are copies of the Claude files. Change the Claude file first, then copy it. `task test:mirrors` (part of `task test:hooks`) fails when a copy differs. `.codex/hooks.json` finds the hook through `git rev-parse --show-toplevel`, so it works on each machine and in each worktree (not checked in a live Codex session yet).

**In place now:** shared files, the approved plan, no push to `main`, no merge by a subagent, fast checks after an edit, no commit without checks. **Not yet:** task state into the context.

- **Orchestrator or subagent:** a tool call from a subagent has the `agent_id` field. A call from the main agent does not.
- **Approved plan:** the branch name gives the plan. `feat/<change-id>` needs `openspec/changes/<change-id>/proposal.md` with the line `Status: approved`. Only the owner writes this line. `chore/`, `fix/` and `docs/` branches need no proposal. Code edits on `main` are blocked.
- **Commit rule:** the hook runs before the command, so `git add` in the same call has not run yet. That is why it checks the staged, the changed and the new files of the worktree, not only the staged ones. A file that you do not commit can block the commit too. Tools come from `mise.toml` through `mise which`, so the shell does not need to activate mise. If a tool is not installed, its check is skipped.
- **Time limit:** a `PreToolUse` hook that passes its time limit does not block the tool call. So the commit rule has its own budget of 45 s, below the hook limit of 60 s, and a check that does not finish in time blocks the commit.
- **Limit:** a hook reads the command line. It stops honest mistakes, but it is not a security boundary: an agent can write a script and run it. The push rule is an allow list (only `git push <remote> <branch>` with explicit names that are not `main`) to keep the gaps small. Full protection of `main` needs a server rule: branch protection on a paid plan, or a public repository.
- `CLAUDE.md` and `.claude/` are often in a global gitignore. The repository `.gitignore` allows them, so they are committed.

## Skills

Installed and in use: Matt Pocock's skills (`grilling`, `tdd`, `diagnosing-bugs`, `prototype`, `code-review`, `codebase-design`, `domain-modeling`), `im-decompose-*`, `im-trace-*`, `im-validate-*`, `find-solution`, `webapp-testing`, `visual-plan`. Candidates to install are in [04](04-tools.md).

Rewrite the `to-spec` and `to-tickets` templates for OpenSpec files ([ADR 0002](../adr/0002-specs-openspec-hybrid.md)).

## GitHub task board

The board is a GitHub Project with the number `<PROJECT_NUMBER>` and the ID `<PROJECT_ID>`. It is linked to the `<OWNER>/<REPO>` repository. The agent creates issues with `gh`. The token needs the `project` scope.

> **Set up for a new project:** create the GitHub Project and link it to the repository. Then replace `<OWNER>`, `<REPO>`, `<PROJECT_NUMBER>` and `<PROJECT_ID>` in this section. `gh project list --owner <OWNER>` shows the number. `gh project view <PROJECT_NUMBER> --owner <OWNER> --format json --jq .id` shows the ID.

**Rule:** the truth about a task is in the OpenSpec spec. The issue on the board gives a short summary and links to the change ([05 of the frontend standard](../frontend-standard/05-spec-driven-development.md#3-repository-layout-openspec), [ADR 0002](../adr/0002-specs-openspec-hybrid.md)).

| Field | Values |
| --- | --- |
| Status | Backlog → Ready → In progress → AI review → Human review → Done |
| Area | core-backend, core-frontend, product-module, infra, docs, learning |
| Slice | The number of the slice in `tasks.md` of the change |
| Size | S, M, L |

Labels: `feature`, `bug`, `spike`, `chore`, `documentation`, `learning` and **`needs-decision`**: the issue waits for an expensive decision of the owner ([05](05-human-in-the-loop.md#decision-card)). The issue and PR templates are in `.github/`.

**When the agent works with the board**

| Moment | What the agent does |
| --- | --- |
| The proposal is approved | `to-tickets` creates one issue for each slice from `tasks.md` (template "Task from a proposal") and adds it to the board with Status = Ready and with Area, Slice and Size filled in |
| An expensive decision is necessary | An issue from the "Decision" template with the `needs-decision` label |
| A technical question has no answer | An issue from the "Spike" template; dependent issues mention it |
| Work starts | Status = In progress |
| The PR is open | Status = AI review; the PR contains `Closes #N`. Review by an agent in a fresh context and by an external expert, ChatGPT ([03](03-review-verification.md)) |
| The AI comments are handled | Status = Human review: the PR waits for the owner's review and merge |
| The PR is merged | GitHub closes the issue, and the board moves it to Done |

Set the board automations ("Item added" → Backlog, "Item closed" and "Pull request merged" → Done, "Auto-add to project" for repository issues) and the "Board" view manually in the project settings: the API cannot create them. **After each change to the list of statuses,** select the automation values again: they point to the old options and silently stop working.

**Commands**

```sh
# create an issue and add it to the board
url=$(gh issue create -R <OWNER>/<REPO> --title "..." --body-file body.md --label feature)
gh project item-add <PROJECT_NUMBER> --owner <OWNER> --url "$url" --format json --jq .id   # board item ID

# field and option IDs: once per session
gh project field-list <PROJECT_NUMBER> --owner <OWNER> --format json

# set a field
gh project item-edit --project-id <PROJECT_ID> --id <item ID> \
  --field-id <field ID> --single-select-option-id <option ID>
```

## No push to main

On the free GitHub plan, a private repository cannot protect the `main` branch. So agents do not push to `main` and do not merge into it themselves: work goes into a branch, and the merge goes through a PR after the owner approves it. A `PreToolUse` hook on `git push` to `main` enforces the rule (see [hooks](#hooks)).

## MCP

| Server | Why |
| --- | --- |
| Playwright | The agent checks the public page and the dashboard in a browser |
| Context7 | Current library documentation |
| Figma | Dashboard and public page designs |

## Secrets

- The agent works only with local and test secrets: test accounts of external services and the test mode of the payment service.
- Production secrets do not go into the working copy. `permissions.deny` in `.claude/settings.json` stops Claude Code from reading `.env`, `.env.local`, `.env.production`, `.env.*.local` at any depth, and `~/.ssh`. The rule covers the Read tool and the file commands that Claude Code knows in Bash (`cat`, `head`, `tail`). `.env.example` stays readable. Limit: a script that opens the file itself is not stopped, so this is a guard against mistakes, not a security boundary. The Codex CLI has no such rule yet.
