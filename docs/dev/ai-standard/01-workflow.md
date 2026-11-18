# 01 — Workflow

**Status:** DRAFT
**Decisions:** A1, A2, A3, A4, A9

## Four phases

Source: [Agentic Design Patterns](https://mokevnin.github.io/agentic-coding-design-patterns/ru/explore-plan-code-commit.html). The phases map to the [task cycle](../dev-process.md#cycle-for-each-task).

| Phase | What the agent does | Cycle steps | Output | Gate |
| --- | --- | --- | --- | --- |
| Explore | Reads the capability spec, the code, `CONTEXT.md`, ADRs. Changes nothing. Asks questions | 1. Clarify | Questions and answers, `CONTEXT.md` edits | — |
| Plan | Writes the OpenSpec proposal, deltas, `tasks.md`. For a new screen or flow: a prototype for the owner and the "Screens and contract" table. Changes the contract after approval | 2–5 | `openspec/changes/<id>/`, `design/` screenshots and prototype HTML, `openapi.yaml` | The owner writes `Design: approved` after the prototype → a fresh agent reviews the plan → the owner writes `Status: approved`; expensive decisions come as a card ([05](05-human-in-the-loop.md#decision-card)) |
| Code | Implements with TDD, strictly inside the proposal. A task in more than one layer starts with a tracer slice | 6–7 | Code and tests | Tests, linters, code generation |
| Commit | Checks the scenarios, updates documents, commits, prepares the PR | 8–9 | A PR with the "scenario → test" links | AI review and the human |

The short path for small fixes and for bugs with a reproduction: the "code" and "commit" phases only ([development process](../dev-process.md#cycle-for-each-task)).

## Patterns we use

| Pattern | When | With |
| --- | --- | --- |
| Grilling | The task is unclear; there are "what if" questions | `grilling`, `grill-with-docs` |
| Design it twice | Choosing an approach or architecture | `/find-solution` (blind-solver + solution-discriminator) |
| Tracer bullet | Required for each task in more than one layer: first a thin end-to-end slice, then parallel work | The first slice goes through all [containers](../architecture.md#c4-containers): API, database, web. See [parallel work](../dev-process.md#one-task-several-agents) |
| Design approval | A new screen or flow | A clickable prototype as a Claude Design artifact, from the design system of the repository, with sample data; the owner writes `Design: approved` ([cycle](../dev-process.md#cycle-for-each-task)) |
| TDD with an agent | All new code | `tdd` |
| Hypothesis-driven debugging | A bug with an unclear cause | `diagnosing-bugs` |
| One writes, another checks | Each PR and each plan | [03](03-review-verification.md) |
| Throwaway prototype | A technical question without an answer | `prototype`; the result goes into an ADR |
| Progress file | A task longer than one session, with the same goal | `openspec/changes/<id>/progress.md` from [the template](../../../openspec/templates/progress.md). Read it again after each context compaction ([07](07-delegation.md#after-a-context-compaction)) |
| Delegation by contract | A subagent does a part of the task | The executor contract and the brief template ([07](07-delegation.md)) |
| Session handoff | The session ends before the task ends, or the work type changes (for example, from plan to prototype) | A handoff note ([below](#session-handoff)) |
| Reproducible bootstrap | Each new session and each worktree | `mise run setup` ([below](#reproducible-bootstrap)) |

## Session handoff

Source: [Agentic Design Patterns, "Handoff"](https://mokevnin.github.io/agentic-coding-design-patterns/ru/handoff.html).

An automatic summary often loses why an option was rejected. Then the next session does the same research again. So, before a session ends, the agent writes a handoff note for the goal of the next session, while it still has the context.

**The note contains:**

- the current state and the goal of the next session;
- the key decisions and the reasons for them;
- the approaches that were tried and rejected;
- one concrete next step;
- links to specs, ADRs, commits and issues, not copies;
- the skills and tools that the next session needs.

**Rules:**

- Put the note in the scratchpad or a temporary folder, not in the repository. It is for one transfer only.
- Move long-lived knowledge into the spec, an ADR or the progress file. The note does not replace them.
- Remove secrets and personal data from the note.
- The next session starts with the note and follows its links when necessary.

## Reproducible bootstrap

Source: [Agentic Design Patterns, "Reproducible agent bootstrap"](https://mokevnin.github.io/agentic-coding-design-patterns/ru/reproducible-agent-bootstrap.html).

A session must not guess how to start the project. If a test fails, the agent must know if its change caused the failure or if the environment was already broken. So one command brings the project to a checked "green baseline", and the agent starts work only after it.

**`mise run setup` does this:**

1. mise installs the tool versions from `mise.toml` (Go, Node, pnpm, linters). The task stops with a clear message when Docker is not running.
2. Installs the web dependencies from the lock file.
3. Creates `.env` from `.env.example` only if `.env` does not exist, with the ports of this worktree (`deploy/worktree-env.sh`). It never uses production secrets.
4. Starts the services and applies the migrations (`mise run up`). There is no seed data yet.
5. Does a smoke test through the real interface: a request to `$CORE_URL/healthz` with time limits.
6. Prints "baseline: green", or stops with a non-zero exit code and the next safe step.

**Rules:**

- The command is idempotent: you can run it again at any time.
- Each worktree gets its own ports and Compose project name, so parallel agents do not interfere ([parallel work](#parallel-work)).
- CI runs `mise run setup` on a clean checkout before the end-to-end tests, so the command does not become outdated.
- **Not yet:** a smoke test of a public page. Add it when the first public page reads data from the core.

## Anti-patterns

- **Code without a plan** and "the whole feature in one prompt": you cannot compare the result with anything.
- **Premature specification:** a detailed spec of something that nobody understands yet. Do a prototype or grilling first.
- **Premature success:** "done" without checks that ran (A8).
- **A bloated `CLAUDE.md`:** everything in one file, so the agent stops seeing what is important.
- **Vibe coding:** accepting code that nobody read and nobody checked.
- **Improvements on the way:** the agent does things that are not in the proposal. Put the idea into a new proposal.

## Parallel work

The rules are in the [development process](../dev-process.md#parallel-work). In short:

- The orchestrator fixes the proposal and the contract. Only after that, subagents work in parallel, each in its own folder.
- Only the orchestrator changes shared files: `openapi.yaml`, migrations, `events.md`, `CONTEXT.md`, `THESAURUS.md`, `openspec/specs/`. A hook enforces this rule ([02](02-harness.md#hooks)).
- If a field is missing, the subagent stops and tells the orchestrator.
- Each brief starts with the executor contract and follows the brief template. The orchestrator checks each subagent in its first minute and accepts the work only on evidence ([07](07-delegation.md)).
- Independent tasks go into separate git worktrees. No more than 2–3 streams: the limit is the review capacity of one person.
