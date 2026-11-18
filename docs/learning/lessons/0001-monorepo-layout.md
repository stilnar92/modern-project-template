# 0001 — Where the backend, the frontends and the contract live: one repository or several

From PR #32 · 28.10.2026 · Topics: repository layout, tooling

## Problem

The product has a Go backend, some TypeScript frontends and an API contract that connects them. Agents write most of the code. Before the first line of code, we must decide where each part lives. Two of our documents described two different layouts, so the question had to be closed.

## Predict first

A change adds one field to an API response. The backend, the contract and one frontend must change. How many pull requests does it take with three repositories? What can go wrong between the first merge and the last?

## Theory

A **monorepo** is one repository that holds several projects. A **polyrepo** setup gives each project its own repository. The important difference is the unit of change:

- In a monorepo, one commit can change the contract, the server and the client together. They cannot be at different versions, because there is only one version.
- In a polyrepo, each repository has its own history. A change across repositories is several commits that you must release in the correct order.

A monorepo has costs. Each tool must know which part of the tree it owns. CI must not run all checks for each change. A person, or an agent, sees more code than the task needs.

One repository does not mean one toolchain. A Go **module** is a folder with a `go.mod` file, like a package with its `package.json`. A pnpm **workspace** is a set of packages under one lockfile. Each can own one folder in the same repository and ignore the other folders.

Go has one more guard. The compiler does not let code outside a module import a package below a folder named `internal/`. This is like a package that does not export a file, but the compiler does the check, not a convention.

Sources:

- Winters, Manshreck, Wright, "Software Engineering at Google", chapter 16 "Version Control and Branch Management": the argument for one repository and its costs.
- [Monorepo Explained](https://monorepo.tools/): what a monorepo is, what it is not, and the tooling it needs.
- [Organizing a Go module](https://go.dev/doc/modules/layout): the `cmd/` and `internal/` layout that the Go team recommends.
- [pnpm workspaces](https://pnpm.io/workspaces): how one lockfile serves several packages.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Several repositories | One repository each for the backend, the frontends and the contract | A contract change is several pull requests, and the parts can be at different versions |
| One repository, Go at the root | `cmd/` and `internal/` at the root, next to `web/` | Go files and Node files mix at the root. A rule file at the root applies to all the code |
| One repository, folders by stack | `core/` has the Go module, `web/` has the pnpm workspace, `api/` has the contract | The backend and the frontends are at different depths (`core/` and `web/apps/...`) |

## Edge cases

| What if | What happens |
| --- | --- |
| A change is only in `core/` | CI runs by path: the `web/` checks do not start, and the opposite |
| A change is in `api/openapi.yaml` | Both sets of checks start, because the contract belongs to both sides |
| An app is live and its deploy hook builds it from its folder | It stays in its folder with its own lockfile, outside the workspace. A move would put the live app at risk for no gain |
| An agent works in `web/` | It reads the root rule file and the `web/` rule file. It does not get the Go rules |
| Code outside the core tries to import a core package | The Go compiler refuses the import, because the packages are below `core/internal/` |
| One command must build or test both stacks | A task runner at the root calls `go` and `pnpm`. It was a Taskfile in this PR, and it is `mise.toml` now ([ADR 0010](../../dev/adr/0010-mise-single-toolchain.md)) |

## Decision

We use one repository with folders by stack: `core/` (one `go.mod`), `web/` (one pnpm workspace with `apps/` and `packages/`), `api/openapi.yaml`, and `deploy/`. The tools do not mix, the rule files are in layers, and CI runs by path. The contract is in its own folder, because neither side owns it ([lesson 0008](0008-contract-first-codegen.md)).

One repository was the starting point: everything was already there. The record holds the comparison of the two layouts, not a comparison with several repositories.

This would be wrong if separate teams needed separate access rights or separate release cycles. It would also be wrong if the repository became so large that path filters could not keep CI fast.

Record: [ADR 0007](../../dev/adr/0007-monorepo-layout.md).

## In our code

1. `core/go.mod`: the one Go module. Each folder in `core/cmd/` is one program.
2. `web/pnpm-workspace.yaml`: the workspace that holds `web/apps` and `web/packages`.
3. `mise.toml`: the shared commands that call `go` and `pnpm`.
4. [ADR 0007](../../dev/adr/0007-monorepo-layout.md): the layout and its consequences.

## Check yourself

1. Why does a change in `api/openapi.yaml` start both the core checks and the web checks, when a change in `core/` starts only one set?
2. A live app has its own folder and lockfile, and its deploy hook builds it from there. Why is it better to leave it there for now?
3. What does the folder name `internal/` give us that a note in a README does not?

<details><summary>Answers</summary>

1. The server code and the client code both come from the contract. A contract change can break either side, so both must prove that they still build and pass. A change in `core/` cannot change the web code.
2. The app is live. The deploy hook on the server builds it from its folder. A move gives no gain now and can stop the app. A separate decision moves or archives it later.
3. The compiler does the check. An import of a package below `internal/` from outside the module does not build. A note in a README depends on each author, person or agent, to read it and obey it.

</details>
