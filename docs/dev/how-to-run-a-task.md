# How to run a product task

12.11.2026 · @ilnar

A short guide for the owner: what to do, what to run and what to tell the agent for an epic, a task or a bug. The full rules are in the [development process](dev-process.md) and the [AI standard](ai-standard/README.md). If this guide and those documents differ, those documents win.

## Roles

| Who | What they do |
| --- | --- |
| The owner (you) | Decides what to build, approves the plan, makes expensive decisions, merges |
| The orchestrator (the main agent session) | Explores, writes the plan, the contract and the code, runs the checks, opens the PR |
| Hooks | Block code on `feat/` without an approved plan, block edits of shared files by subagents, block a push to `main` ([02](ai-standard/02-harness.md#hooks)) |

You do not write code. You make three decisions: **what** to build, **whether the plan is correct**, and **whether to merge**.

## 0. Start each session

1. Open a new worktree for each task. Do not run more than 2–3 tasks at the same time: the limit is your review capacity.
2. Run `mise run setup`. Continue only after it prints `baseline: green`.
3. Start Claude Code in the worktree.

## 1. Select the path

| What you have | Path | Branch |
| --- | --- | --- |
| **Epic:** a large capability, many screens or modules | Clarify once, then one full cycle for each part | One `feat/<change-id>` for each part |
| **Task:** one change of behavior | Full cycle | `feat/<change-id>` |
| **Bug:** the code does not do what the spec already says | Short path | `fix/<name>` |
| Tooling or documents | Short path | `chore/<name>` or `docs/<name>` |

## 2. Full cycle: epic or task

| Step | You do | You tell the agent | Result |
| --- | --- | --- | --- |
| 1. Clarify | Answer the questions | "Grill me on <idea>" (`grilling` skill) | Changes in `CONTEXT.md`, `docs/dev/THESAURUS.md` and ADRs |
| 2. Proposal | Read it and correct it | `/to-spec <change-id> <request>` | `openspec/changes/<change-id>/proposal.md` and GIVEN/WHEN/THEN scenarios |
| 2a. Expensive decision | Select an option on the [decision card](ai-standard/05-human-in-the-loop.md#decision-card) | The agent brings the card. For a hard choice: `/find-solution` | An ADR after your choice |
| 3. Design: only for a new screen or flow | Comment on the clickable prototype. When you agree, write `Design: approved` in the proposal | "Make a prototype of the recommended flow" | Screenshots in `design/` and the "Screens and contract" table without MISSING |
| 3a. Plan review | Read the findings | A fresh agent reviews the plan | Fixes in the proposal |
| **Gate** | **Write `Status: approved` in the proposal** | — | The hook now permits code on the branch |
| 4. Contract | — | The agent changes `api/openapi.yaml` and runs `mise run gen` | Generated Go server and TS client |
| 5. Slices | Check the order | `/to-tickets <change-id>` | `tasks.md` and issues on the [GitHub board](ai-standard/02-harness.md#github-task-board) |
| 6–7. Code | Look at `progress.md` if you want | "Implement slice 1". The first slice is a tracer: database → API → interface → E2E test | Code and tests, written with TDD |
| 8. Review | Read the PR summary | The agent runs the AI review ([03](ai-standard/03-review-verification.md)), `/code-review` and `mise run review:codex` | The review comments are fixed or answered |
| 9. Merge | **Merge** after CI is green | — | Deltas merged into `openspec/specs/`, the change archived, lessons in `docs/learning/lessons/` for the problems that had a real choice |

**Epic.** Do step 1 for the whole epic. Then write one proposal for each independent part. Start with the part that the other parts depend on.

## 3. Short path: bug

1. Tell the agent the symptom and the steps to reproduce it.
2. If the cause is not clear, the agent uses the `diagnosing-bugs` skill: it checks hypotheses one by one.
3. The agent writes a failing test first. For a bug, this test is the spec. Then it writes the fix.
4. The agent opens a PR from a `fix/` branch. Review and merge as in steps 8 and 9.
5. If the "bug" needs new behavior, it is a task. Go to the full cycle.

## 4. Commands

| Command | When |
| --- | --- |
| `mise run setup` | At the start of each session and worktree |
| `mise run up` / `mise run down` | Start or stop the local environment |
| `mise run dev` | See the dashboard and the public pages in a browser |
| `mise run mock` | Work on the frontend against a mock API before the backend exists |
| `mise run gen` | After each change of `api/openapi.yaml` |
| `mise run check` | Before the PR. A change is done only when it passes |
| `mise run test:e2e` | After `mise run up`, to check the scenarios |
| `mise run review:codex` | A review of the branch by a second model |

## 5. Rules to remember

- Only you write `Status: approved` and `Design: approved`.
- Change the document first, then the code. A small difference from the spec: the agent edits the spec and continues. A large difference: the task goes back to step 1.
- The agent makes cheap decisions and lists them in the PR. You make expensive decisions through a decision card: the data schema, isolation between accounts, external providers, money, personal data. If the class is not clear, the decision is expensive.
- A task longer than one session has `openspec/changes/<change-id>/progress.md`. Before a session ends, ask the agent for a handoff note ([01](ai-standard/01-workflow.md#session-handoff)).
- If the agent did not run a check, it must say "not verified".
