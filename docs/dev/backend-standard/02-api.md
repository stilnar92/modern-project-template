# 02 — API

**Status:** DRAFT
**Decisions:** B2, B3, B4

The general conventions (REST, `/v1`, plural resource names, cursor pagination) are in the [tech stack](../tech-stack.md#api-conventions). The API sections and how the server finds the account are in [ADR 0005](../adr/0005-api-surfaces-and-ownership.md). This section gives the implementation rules.

## Contract

- Each API change starts with `api/openapi.yaml`, then code generation, then the implementation. Do not write code for a contract that does not exist.
- Handlers implement the interface that oapi-codegen generates. Hand-written request and response types are forbidden.
- CI generates the code again and fails if the result differs from the committed code.

## Errors

The format is Problem Details, [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457), `Content-Type: application/problem+json`:

```json
{
  "type": "https://domain/errors/validation",
  "title": "Invalid data",
  "status": 422,
  "detail": "The quantity must be greater than zero",
  "errors": [{ "pointer": "/quantity", "detail": "must be greater than 0" }]
}
```

- `type` is a stable error ID. The frontend branches on it, not on the text.
- The contract keeps the list of `type` values. A new type is a contract change.
- Field validation errors go into the `errors` extension, with a JSON Pointer to the field.
- Internal details (SQL, stack traces, IDs of other accounts) do not go into the response. They go into the log with the request ID, and the response contains the request ID.

## Compatibility

- Inside `/v1`: only new endpoints, new optional fields, and new values that the client can ignore.
- To remove or rename a field: mark the field `deprecated` in the contract, and send the `Sunset` header with a date ([RFC 8594](https://www.rfc-editor.org/rfc/rfc8594)). Remove the field on that date or later.
- Source of this approach: the chapter on deprecation in [Software Engineering at Google](https://abseil.io/resources/swe-book).

## Other rules

- **Time:** RFC 3339 in UTC. Keep the account time zone separately and apply it for display.
- **Money:** an integer amount in minor units (for example, cents) and a currency code. No floating-point numbers.
- **Long operations** (AI tasks, import): return `202` with a task ID. Get the status with `GET` or SSE.
- **Repeats from the dashboard:** create requests that must not run twice (for example, a message to a user) accept the `Idempotency-Key` header.
- **Cache of public responses:** `Cache-Control` by the [rules of Jake Archibald](https://jakearchibald.com/2016/caching-best-practices/). Public HTML pages get a short cache that an event resets; hashed static files get a permanent cache.
