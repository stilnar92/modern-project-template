# 05 — Spec-Driven Development

**Status:** DRAFT
**Depends on:** [01 — Vision & Principles](01-vision-principles.md), [02 — Documentation Standards](02-documentation-standards.md)
**Used by:** [08 — Agent Workflows](08-agent-workflows.md), [37 — Task Playbook](37-task-playbook.md), [17 — Quality Gates](17-quality-gates.md)
**Tooling:** [OpenSpec](https://openspec.dev/) — adopted as the spec framework.

---

## 1. Purpose

No feature without a spec (P1). The spec is the unit of planned work: it states the problem, the requirements, and the acceptance scenarios *before* implementation starts, and it is the reference that code, tests, and the AI review (D10) are checked against.

Specs are managed with **OpenSpec**: living, capability-scoped requirement documents checked into the repo, evolved through change proposals. OpenSpec is agent-agnostic and natively integrates with Claude Code (harness setup in [06](06-ai-harness.md)).

## 2. What requires a spec change

| Change | Spec? |
|---|---|
| New feature or user-visible behavior change | MUST |
| Bug fix that changes intended behavior (not just restores it) | MUST |
| Refactor crossing module boundaries ([12](12-architecture.md)) | MUST — proposal with spec deltas, even if requirements don't change |
| Bug fix restoring specified behavior | MAY — the reproduction + failing test acts as the spec |
| Trivial change (typo, copy tweak, dependency bump within policy) | No spec |

If it's unclear whether a change is trivial, it isn't.

## 3. Repository layout (OpenSpec)

```
openspec/
├── specs/                      # current truth: one folder per capability
│   ├── <capability>/spec.md    #   e.g. auth-login, checkout-cart
│   └── …
└── changes/                    # in-flight and archived change sets
    └── <change-id>/
        ├── proposal.md         # why + what changes, impact analysis
        ├── design.md           # technical decisions and rationale
        ├── tasks.md            # broken-down implementation steps
        └── specs/…             # spec deltas for affected capabilities
```

- **Specs are per capability**, not per ticket: `openspec/specs/<capability>/spec.md` is the current truth for that capability. Capability boundaries SHOULD align with module boundaries ([12](12-architecture.md)).
- **Changes are proposals against specs**: work enters as `openspec/changes/<change-id>/` containing proposal, design, tasks, and spec deltas.
- The tracker ticket (e.g. Jira) links to the change-id; the proposal links back. The spec, not the ticket, is the source of truth for scope — tickets summarize, specs define.
- PRs MUST reference the change-id.

## 4. Spec format

OpenSpec's requirement–scenario structure, Gherkin-style:

```markdown
### Requirement: <what the system SHALL do>
#### Scenario: <specific case>
- GIVEN <initial state>
- WHEN <action>
- THEN <expected outcome>
```

- Requirements use SHALL/MUST language ([02 §4](02-documentation-standards.md)).
- **Scenarios are the acceptance criteria**: each scenario is a concrete, observable pass/fail statement and maps to at least one test or explicit verification step.
- Out-of-scope statements live in the proposal; blast radius ([37](37-task-playbook.md) step 2) goes in the proposal's impact analysis.
- Large changes are sliced in `tasks.md`: MVP slice first, then increments — each slice tied to its scenarios.

## 5. Lifecycle

```
propose ──review/approve (human)──▶ implement ──gates pass──▶ archive
   ▲                                    │
   └────────── scope change ────────────┘
```

1. **Propose** — human or agent creates the change (`/openspec:proposal`): proposal, design, tasks, spec deltas — *before any code*. Agents drafting proposals MUST surface open questions rather than resolve them by assumption.
2. **Review & approve** — a human approves the proposal (section 04, not written yet). Only approved changes are implemented.
3. **Implement** — execute `tasks.md`; the diff must stay within the proposal's scope.
4. **Archive** — after quality gates pass ([17](17-quality-gates.md)) and scenarios are demonstrably met, spec deltas merge into `openspec/specs/` (the new current truth) and the change folder is archived as the historical record.

**Scope changes reopen the proposal.** If implementation reveals the spec is wrong or incomplete, the proposal and deltas are amended and re-approved first — code never silently diverges from an approved spec ([01 §2.2](01-vision-principles.md)).

## 6. Traceability

- **Scenario → test:** every scenario maps to at least one test or verification step; the mapping is stated in the PR ("Scenario 'expired card' → `checkout.spec.ts`; 'split-screen layout' → manual verification, screenshot attached").
- **Change → code:** PR references the change-id; commits for sliced work reference the task.
- **Spec → review:** the AI review checkpoint (D10) checks the diff against the proposal and spec deltas — implemented-but-unspecified behavior is a finding, same as specified-but-missing.

## 7. Specs and AI agents

- The approved proposal is the agent's contract: the agent implements what it says — not what the ticket title implies, and not improvements it thinks of along the way (those become new proposals).
- Ambiguity in an approved spec is a blocker to raise, not a gap to fill creatively.
- Specs and proposals are written to be executable by an agent with no conversation context: self-contained, explicit, measurable (same bar as [02 §4](02-documentation-standards.md)). Archived changes preserve that context across sessions and team transitions.

## 8. Spec vs design.md vs ADR

- **spec.md** — *what* the capability does (requirements + scenarios); current truth.
- **design.md** — *how* a specific change implements it; decisions local to that change.
- **ADR** ([02 §5](02-documentation-standards.md)) — decisions whose consequences outlive the change (tool choice, standard deviation, cross-cutting trade-off). If a `design.md` decision has lasting consequences, it is promoted to an ADR and linked.
