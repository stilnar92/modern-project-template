# 0029 — The same validation rule in the server and in the form

From PR #109 · 12.11.2026 · Topics: validation, contract, code generation

## Problem

The server refuses a service name that is too long. The form must show the same limit before the user sends the data. So the rule exists in two places and in two languages. Someone changes the limit on the server, and the form still shows the old one.

## Predict first

The contract says that a name has a maximum of 100 characters. A developer writes the form schema by hand with `max(100)`. Three months later the contract changes to 60. Which checks fail? What does the user see?

## Theory

Validation exists twice for two different reasons.

- **The server is the trust boundary.** A client can skip each check in the browser, so the server must check all input. This check is about safety and correct data.
- **The form gives fast feedback.** The user sees the problem at the input, with no request. This check is about the experience only.

Two copies of one rule drift. Three models keep one source:

- **Only the server validates.** The form sends the data and shows the errors of the server. No drift is possible, but each mistake costs a request.
- **Shared code.** The server and the client import one schema. This works only when both use the same language.
- **A contract with generation.** A neutral description (OpenAPI, JSON Schema) holds the rules. A generator writes the validator for each side ([lesson 0008](0008-contract-first-codegen.md)).

A contract can express the shape of the data: types, required fields, lengths, ranges, patterns and enums. It cannot express a rule that needs other data: "the name is unique", "the time slot is free". Those rules stay on the server. Their errors come back in the response, and the form must show them too ([lesson 0028](0028-server-errors-on-screen.md)).

A schema library such as zod parses: it takes `unknown` and returns a typed value or a list of errors with paths. So one schema can check a form and also check a response body that the client must not trust.

Sources:

- [OWASP Input Validation Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html), section "Common Pitfalls", "Client-only checks": why the server must check again.
- [Zod documentation](https://zod.dev/): schemas, `safeParse` and error paths.
- [Hey API: the Zod plugin](https://heyapi.dev/openapi-ts/plugins/zod): zod schemas from an OpenAPI file.
- [Mantine: schema-based validation](https://mantine.dev/form/schema-validation/): how a form uses a schema.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| zod schemas by hand | A developer writes each schema next to the form | The rules are a copy of the contract. Nothing fails when the contract changes |
| valibot | The same model with a smaller library | Agents mix its API with the API of zod |
| zod schemas generated from the contract | A generator reads `api/openapi.yaml` and writes one schema for each contract schema | One more generator, and it is before version 1.0 |

## Edge cases

| What if | What happens |
| --- | --- |
| The API of the generator changes | We pin the version and use only its zod plugin. The types and the client stay with another tool |
| The generator also writes request and response aliases | They repeat the schemas, and knip reported them as duplicate exports. The config turns them off |
| The linter, the formatter or the spell check reads the generated file | They skip it. `.gitattributes` marks it as generated |
| A rule is not in the contract, for example "the name is unique" | The form accepts the value. The server refuses it, and the error goes to the field or to a notification |
| The server sends an error body that is not in the error format | The client checks the body with the generated `zProblem` schema. The check fails, and the error keeps only the status |
| The contract changes | `mise run gen:web` writes the types and the schemas again |

## Decision

We generate zod 4 schemas from the contract with `@hey-api/openapi-ts`, and we use them through `@mantine/form`. The reason: the rules come from the contract, and nobody copies them by hand. The generator writes only the schemas of `components/schemas`. openapi-typescript and openapi-fetch still give the types and the client.

The server still checks each request. The generated schema only moves the same answer to an earlier moment.

This would be wrong if a form needs many rules that the contract cannot express. Then the form schema must extend the generated one by hand, and that part can drift again. It is also a risk while the generator is before 1.0: a new version can change the output.

No form uses the schemas yet. The first use is the check of the error body.

Record: [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md), "Form validation", and its stress round.

## In our code

1. `api/openapi.yaml`: the source of the rules.
2. `web/packages/api-client/openapi-ts.config.ts`: the zod plugin only, with the aliases turned off.
3. `web/packages/api-client/src/gen/zod.gen.ts`: the output. Do not edit it.
4. `web/packages/api-client/src/problem.ts`, `ApiError`: `zProblem.safeParse` checks an error body.

## Check yourself

1. The form already checks the input with the same rules as the server. Why must the server check it again?
2. A rule says "a user cannot have two services with the same name". Where does this rule live, and how does the user see its error?
3. Why does the project not generate the client and the types with the same tool that generates the schemas?

<details><summary>Answers</summary>

1. The form runs in the browser, and the user controls the browser. A script or a changed page can send any body directly to the API. The check in the form helps the user. Only the check on the server protects the data.
2. The rule needs the other services in the database, so the contract cannot express it. It lives on the server. The server refuses the request, and the response carries the error. The client shows it next to the field or in a notification.
3. That generator is before version 1.0, and its API changes. The project limits the risk: it pins the version and takes only the zod output. The types and the client come from tools that the project already uses.

</details>
