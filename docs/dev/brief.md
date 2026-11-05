# Brief: what we build

The product on one page, for agents and for each person who takes a task. Fill in each section in a few lines. Put the details in other documents and link to them. Keep the brief short: an agent reads it before each task.

## What it is

<!-- Two or three sentences: the problem, who has it, and what the product does about it. -->

## Who it is for

<!-- The main user types. For each type: what they do in the product and who pays. -->

## How it is built

<!-- The layers of the product. Keep the table short. -->

| Layer | What it contains | Status |
| --- | --- | --- |
| Shared core | <!-- modules that each product feature needs --> | |
| Product modules | <!-- features on top of the core --> | |

The containers and the code structure are in the [architecture](architecture.md).

## Main scenarios

<!-- Three to five end-to-end scenarios. Each scenario is one line: actor, action, result.
Example: "A user signs in with a code, changes the profile name and sees it on the public page." -->

1.
2.
3.

## Not in the MVP

<!-- What we do not build now, so that agents do not add it. -->

## Limits

- **Account isolation:** a filter in code and Postgres rules from the first query ([ADR 0009](adr/0009-account-isolation-rls.md)).
- **Stack:** a Go core, Postgres and River; the public site on Astro; the dashboard on React and Mantine; the contract `api/openapi.yaml` ([tech stack](tech-stack.md)).
<!-- Add legal, data residency, budget and hosting limits here. -->

## How we work

OpenSpec spec → contract → code with TDD → review → PR. The rules are in the [development process](dev-process.md), the [backend](backend-standard/README.md), [frontend](frontend-standard/README.md) and [AI](ai-standard/README.md) standards. Decisions are in the [ADRs](adr/).
