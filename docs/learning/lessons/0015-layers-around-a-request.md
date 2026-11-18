# 0015 — Every request needs the same things: the layers around a handler

From PR #105 · 12.11.2026 · Topics: HTTP middleware, error handling

## Problem

Each request to the API needs the same three things. It needs an ID, one log record, and a safe answer when the code fails. We do not want to write this in each handler. We also must decide what the client gets when a handler panics.

## Predict first

A handler panics. One layer writes the access log, and another layer catches the panic. Which of the two must be the outer one, if the log must show the status 500? What does the log show in the other order?

## Theory

**Middleware** is a function that takes a handler and returns a new handler. The new handler does something before the call, calls the inner handler, and does something after it. In Go the shape is `func(next http.Handler) http.Handler`. Express middleware and Django middleware are the same idea.

The layers make an onion. A request goes from the outer layer to the inner one, and the response goes back through the same layers in reverse. So the order is part of the behavior:

- A layer sees only what the layers inside it do. An outer log layer sees the final status. An inner log layer does not see what an outer layer changes.
- A layer that puts a value into the request context gives it only to the layers inside it.

Go has no exceptions. A function returns an error value, and the caller checks it. A **panic** is for a programming error, for example a nil pointer. It stops the function and goes up the call stack. `recover()` in a deferred function stops the panic, like a `try/catch` around the call.

The Go HTTP server already recovers a panic in a handler, so one bad request does not stop the process. But it only writes the stack to the error log and closes the connection. The client gets no response. A recovery layer of our own can send a normal 500 response.

One limit: `recover` works only in the same goroutine. A panic in a goroutine that the handler started stops the whole process.

Sources:

- [net/http, type Handler](https://pkg.go.dev/net/http#Handler): what the server does when a handler panics, and `ErrAbortHandler`.
- [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover), the Go blog: the mechanism.
- [Making and Using HTTP Middleware](https://www.alexedwards.net/blog/making-and-using-middleware), Alex Edwards: the pattern in Go and the flow of control through the layers.
- [Django: Middleware](https://docs.djangoproject.com/en/5.2/topics/http/middleware/): the same onion, with the order in a list.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Code in each handler | Each handler logs and recovers | Somebody forgets it. The handlers repeat the same code |
| chi/middleware with httplog | Ready-made layers | A panic gives a plain 500, not our Problem format |
| samber/slog-http | A ready-made log layer for `slog` | Extra dependencies |
| Our own middleware | About 80 lines: request ID, access log, recovery | We write and test it |

## Edge cases

| What if | What happens |
| --- | --- |
| A handler panics | The client gets a 500 Problem response. The panic text and the stack go to the log and to Sentry, never to the client |
| The recovery layer is outside the access log | The panic passes through the access log before something stops it. The log gets no record. The client still gets the 500 |
| The request ID layer is inside the access log | The access log record has no request ID |
| The panic value is `http.ErrAbortHandler` | The recovery layer panics again with the same value. The server uses this value to stop a response, so it is not an error |
| Docker calls `/healthz` each 3 seconds | These records go to the Debug level, so they do not fill the log. A 5xx answer of a system check stays at Info |
| A URL has personal data in the query string | The access log writes only the method, the path, the status, the size and the duration |

## Decision

We wrote our own middleware. The order, from the outside in, is `otelhttp → RequestID → AccessLog → Recover → routes`. The trace layer is the outer one, so each record has a trace ID. `RequestID` is outside the log and the recovery, so both have the ID. `Recover` is inside `AccessLog`, so the log sees the 500 of a panic. A library would save about 80 lines, but its panic answer is a plain 500, and our API answers each error in the Problem format.

This would be wrong if the set of layers grows large, for example compression, rate limits and CORS. Then a maintained library is cheaper than our own code. A new library for this block needs a new ADR.

Record: [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md).

## In our code

1. `core/internal/platform/httpmw/httpmw.go`: the package comment gives the order. Then read `RequestID`, `AccessLog` and `Recover`.
2. `core/cmd/api/main.go`, `newHandler`: the code wraps from the inside out, so the last line is the outer layer.
3. `core/internal/platform/httpmw/httpmw_test.go`, `TestRecover` and `TestRecoverPassesAbortHandler`: the proof that the panic text stays out of the response.
4. [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md): the options for this block.

## Check yourself

1. Somebody moves `Recover` outside `AccessLog`. What changes for the client, and what changes in the log, when a handler panics?
2. Why does our recovery layer exist, when the Go server already recovers a panic in a handler?
3. A handler starts a goroutine for slow work, and the goroutine panics. Does the recovery layer help? Why?

<details><summary>Answers</summary>

1. Nothing changes for the client: it still gets the 500 Problem response. The access log loses the record of that request, because the panic goes through `AccessLog` before `Recover` stops it.
2. The server only logs the stack and closes the connection, so the client gets no response. Our layer sends a 500 in the Problem format, adds the request ID to the log record, and sends the error to Sentry.
3. No. `recover` stops a panic only in its own goroutine. The new goroutine has no recovery layer above it, so the panic stops the whole process.

</details>
