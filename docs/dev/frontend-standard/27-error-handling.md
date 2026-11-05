# 27 — Error Handling & Resilience

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### Errors as values (Go-style, typed)

Source: [errore](https://errore.org/) (remorses).

- **Return errors, don't throw** — functions return union types `Error | T` instead of throwing exceptions; callers check with `instanceof Error` and TypeScript narrows the type.
- **Compile-time enforcement** — forgetting to handle an error is a compile error: the union can't be used without narrowing. Especially valuable for AI-generated code — no invisible failures from forgotten `try/catch`.
- **Self-documenting signatures** — the return type declares every possible error type; the contract is visible without reading the implementation.
- **Typed custom errors** — error classes carry structured data (e.g. `createTaggedError` with template interpolation: `"User $id not found"` → `err.id`).
- **Exhaustive matching** — `matchError` gives compile-time exhaustiveness over error types, like a discriminated-union switch.
- **Errors + null compose** — `Error | T | null` works naturally with `?.` and `??`.
- **Flat control flow** — expression-based handling instead of try-catch nesting; resource cleanup via `await using` / `AsyncDisposableStack` instead of nested try-finally.
- **Boundary adapters** — exception-throwing libraries are wrapped at the boundary (`errore.try()` or equivalent) so the rest of the codebase stays in the errors-as-values world.
- **Zero-dependency core** — the pattern needs no library: custom classes extending `Error` + union returns are enough; the package only adds conveniences.
- **Explicit discards** — intentionally ignoring an Error-typed return uses the `void` keyword, distinguishing deliberate choices from accidental omissions.

Open question for elaboration: adopt as the project-wide convention (hooks/services layer), or only at API/service boundaries while React error boundaries handle the render layer.

### Errors in concurrent promises

Source: [Effective Promise Concurrency](https://www.builder.io/blog/promises) (Steve Sewell, Builder.io). Dovetails with errors-as-values above.

- **`try/catch` around `Promise.all` is a trap** — promises rejecting *after* the first rejection become unhandled rejections; the catch only sees the first one.
- **Per-promise `.catch(e => e)` + `instanceof Error` checks** — each promise resolves to value-or-Error, then results are narrowed individually. This *is* the errors-as-values pattern applied to concurrency.
- **`Promise.allSettled`** when every outcome must be captured with full information (`status`/`value`/`reason`).
- **`AggregateError`** to combine multiple failures into one error with `.errors` detail — fits typed-error conventions above.
