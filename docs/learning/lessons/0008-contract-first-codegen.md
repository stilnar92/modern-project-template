# 0008 — How the backend and the frontend stay in agreement about the API: one contract, generated code

From PR #47 · 28.10.2026 · Topics: API contract, code generation

## Problem

The backend is in Go and the frontends are in TypeScript. Each side needs the shape of each request and response. If each side writes the shapes by hand, they drift apart: the backend renames a field, and the frontend finds out in production. We had to decide where the true description of the API lives.

## Predict first

Somebody adds a new operation to the API description and writes no backend code for it. In a project where the description is only documentation, what happens? What must happen for this mistake to be impossible to merge?

## Theory

An **API contract** is a description of the API that a machine can read: the paths, the parameters, and the schemas of requests and responses. OpenAPI is the common format for HTTP APIs.

There are two directions:

- **Code first.** You write the server, and a tool makes the description from the code. Django REST Framework works this way. The description is always true for the server, but it exists only after the server code does.
- **Contract first.** You write the description, and tools make code from it. The two sides can agree on the contract and then work in parallel.

**Code generation** turns the contract into types on both sides. On the server, a generator makes an interface with one method for each operation. In Go, a type satisfies an interface when it has all its methods. There is no `implements` keyword, but the compiler checks the method list, like a TypeScript interface. If the contract has an operation and the code does not, the program does not build.

On the client, a generator makes TypeScript types for each path. A typed fetch client then knows which paths exist and what each returns. A wrong path is a type error, not a 404 in the browser.

One question stays: do you commit the generated files? If you commit them, the review shows them and the build does not need the generator. But they can go stale, so CI must generate again and compare.

Sources:

- [OpenAPI Specification](https://spec.openapis.org/oas/latest.html): the format of the contract.
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen): the Go generator, with its "strict server" mode.
- [openapi-typescript](https://openapi-ts.dev/introduction) and [openapi-fetch](https://openapi-ts.dev/openapi-fetch/): the TypeScript types and the typed client.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Types by hand on both sides | Each side writes its own types from a document | They drift apart, and nothing sees it |
| Code first | The server code is the source, and a tool makes the schema | The frontend waits for the backend. The contract changes as a side effect of code |
| Contract first, generate at build time | The generated files are not in git | The build needs the generator. A reviewer does not see what changed |
| Contract first, generated files in git | The files are committed, and CI generates again and compares | Larger diffs. A forgotten generation fails CI |

## Edge cases

| What if | What happens |
| --- | --- |
| The contract changes and nobody runs the generator | `gen:check` runs the generators and fails when the result differs from git. The CI output shows the diff |
| The contract has an operation that the code does not implement | The generated interface has a method that no type has. The Go build fails |
| Somebody edits a generated file by hand | The same check fails, because a fresh generation differs from the edited file |
| Two developers have different generator versions | The Go generator is a tool in `go.mod`, like a dev dependency in `package.json`, so each person runs the same version |
| The frontend needs an endpoint that the backend did not build yet | A mock server answers from the contract. The frontend work can start |
| A contract change breaks existing clients | A linter compares the contract with the main branch and fails on a breaking change |

## Decision

`api/openapi.yaml` is the one source. `mise run gen` makes the Go server code and the TypeScript client from it. We use the strict server: each handler gets a typed request and returns a typed response. The generated files are in git, and CI fails when they are stale. A change in `api/` runs the checks of both sides ([lesson 0001](0001-monorepo-layout.md)).

The generated Go code is a platform package. Modules implement its interface and do not edit it.

This would be wrong for an API with one small client and one developer who owns both sides, where the contract adds a step with no gain. It also has a limit: the types prove the shape of the data, not the behavior. A handler can return a valid shape with wrong values, so tests are still necessary.

There is no ADR for this decision. The API rules are in backend standard 02.

## In our code

1. `api/openapi.yaml`: the contract.
2. `core/internal/platform/openapi/generate.go`: the one `//go:generate` line. The result, `openapi.gen.go`, is next to it.
3. `core/internal/platform/health/health.go`, `GetHealth`: our code implements one method of the generated interface.
4. `web/packages/api-client/src/index.ts`: the typed client. Its types come from the generated `schema.d.ts`.

## Check yourself

1. You add `GET /v1/public/sites/{slug}` to the contract and run the generator. Why does the Go build fail, and why is that good?
2. Why do we commit generated files, and what protects us from a stale file?
3. The types on both sides match the contract. What can still be wrong between the backend and the frontend?

<details><summary>Answers</summary>

1. The generated interface now has a new method, and no type implements it. The build fails at the place where the server gets its handlers. This is good because the contract cannot promise an operation that the server does not have.
2. A reviewer can read them in the PR, and the build does not depend on the generator. CI runs the generators and compares the result with git. A difference fails the PR.
3. The behavior. The contract says that a field is a string, not which string is correct. A handler can return the wrong account, the wrong status code for a case, or an empty list. Tests and the isolation checks cover that part.

</details>
