# 0014 — An old client must keep working after a deploy

From PR #89 · 11.11.2026 · Topics: API design, compatibility, quality gates

## Problem

The server and its clients do not update at the same moment. After a deploy of the server, a browser tab still runs the old frontend bundle. Backend rule B4 says that a change inside `/v1` must not break an old client. Until this PR, the rule was only text, and nothing checked it.

## Predict first

You rename the response field `price` to `amount`. You change the server and the frontend in one PR, and all tests pass. Who gets an error after the deploy?

## Theory

An API contract is a promise to code that you do not deploy. Old clients live longer than you think: an open browser tab, a cached bundle, a mobile application that the user did not update, a script of a partner. A **breaking change** is a change that makes a correct old client fail.

The useful question for each change is: "what does a client that knows only the old contract see?"

- **Safe (additive):** a new endpoint, a new optional request field, a new response field. An old client does not send the new field and ignores what it does not know.
- **Breaking:** remove or rename a field, change a type, make an optional request field required, add a required request field, remove an endpoint.
- **It depends:** a new value in a response enum. It is safe only if the client has a default branch for a value that it does not know.

When you must remove something, you do it in steps over time. This is **deprecation**: mark the field as deprecated, tell the clients the removal date, keep it until that date. The `Sunset` HTTP header (RFC 8594) carries that date in the response. The other way is a new version (`/v2`), where both versions work for some time.

Analogy: a major version in semver. Inside `/v1` you can make only "minor" changes. The check for breaking changes is the same idea as a TypeScript error when a library removes an export.

Sources:

- Winters, Manshreck and Wright, "Software Engineering at Google", [chapter 15, "Deprecation"](https://abseil.io/resources/swe-book/html/ch15.html): why removal is a planned process.
- [RFC 8594, The Sunset HTTP Header Field](https://www.rfc-editor.org/rfc/rfc8594): how a response tells the date when a resource goes away.
- [APIs as infrastructure: future-proofing Stripe with versioning](https://stripe.com/blog/api-versioning), Brandur Leach: how a large public API keeps old clients working.
- [oasdiff: breaking changes](https://github.com/oasdiff/oasdiff/blob/main/docs/BREAKING-CHANGES.md): what the tool reports as breaking.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Change the contract freely, deploy server and client together | No rule | Each client that is not updated in the same second can break |
| A rule in text and a review | A person reads the contract change | People miss a changed type in a large file |
| Additive-only inside a version, with an automatic check | A tool compares the new contract with the one on `main` and fails on a breaking change | Old fields stay for some time. A real removal needs the deprecation steps |
| A new version for each breaking change | `/v1` and `/v2` work together | Two implementations to keep and test |
| A version for each client by date, with a translation layer | Each client keeps the behavior of its date | A lot of code. It fits a large public API |

## Edge cases

| What if | What happens |
| --- | --- |
| A field changes its type, for example `status` from integer to string | oasdiff fails `mise run lint:api`. This was the deliberate test of the gate |
| The change adds only metadata: `servers`, tags, examples | It is not breaking. The check passes |
| We must remove or rename a field | Mark it `deprecated` in the contract and send the `Sunset` header with a date. Remove it on that date or later |
| The frontend needs a new error type | It is a new enum value in a response. The client must ignore a value that it does not know, so it branches with a default |
| The behavior changes and the contract does not | The check cannot see it. It compares two contract files, not two servers |
| CI has no copy of `main` | The check needs `origin/main`, so the CI job fetches the full history |
| The branch is old and `main` moved | The task compares with `origin/main`. `OASDIFF_BASE` selects another ref |

## Decision

Inside `/v1` we make only additive changes: new endpoints, new optional fields, and new values that the client can ignore. oasdiff compares `api/openapi.yaml` with the version on `main` in `mise run lint:api` and in CI. A warning fails too. A removal goes through deprecation with a `Sunset` date.

This rule has a cost that grows: old fields collect in the contract. A new version becomes the better tool when many removals wait, or when a concept changes its meaning. The check also proves only the shape of the contract. Tests must cover behavior.

Record: the section "Compatibility" in `docs/dev/backend-standard/02-api.md`. Who owns each API surface is in [ADR 0005](../../dev/adr/0005-api-surfaces-and-ownership.md).

## In our code

1. `docs/dev/backend-standard/02-api.md`, "Compatibility": the rule and the deprecation steps.
2. `mise.toml`, the task `lint:api`: `oasdiff breaking` against `origin/main`, with `--fail-on WARN`.
3. `api/openapi.yaml`: the contract that the check reads.
4. `.github/workflows/ci.yml`, the `repo` job: the full fetch that gives the check `origin/main`.

## Check yourself

1. You add a new optional field to a response. Why is this safe for an old client?
2. Server and frontend change in the same PR and deploy together. Why can a rename still break a user?
3. The breaking-change check is green. What can still break an old client?

<details><summary>Answers</summary>

1. An old client reads only the fields that it knows and ignores the others. Nothing that it needs changed. This is true only if the client does not reject fields that it does not know.
2. "Together" is not one moment. A browser tab that was open before the deploy runs the old bundle and reads the old field name from the new server. The same is true for each client that we do not deploy.
3. A change of behavior with the same contract: the same field with another meaning, another default, another order of a list. The tool compares two contract files and does not run the server. Tests of behavior must find this.

</details>
