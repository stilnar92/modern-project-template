# 0009 — How to trust code that an agent wrote: fast checks first

From PRs #55, #57, #61, #89 · 11.11.2026 · Topics: quality gates, agent workflow

## Problem

Agents write the product code. The owner cannot read each line with full attention, and an agent can forget a rule that is only text. We need a way to trust a change that does not depend on memory or on a careful reader.

## Predict first

You add a new linter to CI. The CI run is green on the first try. What do you know now about your code, and what do you not know about the linter?

## Theory

A **quality gate** is an automatic check that stops a change when it finds a problem. The agent does the work, runs the check, reads the error and fixes it. A rule that a tool checks is stronger than a rule in text, because nobody can forget it.

Three ideas make gates work:

- **Tiers by speed.** A check that takes seconds runs often: after each edit and before each commit. A check that takes minutes runs in CI. A check that takes longer, or that needs the network, runs on a schedule. A slow check in the fast tier makes people and agents skip the loop. Analogy: the editor shows a type error immediately, `tsc` runs before the push, and the end-to-end tests run in CI.
- **A new gate needs a start state.** An old codebase gives many findings on the first run. You fix all of them, or you record them in a **baseline** file. With a baseline, the gate fails only on new findings, and the file only gets smaller.
- **A gate must fail once on purpose.** A green check has two explanations: the code is good, or the check does not run. Make a deliberate violation and see the red result. This is the "red first" step of test-driven development, applied to the tool.

Sources:

- Humble and Farley, "Continuous Delivery", chapter 7, "The Commit Stage": a fast first stage, slower stages after it.
- [Continuous Integration](https://martinfowler.com/articles/continuousIntegration.html), Martin Fowler: the section "Keep the Build Fast" and the idea of a second, slower build.
- [dev-quality-gates](https://github.com/mokevnin/dev-quality-gates): the list of jobs and tools that we applied.
- [Claude Code hooks](https://code.claude.com/docs/en/hooks): how a hook runs a check before a tool call.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Rules in text only | `CLAUDE.md` tells the agent what to do | An agent can forget a line. Nothing fails |
| A person reads each change | The owner is the gate | Slow. The reader gets tired and misses things |
| All checks before each commit | One large gate | The agent loop becomes slow |
| Tiers: seconds at the edit and the commit, the rest in CI and on a schedule | Each check runs where its speed fits | Each gate needs a row in a table: where it runs |
| A baseline for a new gate | Old findings are recorded. Only new findings fail | The old findings stay until somebody removes them |

## Edge cases

| What if | What happens |
| --- | --- |
| The commit hook passes its time limit | A hook that runs too long does not block the tool call. So the rule has its own budget of 45 seconds, and a check that does not finish blocks the commit |
| The agent runs `git add -A && git commit` | The hook runs before the command, so the new file is not staged yet. The rule checks staged, changed and new files |
| A tool exits 0 on warnings | The gate never fails. We pass the strict flags (`--strict-exit`, `--fail-severity warn`), so a warning fails too |
| The tool is not on the `PATH` of the hook shell | The check does not run and nobody sees it. The hook gets the tool paths from mise. The deliberate violation shows this type of failure |
| The first run gives 2007 findings (rumdl) | We disabled one rule that conflicts with our style, used the autofix for 99, and fixed the others. We did not need a baseline |
| An agent disables a rule to make a change pass | `CLAUDE.md` forbids it. An exception covers one path and one rule, has a reason, and needs the owner |
| A check needs the network or takes minutes | It leaves the fast tiers. The race detector and the online link check run each night |

## Decision

We use tiers. After an edit and before a commit, only the format checks, `go vet`, oxlint and the secret scan run. `mise run check` runs all linters, the tests and the generated-code check, and CI runs the same tasks. Each finding fails the gate: there is no level of warnings that you can merge. Each new gate started at zero findings, and each PR that added a gate shows the deliberate violation.

This would be wrong for a large old codebase. There, "zero findings first" can stop all other work, and a baseline is the correct start. It is also not complete: a gate proves only what it checks, so a review in a fresh context stays.

Record: the section "Quality gates" in `docs/dev/ai-standard/03-review-verification.md`. The toolchain is in [ADR 0010](../../dev/adr/0010-mise-single-toolchain.md).

## In our code

1. `docs/dev/ai-standard/03-review-verification.md`, the table "Quality gates": each gate and where it runs.
2. `.claude/hooks/guard.mjs`, rule 5: the fast checks before `git commit`, with the time budget.
3. `mise.toml`, the tasks `check` and `lint`: one command for all checks.
4. `.github/workflows/nightly.yml`: the slow checks on a schedule.

## Check yourself

1. Why does the commit hook not run the full test suite?
2. A new gate is green on its first run. What do you do before you trust it, and why?
3. When is a baseline file better than "fix all findings now"?

<details><summary>Answers</summary>

1. The tests take much longer than seconds. A slow check at each commit makes the agent loop slow, and a hook that passes its time limit does not block. So the hook runs only fast checks, and the tests run in `mise run check` and in CI.
2. Make a deliberate violation and see the gate fail. A green result does not tell if the code is good or if the check never ran, for example because the tool is not installed or exits 0 on warnings.
3. When the first run gives more findings than you can fix in one change. The baseline lets the gate stop new findings immediately. The rule is that the baseline only gets smaller.

</details>
