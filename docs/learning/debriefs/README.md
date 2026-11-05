# Debriefs

The agent writes a debrief at the end of each task that changes the backend, as a teacher after a lesson. File name: `NNNN-<OpenSpec change id>.md`, numbered in order. A small fix with nothing new needs no debrief. If it has a new concept, add the concept to the [concept map](../concepts.md).

## How to write

- **Big picture first, then details.** Write for the learner described in the [concept map](../concepts.md).
- **Explain "why", not only "what".** A decision without a reason is hard to remember.
- **Use analogies** with stacks that the learner knows, where they are honest. Say where an analogy fails. Say where past experience can mislead.
- **Do not hide complexity, but mark it:** "you can skip this for now, we come back when it appears again".
- **Link to the past:** a repeated pattern links to the debriefs where it appeared.
- **Keep it short:** 5–10 minutes to read.

## Template

```markdown
# NNNN — <the task in plain words>

Date · change `openspec/changes/<id>` · PR <link>

## What we did
3–5 sentences: what the system can do now and why the user needs it.

## How we did it
The steps in order, with links to files. How to read the diff: where to start, what is important, what is secondary.

## Why this way
The options and why we selected this one. The cheap decisions of the agent, in plain words.

## New
- **<concept>** — in plain words. Analogy: … Source: <link>.
- You can skip for now: …

## Seen before
- **<pattern>** — it was in [NNNN](NNNN-....md). Same: … Different: …

## Check yourself
1. A "why" question.
2. A "what breaks if…" question.
<details><summary>Answers</summary>…</details>

## Try it yourself (optional)
A small exercise on what appeared, at the level from the concept map.
```
