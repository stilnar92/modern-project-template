# 0028 — The server returns an error: what the user sees

From PR #109 · 12.11.2026 · Topics: error handling, forms, routing

## Problem

A user saves a service, and the server answers with an error. The answer can mean many things: a wrong field, a business rule, an ended session, a server failure. The network can also fail with no answer at all. Each case needs a different place on the screen and a different next step for the user. Without one rule, each screen decides alone, and the same error looks different in each place.

## Predict first

A user opens an address that does not exist in a single-page app. No request goes to the API. The error page knows only how to read API errors. What does it show? What should it show?

## Theory

An error has two properties that decide its place in the interface: **who can fix it** and **where its cause is**.

- **Next to the field.** The user can fix it, and the cause is one input. The form keeps all values, and the message is at the input.
- **A notification.** An action failed, and the cause has no place on the page. The page is still correct, so it stays.
- **An error page.** The page itself cannot show: its data did not load, or the render failed. Nothing correct is left to keep.

The interface must select the place from machine-readable data: the HTTP status, the `type` of the error and the list of field errors. It must never read the text. Texts change, and they are in the language of the user. RFC 9457 gives one body format for this ([lesson 0011](0011-api-error-format.md)). A field error points to its input with a JSON Pointer (RFC 6901), for example `/items/0/name`.

Not each failure has that body. A network failure has no response. A proxy can answer with an HTML page. The router throws its own response for an unknown address. So the client needs **one function at the boundary** that turns any thrown value into one shape. All code after it handles only that shape.

A React error boundary catches errors of rendering only. It does not catch an error in an event handler or in asynchronous code. A router with data loading adds its own boundary for each route: it catches errors of the loader and of the render. A failed mutation is an event, so no boundary sees it. It needs a handler of its own, usually one global handler with a way to turn it off for one mutation.

Sources:

- [RFC 9457: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457.html): the body format and its extension members.
- [RFC 6901: JSON Pointer](https://www.rfc-editor.org/rfc/rfc6901.html): the path syntax and the `~0` and `~1` escapes.
- [React Router: Error Boundaries](https://reactrouter.com/how-to/error-boundary): which errors a route boundary catches.
- [React Query Error Handling](https://tkdodo.eu/blog/react-query-error-handling), Dominik Dorfmeister: global handlers in the query and mutation caches.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Each call handles its own error | A `try`/`catch` or an `onError` at each request | Each screen maps statuses again. A forgotten handler shows nothing |
| `react-error-boundary` around widgets | A boundary component around each part of the page | One more dependency. It is necessary only when one widget must fail alone |
| Route `ErrorBoundary`, one mapper and one global mutation handler | The router catches loader and render errors. One function maps each error to a kind. The mutation cache shows a notification | A form must turn the global handler off, or the user sees the error twice |

## Edge cases

| What if | What happens |
| --- | --- |
| The address does not exist | The router throws its own 404, which is not an API error. The first version showed "the service is temporarily unavailable". Now it maps to "not found", with a link to the home page and no "reload" button |
| The network fails | No response exists. The mapper returns a problem with status 0, and the kind is "service" |
| The error body is not in the format, for example the HTML page of a proxy | The schema check fails. The error keeps only the status |
| The status is 422, and the body has no field errors | The kind is "refused": a business rule, not a field. The notification shows the `detail` of the server, because the backend owns the rule and its text |
| A form mutation fails with field errors | The form sets `meta.handlesErrors`, so the global handler does nothing. Field errors go to the fields, and the values stay |
| A form mutation fails with an error that is not in a field | The form sends it to the same notification function |
| The field is in a list: `/items/0/name` | The pointer becomes the form path `items.0.name`. The escapes `~1` and `~0` become `/` and `~` |

## Decision

We use the route `ErrorBoundary` of react-router, one mapper in the API client, and one global mutation handler. The mapper gives six kinds: validation, refused, unauthenticated, forbidden, not found, service. Each kind has one title and one action text. The interface branches on the kind and never on the text.

The reasons: the router boundary is built in, and one place knows the error format. We did not take `react-error-boundary`.

This would be wrong for a page with independent widgets, where one failed widget must not hide the others. Then that widget needs a boundary of its own. A failed query is also not in this path: it shows its error in its own area of the page.

Record: [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md), "Error handling".

## In our code

1. `web/packages/api-client/src/problem.ts`: `toProblem`, `problemKind` and `fieldErrors`. Read `problem.test.ts` next to it for the cases.
2. `web/apps/dashboard/src/errors/RouteError.tsx`: the error page, and `routeProblem` for the 404 of the router.
3. `web/apps/dashboard/src/errors/applyProblem.ts` and `notifyProblem.ts`: the field path and the notification path.
4. `web/apps/dashboard/src/main.tsx`: the `MutationCache` handler and the `handlesErrors` switch.

## Check yourself

1. Why does a React error boundary not show the error of a failed "save" request?
2. A colleague wants to show `detail` from the server for each error, because it is simpler. For which kinds is that wrong, and why?
3. What does the user see if a form forgets to set `handlesErrors`, and the server returns field errors?

<details><summary>Answers</summary>

1. A boundary catches errors that happen during rendering. A "save" request starts in an event handler and fails later, in asynchronous code. React is not rendering at that moment, so no boundary receives the error. The mutation handler must show it.
2. It is wrong for each kind except "refused". A network failure and a proxy page have no `detail` at all. For the other kinds the interface needs its own text with a next step, in the language of the user. Only a refusal by a business rule shows the text of the server, because the backend owns that rule.
3. The global handler shows a notification with the general validation text. The form also puts the errors next to the fields. The user gets the same error twice, and the notification says less than the fields do.

</details>
