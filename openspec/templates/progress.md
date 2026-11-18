# Progress: <change-id>

<!-- The orchestrator keeps this file for a task longer than one session
(AI standard 01, "Progress file"). Read it after each context compaction or
resume, before you do anything (AI standard 07). Update it each time the loss
of the last hour of work would hurt. Long-lived knowledge goes into the spec
or an ADR, not here. No secrets and no personal data. -->

Updated: <date and time>

## Goal and phase

Goal: <the observable result of the change>
Phase: <explore | plan | code | commit>
Slice now: <slice number and name from tasks.md>

## Subagents

| Agent | Worktree or folder | Slice | State | Last check |
| --- | --- | --- | --- | --- |
| <name or ID> | <path> | <slice> | <running, done, stopped> | <time and what you saw> |

## Decisions and reasons

- <decision>: <reason>. Rejected: <option and why>

## Open problems

- <problem, evidence, owner>

## Next step

<one concrete action>
