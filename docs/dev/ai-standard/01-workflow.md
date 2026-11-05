# 01 — Workflow

**Status:** DRAFT
**Decisions:** A1, A2, A3, A4, A9

## Four phases

Source: [Agentic Design Patterns](https://mokevnin.github.io/agentic-coding-design-patterns/ru/explore-plan-code-commit.html). The phases map to the [task cycle](../dev-process.md#cycle-for-each-task).

| Phase | What the agent does | Cycle steps | Output | Gate |
| --- | --- | --- | --- | --- |
| Explore | Reads the capability spec, the code, `CONTEXT.md`, ADRs. Changes nothing. Asks questions | 1. Clarify | Questions and answers, `CONTEXT.md` edits | — |
| Plan | Writes the OpenSpec proposal, deltas, `tasks.md`. Changes the contract after approval | 2–4 | `openspec/changes/<id>/`, `openapi.yaml` | A fresh agent reviews the plan → the human approves; expensive decisions come as a card ([05](05-human-in-the-loop.md#decision-card)) |
| Code | Implements with TDD, strictly inside the proposal | 5–6 | Code and tests | Tests, linters, code generation |
| Commit | Checks the scenarios, updates documents, commits, prepares the PR | 7–8 | A PR with the "scenario → test" links | AI review and the human |

The short path for small fixes and for bugs with a reproduction: the "code" and "commit" phases only ([development process](../dev-process.md#cycle-for-each-task)).

## Patterns we use

| Pattern | When | With |
| --- | --- | --- |
| Grilling | The task is unclear; there are "what if" questions | `grilling`, `grill-with-docs` |
| Design it twice | Choosing an approach or architecture | `/find-solution` (blind-solver + solution-discriminator) |
| Tracer bullet | A new capability: first a thin end-to-end slice | The first slice goes through all [containers](../architecture.md#c4-containers): API, database, web |
| TDD with an agent | All new code | `tdd` |
| Hypothesis-driven debugging | A bug with an unclear cause | `diagnosing-bugs` |
| One writes, another checks | Each PR and each plan | [03](03-review-verification.md) |
| Throwaway prototype | A technical question without an answer | `prototype`; the result goes into an ADR |
| Progress file | A task longer than one session, with the same goal | A progress file in `openspec/changes/<id>/` |
| Session handoff | The session ends before the task ends, or the work type changes (for example, from plan to prototype) | A handoff note ([below](#session-handoff)) |
| Reproducible bootstrap | Each new session and each worktree | `task setup` ([below](#reproducible-bootstrap)) |

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

**`task setup` must:**

1. Check the tool versions (Go, Node, pnpm, Docker) and stop with a clear message that tells how to fix the problem.
2. Install the dependencies from the lock files.
3. Create `.env` from `.env.example` only if `.env` does not exist. Never use production secrets.
4. Start the services, apply the migrations and load the seed data.
5. Do a smoke test through the real interface (for example, an HTTP request to `/healthz` and to one public page), with time limits.
6. Print "baseline: green", or stop with a non-zero exit code and the next safe step.

**Rules:**

- The command is idempotent: you can run it again at any time.
- Each worktree gets its own ports, database name and Compose project name, so parallel agents do not interfere ([parallel work](#parallel-work)).
- CI runs `task setup` on a clean checkout, so the command does not become outdated.

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
- Only the orchestrator changes shared files: `openapi.yaml`, migrations, `events.md`, `CONTEXT.md`, `openspec/specs/`. A hook enforces this rule ([02](02-harness.md#hooks)).
- If a field is missing, the subagent stops and tells the orchestrator.
- Independent tasks go into separate git worktrees. No more than 2–3 streams: the limit is the review capacity of one person.
