# 0010 — mise runs the tools and the commands

Date: 2026-11-11
Status: accepted. It replaces the Taskfile part of [ADR 0007](0007-monorepo-layout.md)

## Context

Two tools ran the project. mise installed the tool versions from `mise.toml`. Task ran the shared commands from `Taskfile.yml`. Task ran with the pinned tool versions only when mise was active in the shell. An agent shell often does not activate mise.

AI agents do most of the work, and each agent has its own git worktree. A new worktree must work at once, with no manual step.

The project is new. A move to another task runner is cheap now and becomes more expensive later.

## Decision

`mise.toml` holds the tool versions and the shared commands. `mise run <task>` runs a command. `mise tasks` lists them. `Taskfile.yml` is deleted. CI calls the same tasks.

A new worktree needs no setup. mise asks for trust before it loads a project config, but two rules remove that step (mise 2026.7.5 and later, checked on 2026.10.25):

- A config in a git worktree is trusted when the same path in the main checkout is trusted. Trust the main checkout once.
- Outside CI, `mise run`, `mise install` and `mise exec` trust the config that they load, because to run them is already a decision to run the project code.

In CI, `jdx/mise-action` trusts the checkout. `min_version` keeps these rules.

Tasks are plain shell scripts. Tasks that need `.env` read the file in their own script. Compose tasks pass `--env-file .env`.

## Trade-off

- **Prefer:** simplicity (one tool, one file), operability (a new worktree and CI need no setup), changeability.
- **Give up:** independence of tools (mise, hk and jactionlint have one main author).

## Stress round

| Stressor | Affects | Verdict | Resolution |
| --- | --- | --- | --- |
| A new mise version removes implicit trust | Each new worktree | accepted | Worktree trust still comes from the main checkout. Without it, `mise run` fails with a clear message: run `mise trust` once |
| A fresh clone, not a worktree | One clone | survives | The first `mise run` or `mise install` trusts it. Only `mise tasks` asks for `mise trust` before that |
| mise paranoid mode | Each worktree | accepted | Paranoid mode turns off both rules. We do not use it |
| The mise project stops | All commands | accepted | Tasks are plain shell scripts. A move to another runner is mechanical |
| Many agents run heavy tasks at once | Local machine | accepted | Same as before. A slot limit is a separate decision |

## Consequences

- Easier: one install (`mise install`), one command style (`mise run`), the tools always match `mise.toml`, and `mise tasks --json` gives agents a machine-readable list.
- Harder: task arguments come through `usage` and shell variables (`$usage_file`), not through Task's `CLI_ARGS`.
- A machine that never trusted the main checkout must run `mise trust` once, or any `mise run`.
