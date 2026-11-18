Closes #

**OpenSpec change:** `openspec/changes/<id>` for a `feat/` branch. For a `chore/`, `fix/` or `docs/` branch: "none, tooling".

## Summary

What changed and why, in plain words.

## Scenarios → checks

| Scenario | Test or check |
| --- | --- |
| | |

## Cheap decisions

What the agent decided alone, and why ([rules](https://github.com/stilnar92/modern-project-template/blob/main/docs/dev/ai-standard/05-human-in-the-loop.md#decisions-cheap-and-expensive)).

## Evidence

The commands that ran and their results. Write "not verified" for each thing that did not run in the real environment.

## Checks

- [ ] `mise run check` passes (lint, test, gen:check)
- [ ] Each new check failed once on purpose
- [ ] Migrations only add things (expand → migrate → contract), or there is no migration
- [ ] Review in a fresh context. ChatGPT review: N comments, M fixed, K rejected with reasons
- [ ] Documents updated where the change touches them: spec, `data-map.md`, `events.md`, module READMEs
- [ ] A lesson in `docs/learning/lessons/` for each problem that had a real choice, or none

## Merge danger

**Door:** two-way (a revert undoes it) or one-way (data, external effects, a published contract).

**Blast radius:** what can break when this is wrong.
