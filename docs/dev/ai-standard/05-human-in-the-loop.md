# 05 — Human in the loop: decisions and learning

**Status:** DRAFT
**Decisions:** A1, A12, A13

Agents know Go, the backend and architecture better than the project owner does now. This gives two tasks. First, the decisions that stay with the human must be clear without Go knowledge. Second, the human must learn the backend during development, not only receive finished code.

## Decisions: cheap and expensive

| Class | What belongs to it | Who decides | Where it is recorded |
| --- | --- | --- | --- |
| **Cheap:** easy to undo | Names inside a module, a choice between equal libraries, function structure, log format, the order of steps in a task | The agent | `design.md` of the OpenSpec change: the decision and one line "in plain words"; the PR description lists them. The human reads them when convenient and can cancel them |
| **Expensive:** expensive or impossible to undo | Data schema and keys, account isolation, event format, a breaking API change, an external provider, security, money, personal data, a new service | The human | A decision card → an ADR after approval, from the [ADR template](../adr/TEMPLATE.md) |

If you are not sure about the class, the decision is expensive.

### Decision card

The agent brings an expensive decision in this form. The card is clear without Go knowledge, and it explains each term in place.

```markdown
## Decision: <what we decide, in one sentence>

**In plain words.** 2–3 sentences: what happens and why we must decide now.

| Option | What we get | What it costs | Undo |
| --- | --- | --- | --- |
| A | ... | time, money, risk | easy / expensive / impossible |
| B | ... | ... | ... |

**If we are wrong.** What breaks and how much the fix costs.
**Trade-off.** Prefer: <axes>. Give up: <axes>. Use the nine axes below.
**Stress round.** The stressors that we checked, and the verdict for each one (see below).
**Recommendation.** The option and one reason.
**Terms.** A new word → an explanation through an analogy with what the owner knows (frontend, Django, Node.js) → a link.
**How we check it.** The sign that tells us in a month that the decision is correct.
```

- Use analogies with what the owner knows: well — the frontend; a little — Django, Node.js and databases. sqlc is like openapi-typescript, but for SQL; a River job is like a Celery task, but in Postgres.
- If something is not clear, the owner asks "explain it more simply" or "what breaks if we select B". The agent explains again in a different way; it does not repeat itself.
- For a hard choice with several strong options, use `/find-solution`. The card gets the result.

### Trade-off axes

Each decision records what it prefers and what it gives up. Use these nine axes, so decisions are easy to compare. Source: [Archiplan](https://github.com/archiplan-ai/Archiplan) (MIT).

| Axis | Meaning |
| --- | --- |
| Simplicity | The fewest moving parts. Easy to understand and to change |
| Performance | Response time and throughput in normal work |
| Scalability | Room to grow with load, data or users |
| Reliability | The system stays available and recovers after a failure |
| Security | Protection of data and access |
| Correctness | The system does what the spec says, in all cases |
| Evolvability | New requirements are cheap to add |
| Operability | Easy to deploy, monitor and fix in production |
| Cost | Money and time to build and to run |

The project priorities say what the project is willing to be bad at. They are the reference for each later decision.

### Stress round

Before an expensive decision or an ADR is accepted, the agent attacks it with stressors. Source: [Archiplan](https://github.com/archiplan-ai/Archiplan) (MIT), adapted to our process.

1. **Select stressors from different angles:** scale, security, regulation, failures, multi-tenancy, operational load, long-term evolution. Start with the most coupled modules and with the parts that no earlier round checked.
2. **One stressor is one pressure.** For each stressor, write:
   - the pressure, as one sentence
   - the components that it affects
   - the attractor: the state that the system is pushed toward
3. **Give a verdict:** survives, breaks or accepted.
4. **If a stressor breaks, the agent does not decide alone.** It gives the owner two options, each with its price on the axes:
   - **Fix:** a new requirement. Its origin is this stressor.
   - **Accept:** a signed trade-off in the decision: what we live with, and what we prefer instead.
5. **Repeat** until a round survives, or the owner agrees that no new stressors appear.
6. **Hidden coupling:** if two components break from the same stressor but have no declared link, there is a hidden dependency. Add the link to the design, or separate the shared concern.

Typical stressors for this project:

| Angle | Stressor |
| --- | --- |
| Failure | Postgres is not available for 5 minutes |
| Failure | An external provider stops sending webhooks for an hour |
| Failure | The LLM provider is down, or its price doubles |
| Security | A user tries to read another account's data through the public API |
| Security | A spam script sends 10,000 requests to a public form |
| Abuse | A visitor tries to bypass the rate limits of an expensive feature |
| Regulation | A user asks to delete all their personal data |
| Scale | 1,000 accounts and 100,000 events a day |
| Evolution | A new customer segment needs a new field in a core entity |

Codex runs the round as an adversarial review ([03](03-review-verification.md)). The result goes into the decision card or the ADR.

## Learning during development

The owner learns from the real problems of the tasks. The format, the rules and the list of lessons are in [docs/learning/](../../learning/README.md).

- **A lesson is one problem, not one PR.** When a change solved a problem that had a real choice, the agent writes a lesson in the same PR. One PR can give several lessons. A PR with no real choice gives none.
- **Each lesson has the same sections:** problem, predict first, theory with sources, options, edge cases, decision, the files to read, check yourself.
- **The theory stands without the project,** and each source is real: the agent opens each link before it adds it.
- **The agent takes the options and the edge cases from the session:** what it considered, what it rejected and why. It does not invent them after the work.
- **A later change that replaces a decision** also changes the lesson, in the same PR.
- **To study a lesson,** the owner runs `/lesson <number>`. The agent is then a tutor: it asks before it explains, and it does not change a file ([how](../../learning/README.md#study-a-lesson-with-an-agent)).
- **When the owner asks about a topic,** the agent asks for a prediction first, and then explains.
