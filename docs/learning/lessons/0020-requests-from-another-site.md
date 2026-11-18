# 0020 — Another site makes the browser send a request with our cookie: CSRF

From PR #106 · 12.11.2026 · Topics: security, CSRF, cookies

## Problem

The session is in a cookie ([lesson 0018](0018-sessions-or-tokens.md)). The browser adds a cookie to each request to our domain, also when a page of another site starts the request. A user is logged in and opens a bad page. That page can try to change a price or delete a client in the name of the user.

## Predict first

A page on `evil.example` contains a form with `method="POST"` to our API, and a script submits it. The user is logged in to our site in the same browser. Which parts of the request can the other site control, and which parts does only the browser set?

## Theory

This attack is **cross-site request forgery (CSRF)**. The other site cannot read our cookie and cannot read our response. It does not need to. It only needs the browser to send a request that changes something. The server sees a valid session and does the work.

There are three families of defence:

- **A secret token.** The server gives the page a random value. Each changing request must send it back in a form field or a header. Another site cannot read it. The CSRF middleware of Django works this way. In a single-page app the client must get the token and add it to each request.
- **The `SameSite` cookie attribute.** With `Lax`, the browser does not send the cookie with a cross-site POST. It still sends it when the user follows a link to our site. But "site" is wider than "origin": `a.example.com` and `b.example.com` are the same site.
- **Headers that only the browser sets.** `Origin` gives the origin of the page that started the request. `Sec-Fetch-Site` says `same-origin`, `same-site`, `cross-site` or `none`. JavaScript on a page cannot change these headers. The server refuses a changing request that is not from its own origin.

All three need one rule: GET must not change data. The defences check only the methods that change state.

Sources:

- [OWASP Cross-Site Request Forgery Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html): all the defences and their limits.
- [Sec-Fetch-Site](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Sec-Fetch-Site), MDN: the header and its values.
- [Cross-Site Request Forgery](https://words.filippo.io/csrf/), Filippo Valsorda: the reasoning behind the check in the Go standard library.
- [net/http, type CrossOriginProtection](https://pkg.go.dev/net/http#CrossOriginProtection): the exact rules of that check.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Only `SameSite=Lax` | The browser does not send the cookie with a cross-site POST | One layer. It does not cover a request from a sibling subdomain |
| gorilla/csrf | A secret token in each changing request | The single-page app must carry the token. The library had a CVE |
| filippo.io/csrf | The header check, as a separate module | A dependency for what the standard library now has |
| `http.CrossOriginProtection` with `SameSite=Lax` | The server checks `Sec-Fetch-Site`, or compares `Origin` with `Host`. The cookie attribute is the second layer | It needs a browser that sends these headers |

## Edge cases

| What if | What happens |
| --- | --- |
| A page on another site sends a POST to the API | The browser does not add the `Lax` cookie. The server also sees `Sec-Fetch-Site: cross-site` and answers 403 in the Problem format |
| The request has no `Sec-Fetch-Site` header | The server compares the host of `Origin` with `Host` |
| The request has no `Sec-Fetch-Site` and no `Origin` | The server lets it pass: it is not a request from a current browser. A script such as `curl` has no cookie of a victim |
| A payment or bot service calls a webhook | It is a server, not a browser. `/v1/webhooks/` skips the check, and the signature of the webhook protects it |
| A GET changes data | No defence checks it. Safe methods always pass, so a GET must only read |
| The user follows a link from another site to the dashboard | It is a GET, and `Lax` sends the cookie. The user stays logged in |

## Decision

We use `http.CrossOriginProtection` from the Go standard library around all routes, with `SameSite=Lax` on the session cookie as a second layer. It adds no dependency and no token for the frontend. The deny handler answers 403 in the Problem format, like each other error of the API.

This would be wrong if we must support browsers that send neither header, or if a frontend on another origin must call the API with the cookie. The second case needs an explicit list of trusted origins.

Record: [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md), [backend standard 05](../../dev/backend-standard/05-security.md).

## In our code

1. `core/cmd/api/main.go`, `crossOrigin`: the check, the webhook exception and the deny handler.
2. `core/cmd/api/main.go`, `newHandler`: the check is inside the recovery and log layers ([lesson 0015](0015-layers-around-a-request.md)).
3. `core/internal/accounts/http.go`, `SetCookie`: `HttpOnly` and `SameSite=Lax`.
4. `core/cmd/api/main_test.go`, `TestHandler`: a cross-site POST gets a 403.

## Check yourself

1. A page on another site sends a form with POST to our API. What stops it, and in what order?
2. Why do webhooks skip the check, and why is that safe?
3. A developer adds `GET /v1/dashboard/clients/delete?id=...`. Why is this route open to CSRF?

<details><summary>Answers</summary>

1. First the browser: the cookie is `SameSite=Lax`, so a cross-site POST goes without it. Then the server: `CrossOriginProtection` sees `Sec-Fetch-Site: cross-site` and answers 403 before a handler runs.
2. A webhook comes from a server of another service. It has no `Sec-Fetch-Site` from a browser of a victim and no session cookie, so CSRF does not apply. Its signature proves who sent it.
3. The header check lets each GET pass, and `Lax` sends the cookie when the user follows a link. A link on another site is then sufficient to delete a client. A route that changes data must use POST, PUT, PATCH or DELETE.

</details>
