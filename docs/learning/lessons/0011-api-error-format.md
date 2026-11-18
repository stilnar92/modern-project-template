# 0011 — What the API tells a client when something fails: one error format

From PR #62 · 04.11.2026 · Topics: API design, security

## Problem

The API used the default error handlers of the generated server. They send the raw Go error text to the client as plain text. So the first handler that returns a database error would show SQL to the client. The frontend also had no stable field to branch on.

## Predict first

A handler returns the error `relation "accounts" does not exist`. The server sends this text in the response body. Who can read it, and what do they learn from it?

## Theory

An error response has two readers with opposite needs. The **client code** needs a stable, machine-readable answer: what type of error is this, and can the user fix it? The **developer** needs the full detail: the query, the stack, the IDs. If one text serves both, the client gets internal details. An attacker uses them to learn table names, library versions and, later, personal data.

The rule is to split the two. The full error goes to the log and to error tracking. The client gets a fixed, safe body.

**Problem Details** (RFC 9457) is a standard shape for that body. The media type is `application/problem+json`. The main members are:

- `type`: a URI that identifies the error type. Client code branches on it. The default `about:blank` means "a plain HTTP error with no more meaning than the status".
- `title`: a short text for people. It is the same for each occurrence of the type.
- `status`: the HTTP status code.
- `detail`: a text for people about this occurrence.

You can add your own members, for example a list of field errors. Analogy: an exception handler in Django REST Framework that gives all errors the same JSON shape, or one error middleware in Express.

Sources:

- [RFC 9457, Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457): the format. Section 3 defines the members.
- [Error Handling Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Error_Handling_Cheat_Sheet.html), OWASP: why a response must not show internal errors.
- [Improper Error Handling](https://owasp.org/www-community/Improper_Error_Handling), OWASP: what an attacker learns from error text.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| The default handlers of the generated code | They send `err.Error()` as `text/plain` | Internal text reaches the client. No stable field |
| Our own JSON, for example `{"error": "..."}` | One shape that we invent | Each client must learn it. No tools know it |
| Problem Details with free `type` values | The standard shape. Each handler selects a type | Types multiply, and the frontend cannot know the full list |
| Problem Details with the `type` list in the contract | The contract has an enum of types. A new type is a contract change | Each new error type needs a contract change and code generation |

## Edge cases

| What if | What happens |
| --- | --- |
| A handler returns an error | The client gets 500 with `type` `about:blank` and no detail. The error goes to the log and to Sentry |
| The server cannot read the request, for example the body is not JSON | The client gets 400 with a fixed detail. The parser error goes to the log only |
| The path has no route, or the method is wrong | The Go router answers 404 and 405 as plain text. A wrapper keeps the status and replaces the body with a Problem. 405 keeps the `Allow` header |
| Somebody sends `err.Error()` to the client again | A test calls the real handler chain with a fake SQL error and fails when the text is in the body |
| A request has invalid fields | The schema has the `errors` extension with a JSON Pointer for each field. Nothing produced it at the time of this PR |
| The frontend needs a new error type | Add it to the enum in the contract first. Then the generated Go and TypeScript types know it |

## Decision

We use Problem Details with the list of types in the contract. The package `problem` is the only place that writes an error response. A server error gives the client only the status. The real error goes to the log and to error tracking, and the request ID in the `X-Request-ID` header connects the two. The frontend branches on `type`, not on the text.

This would be wrong if `detail` gets raw error text "to help with debugging". The format does not protect anything alone. The protection is the rule that only safe, fixed text goes into `Write`.

Record: the section "Errors" in `docs/dev/backend-standard/02-api.md`.

## In our code

1. `api/openapi.yaml`, the schema `Problem`: the members and the `type` enum.
2. `core/internal/platform/problem/problem.go`: `Write`, `RequestError`, `ServerError` and `Routes`.
3. `core/cmd/api/main.go`, `newHandler`: where the generated server gets our error handlers.
4. `core/cmd/api/main_test.go`, `TestHandler`: the test of the real chain with a fake SQL error.

## Check yourself

1. Why does the frontend branch on `type` and not on `title` or `detail`?
2. A colleague wants to put the database error into `detail` on the test server "to debug faster". What do you answer?
3. Why does the leak test call the real handler chain, and not only the function `ServerError`?

<details><summary>Answers</summary>

1. `type` is a stable ID from a list in the contract. `title` and `detail` are texts for people: somebody can change the words or translate them, and then a comparison of strings breaks.
2. The same code runs in production, so the leak goes there too. The full error is already in the log and in error tracking, and the request ID finds it. That is the fast way to debug.
3. The defect was in the wiring: the server was created with no error handlers, so the generated defaults sent the raw text. A test of `ServerError` alone passes also when nobody connects it. The test of the real chain fails in that situation.

</details>
