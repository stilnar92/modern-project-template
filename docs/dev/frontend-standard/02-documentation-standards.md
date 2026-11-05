# 02 — Documentation Standards

**Status:** DRAFT
**Depends on:** [01 — Vision & Principles](01-vision-principles.md)

---

## 1. Purpose

Defines how the documents in this standard are written, structured, evolved, and kept authoritative — for human readers and, first of all, for AI agents that act on them (P1–P2, [01](01-vision-principles.md)).

## 2. Document set structure

- `README.md` — entry point: what this documentation set is and how to use it.
- `00-overview.md` — the master map: area list with statuses, the **Locked Decisions register**, and source references. The single place where global state lives.
- `NN-<area-slug>.md` — one file per area. One area = one concern; content that belongs to another area is linked, not repeated.

### Numbering

- Numbers give reading order and precedence ([01 §5](01-vision-principles.md)), not physical grouping.
- **Once a file exists on disk, it is never renumbered.** New areas get the next free number and state their group in the header (precedent: area 37, Task Playbook, belongs to group B).

## 3. Area lifecycle

Every area moves through four statuses, tracked in `00-overview.md` and in the file's own header — the two must always agree:

| Status | Meaning | Exit condition |
|---|---|---|
| `TODO` | No file yet. | A first point arrives → file is created as a collector. |
| `COLLECTING` | File holds raw points; no elaboration, no invented content. | Enough material + a deliberate decision to elaborate. |
| `DRAFT` | Full document written; under human review. | Human approval. |
| `DONE` | Approved and in force; code must comply. | Amendment process (§7) reopens it. |

### Collector file format (`COLLECTING`)

```markdown
# NN — Area Name

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### <Topic or source name>
Source: [Title](url) (author).
- point…
- point…
```

Collector rules:
- Points are recorded **as collected** — condensed, but never extended with invented material.
- Every borrowed idea cites its source inline (link + author).
- A point that contradicts an existing point is **kept, with the tension flagged** (⚠️ marker and a one-line note); contradictions are resolved at elaboration, not at collection.
- A point that clearly belongs to another area goes into that area's file; if the file doesn't exist, it is created as a collector and the overview status flipped.

## 4. Writing rules (all statuses)

1. **English**, plain and direct. Terms from the glossary (section 03, not written yet) are used exactly; near-synonyms are not rotated for style.
2. **Written for agent execution** — a rule must be actionable without asking what was meant. If a rule needs interpretation, rewrite it or add an example.
3. **Normative keywords**: MUST / MUST NOT / SHOULD / MAY in the RFC-2119 sense. In `DONE` documents, every rule uses one of them or is unambiguous imperative ("components render; logic lives in hooks/services").
4. **One fact, one place** (P2): state a rule once, link everywhere else (`[12](12-architecture.md)`-style relative links). Decision IDs (`D7`) and principle IDs (`P2`) are cited by ID.
5. **Every rule names its check** where one exists: the lint rule, CI step, or AI-review checklist item that enforces it (P-"every rule aspires to be enforceable", [01 §2.3](01-vision-principles.md)).
6. Tables for enumerable facts, prose for reasoning, code blocks for anything an agent should copy exactly.

## 5. Decisions

Two registers, different lifecycles:

- **Locked Decisions (D1, D2, …)** — upfront, standard-level constraints. Live only in `00-overview.md`; area files restate them in area-specific terms but the register is the source of truth. Changing one is an amendment to the standard (§7).
- **ADRs (Architecture Decision Records)** — decisions made during a project's life (tool picked, deviation accepted, trade-off taken). Stored in `docs/adr/NNNN-<slug>.md` in the project repo.

### ADR format

```markdown
# NNNN — <Decision title>
Date: YYYY-MM-DD
Status: proposed | accepted | superseded by NNNN

## Context      — the forces at play, why a decision was needed
## Decision     — what was decided, stated as fact
## Consequences — what becomes easier, harder, or forbidden
```

ADR rules:
- Any deviation from this standard MUST have an accepted ADR (P6). An unrecorded deviation is a bug.
- ADRs are immutable once accepted — superseded, never edited.
- Agents MAY draft ADRs; only humans accept them (section 04, not written yet).

## 6. Keeping docs and code in sync

- Documentation updates are part of the Definition of Done ([17](17-quality-gates.md)): a change that alters behavior covered by a `DONE` document isn't complete until the document is updated in the same change set.
- When code and a `DONE` document disagree, the document wins until deliberately amended ([01 §2.2](01-vision-principles.md)); the discovering party files the conflict instead of working around it.

## 7. Amendment process

1. Propose the change (human directly, or agent as a draft) with rationale.
2. For `DONE` documents and Locked Decisions: a human approves; the document's history notes what changed and why (git history suffices; significant reversals get an ADR).
3. Downstream documents that cited the changed rule are updated in the same change set (§4.4 makes them findable by ID search).
