# 01 — Modules and coupling

**Status:** DRAFT
**Decisions:** B1

## Module

- A module is the folder `/core/internal/<module>`. It is a core module or a product module ([architecture](../architecture.md#rules-for-the-code-structure)). A module is the same as a capability in OpenSpec ([ADR 0002](../adr/0002-specs-openspec-hybrid.md)).
- A module has one public entry: the service interface and its types. Other modules call only this entry.
- A module owns its tables: only this module contains SQL for them. Do not read or change the tables of another module directly, not even "only SELECT".
- Each module has a `README.md`: boundaries, tables, events, dependencies ([development process](../dev-process.md#documentation-in-layers)).

## Deep modules

Source: John Ousterhout, *A Philosophy of Software Design*.

- The module interface is small, and the implementation is large. A good module hides complexity; it does not repeat it.
- Do not add small wrappers "just in case". A layer that only passes a call on adds code but hides no decision. Source: [Indirection Is Not Abstraction](https://www.silasreinagel.com/blog/2018/10/30/indirection-is-not-abstraction/) (Silas Reinagel).
- When possible, "define errors out of existence": the module handles the expected cases itself and does not push them to all callers.

## Balanced coupling

Sources: Vlad Khononov, *Balancing Coupling in Software Design*, and [coupling.dev](https://coupling.dev). This section describes the method in our own words.

Measure the coupling between two parts on three axes:

| Axis | Question | Our example |
| --- | --- | --- |
| Strength | How much does one part know about the other? From the most to the least: **intrusive** (private tables, internal types), **functional** (the same business rule in two places), **model** (a shared domain model), **contract** (only a published interface) | The "Orders" module knows the "Catalog" interface, but not its tables |
| Distance | How expensive is it to change them together: the same package, another module, another service? A synchronous call keeps the parts closer. An asynchronous event moves them further apart | The core and an external payment provider are different services: a large distance |
| Volatility | How often does the part change for business reasons? Judge this from the business, not from the commit history | Product modules change often. Accounts and addresses are stable |

**The balance rule:** coupling is balanced when strength and distance compensate each other: strong coupling at a small distance, or weak coupling at a large distance. Strong coupling at a large distance is tight coupling. Weak coupling at a small distance is low cohesion: unrelated code in one place. If a part almost never changes, an imbalance is acceptable.

### Distance depends on the level

Measure distance at the level that you design. In our modular monolith, the boundary between two modules is the largest distance. So two modules inside one Go service can have tight coupling, even without microservices.

### Generic parts: strict contracts

Some parts solve a solved problem: the LLM, payments, notifications, S3. Their function rarely changes, but the provider can change, or we can use two providers. Hide each such provider behind a strict interface (port) in the core, with one adapter for each provider. A provider that is very expensive to change in any case (for example, Postgres) does not need this extra layer.

### What this means for us

- Product modules (volatile) depend on the core (stable), not the opposite ([layers point down](../architecture.md#rules-for-the-code-structure)).
- Between modules, use only the contract level: the service interface of the module and its types.
- Call another module synchronously when you need the result at once or a shared transaction. For a reaction "some time later", use an event from the outbox ([04](04-async-integrations.md)).
- Do not put the same business rule into the backend and the frontend. Two copies drift apart when the spec changes. The backend owns the rule, and the frontend shows its result.
- If a change makes you edit several modules at the same time, the boundary is in the wrong place.

## How it is checked

**Automatic checks in CI:**

| Check | What it catches | Blocks the merge |
| --- | --- | --- |
| Import rules (`core/tools/archcheck`) | The core imports a product module, or a module imports the internal packages of another module | Yes |
| Table ownership (`core/tools/archcheck`) | The SQL of a module uses the tables of another module. Each module README has a line `Tables: a, b`. The check parses the sqlc queries in `queries/*.sql` with the Postgres parser (pg_query, the same one sqlc uses) and compares the tables | Yes |
| RLS on each table | A table with `account_id` has no RLS rule, or the rule is not `FORCE` ([ADR 0009](../adr/0009-account-isolation-rls.md)). An integration test reads the Postgres system tables | Yes |
| Coupling report (`core/tools/archcheck`) | The number of links per module from the Go import graph: K̄ (the mean) and hotspots (modules above "mean + spread"). The target for K̄ is 1–3. Source of the metric: [Archiplan](https://github.com/archiplan-ai/Archiplan) (MIT) | No, a report in the PR |

**Where the code of a module is.** The checks use this folder layout:

| Folder | What is in it | Rules |
| --- | --- | --- |
| `core/internal/<module>` | A core module, for example `catalog` | Must not import `internal/product/*` |
| `core/internal/product/<module>` | A product module, for example `reports` | Can import the core |
| `core/internal/platform/<name>` | Shared infrastructure, for example `health`, `db` | A module of its own |
| `<module>/<subpackage>` | Private code of the module, for example `store` | Only the same module imports it. Other modules import the module root |
| `<module>/queries/*.sql` | The sqlc queries of the module | Use only the tables in the module README |

Run all checks with `task lint`. CI runs the same command. The Go linters are in `core/.golangci.yml`. The frontend checks are oxlint (`web/.oxlintrc.json`, with the rule that `web/packages/**` must not import an app), oxfmt and tsgo.

**Coupling review** for a new module and for a large boundary change. The agent does it with this procedure:

1. Read the spec, the module READMEs and the code of the integration points. Do not judge from the folder structure alone.
2. Write a short summary: the modules, their links, and a guess for the volatility of each module, with the confidence. The owner confirms or corrects it.
3. For each link between modules, record: what knowledge it shares, its strength level, explicit or implicit, its distance (sync or async), and the volatility.
4. Flag only links that are unbalanced **and** volatile. Do not flag everything. Do not recommend splitting modules by default: a split increases distance and needs weak coupling first.
5. Put the result in the PR. Each finding names the axis and the rule that it breaks.

The plugin [vladikk/modularity](https://github.com/vladikk/modularity) automates this review. ⚠️ Its license is CC BY-NC-SA, so a commercial project needs the author's permission first. Until then, the agent uses the procedure above.
