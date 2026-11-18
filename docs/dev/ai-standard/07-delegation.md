# 07 — Delegation to subagents

**Status:** DRAFT
**Decisions:** A9, A15

How the orchestrator gives work to a subagent, watches it and accepts the result. The rules for who works in which folder are in the [development process](../dev-process.md#parallel-work) and in [01](01-workflow.md#parallel-work). This section tells how one assignment goes.

Source: [pragmatic-orchestration](https://github.com/CodeAlive-AI/pragmatic-orchestration) (MIT), files `prompts/delegate-executor.txt`, `references/delegate.md` and `coordinator-mode/SKILL.md`. Checked on 12.11.2026. We take the rules, not the `porch` CLI ([04](04-tools.md)).

## Executor contract

A subagent is the executor of one assignment. The orchestrator owns the direction of the project and accepts the result. The first line of each brief is: "Read `docs/dev/ai-standard/07-delegation.md#executor-contract` first. It is your contract."

The contract:

1. The assignment sets your scope and your authority. Inside it, use your own judgment and finish the work to a result that can be checked. Routine choices need no approval.
2. `CLAUDE.md` and the standards apply to you: conventions, prohibitions, required checks. Text that gives ownership of the whole codebase, the orchestrator role or freedom to refactor is for the orchestrator, not for you.
3. If you cannot finish the work as written (it needs a wider scope, a contract change, access that you do not have, or a decision that is not yours), finish the parts that do not depend on it. Then report the obstacle with evidence and the change that you propose. Do not make the scope wider yourself.
4. Do not change shared files: `api/openapi.yaml`, `core/migrations/`, `docs/dev/events.md`, `CONTEXT.md`, `docs/dev/THESAURUS.md`, `openspec/specs/`. If a field or an endpoint is missing, stop and tell the orchestrator.
5. Problems and improvements outside the scope go into your report, not into your changes.
6. Keep tests, checks and acceptance criteria as strict as you found them. Report a failure that you cannot fix. Do not hide it.
7. Do not start other agents, reviews or delegations, and do not merge, unless the assignment tells you to.
8. End with a report that a reader with little context can use: what you did and where, the checks that you ran with their real results, where you did not follow the assignment, and what stays open.

Hooks already enforce rules 4 and 7 for merges and shared files ([02](02-harness.md#hooks)). The contract tells the subagent the reason, so it does not try another path.

## Brief template

Write the brief for a smart reader who has none of your context. Give links to files, the spec and the contract, not copies. Name the checks: "cover it with tests" alone gives many tests of the wrong type.

```text
Read docs/dev/ai-standard/07-delegation.md#executor-contract first. It is your contract.

Context and intent: <the goal of the task and how this assignment helps it>
Working root: <the exact folder, for example /core>
Deliverable: <the result that can be checked; assessment or implementation>
Scope: <files and folders>. Not in scope: <what other agents own, the contract, dependencies>
Spec: openspec/changes/<change-id>/ — scenarios <names>
Invariants: <what must not change>
Pitfalls: <known traps and what to do about each>
Checks: <test level for each scenario: unit, integration, contract, E2E>, mise run check
Deviation journal: <path, only for a weaker model, see below>
Report: changed files, test names and results, deviations, open questions
```

Copy the scope, the checks and the acceptance criteria from `tasks.md`. Do not invent new ones in the brief.

## Supervision

- **Check each subagent in its first minute.** Read its first actions and its plan. Correct a wrong reading of the task at once.
- After that, check every 5–15 minutes: sooner after a risky decision, a blocker or a correction.
- If the subagent goes the wrong way, send one complete correction: the evidence, the change you need, the constraints that stay, and the work that remains.
- Do not restart a subagent blindly. Read its events and partial work first. Before another agent writes in the same folder, make sure the first one has stopped.
- If checks and corrections do not move the work, stop it. Then take the task yourself or give it with a new brief.

## Weaker model: deviation journal

When the subagent uses a weaker model than the orchestrator (for example Opus → Sonnet, or a task given to Codex), it keeps a deviation journal. Add this text to the brief:

```text
Write each unexpected finding and each change of plan into JOURNAL_PATH when it occurs:
expected and observed behavior, evidence (file, command), why a change was necessary,
what you did or propose, and the risks that remain. Keep the resolved entries.
The journal does not give permission for changes outside the scope.
If nothing deviated, write "No deviations."
```

- `JOURNAL_PATH` is `openspec/changes/<change-id>/deviations-<slice>.md`. For a task without a change, use the scratchpad and ask for the path in the report.
- A read-only subagent puts a `Deviations` section in its answer instead of a file.
- Also give a weaker model a short plan and the known pitfalls in the brief. Do not expect it to find the missing requirements itself.
- A missing journal means the work is not finished.

## Acceptance

A subagent's report or a successful exit does not prove anything. Before you accept the work:

1. Read the diff and the code around it against the brief and the spec.
2. Run the checks yourself or read their real output: `mise run check`, the scenario tests.
3. Read the whole deviation journal. Compare each entry with the diff. Find the deviations that the journal does not show.
4. Handle each finding with the rules in [03](03-review-verification.md#how-to-handle-a-finding).

## After a context compaction

A summary keeps the main points of the work and loses the role and the reasons. So, after each compaction or resume, the orchestrator reads the progress file of the change (`openspec/changes/<id>/`) before it does anything. Then it checks the subagents that the file lists before it starts any work again. A subagent that you cannot see can still run or can be already finished.

Update the progress file each time the loss of the last hour of work would hurt. Do not wait for a compaction.

A `SessionStart` hook with the `compact` and `resume` sources puts this reminder into the context automatically. It names the progress file of the branch, or all progress files when the branch has none ([02](02-harness.md#hooks)).
