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

- [ ] `task lint`, `task test` and `task gen:check` pass
- [ ] Each new check failed once on purpose
- [ ] Migrations only add things (expand → migrate → contract), or there is no migration
- [ ] Review in a fresh context. ChatGPT review: N comments, M fixed, K rejected with reasons
- [ ] Documents updated where the change touches them: spec, `CONTEXT.md`, `docs/dev/data-map.md`, module READMEs
- [ ] Learning debrief in `docs/learning/debriefs/` when the backend changed and the team keeps learning notes

## Merge danger

**Door:** two-way (a revert undoes it) or one-way (data, external effects, a published contract).

**Blast radius:** what can break when this is wrong.
