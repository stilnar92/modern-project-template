# 0032 — The first screen must load fast as the application grows

From PR #112 · 12.11.2026 · Topics: routing, code splitting, performance

## Problem

The dashboard had one page, and its route was in the entry file. Many pages will come: services, the calendar, settings. If each page is a static import, the browser downloads the code of all pages before it shows the first one. The router also needs rules before the pages come: where the data loads, where an error shows, where a login check runs.

## Predict first

A page loads its code on demand, and then its component starts a data request. How many network steps happen one after the other before the user sees data? What must the router know to make some of them run at the same time?

## Theory

A bundler follows static imports from the entry file and puts all that code in one bundle. A dynamic `import()` is a split point: the bundler writes the imported module and its own imports to a separate file, a **chunk**. The browser downloads the chunk when the code calls `import()`.

**The route is the natural split point.** The user already expects a short wait at a navigation. Each page is a unit that the user opens or does not open. So the first load has the shell and one page, and it does not grow with each new page.

Splitting has costs:

- **A waterfall.** If the component loads first and requests its data in an effect, the steps run in sequence: chunk, render, request. A router that knows the route table before it renders can start the chunk and the data in parallel.
- **A wait at navigation.** The chunk needs time, so the router must keep the old page or show a pending state.
- **A chunk can fail to load.** The network can fail, or a new deploy removed the old file. The dynamic import rejects, and something must catch it.

This is why the router model matters. A router that is only components (`<Routes>` in JSX) learns the matched route during the render. It has no place for work before the render. A router with a route table as data matches the URL first. Then it can load the code, run loaders and guards, and render only with the result. An error in any of these steps goes to the error boundary of the route.

Sources:

- [Reduce JavaScript payloads with code splitting](https://web.dev/articles/reduce-javascript-payloads-with-code-splitting), web.dev: dynamic imports and chunks.
- [MDN: `import()`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/import): the semantics of a dynamic import.
- [React Router: Route Object](https://reactrouter.com/start/data/route-object), sections "lazy", "loader" and "middleware": what a data route can hold.
- [React Router: Picking a Mode](https://reactrouter.com/start/modes): the declarative, data and framework modes.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Declarative `<Routes>` | Routes are components | No loaders and no middleware. A guard runs after the first render, and the page flickers |
| Framework mode | File routes and a build plugin | The dashboard is a single-page app and needs none of them |
| Data mode, pages as static imports | One route table, one bundle | The first load grows with each page |
| Data mode, each page with `lazy` | One route table. Each page is a chunk | A wait at the first visit of a page |

## Edge cases

| What if | What happens |
| --- | --- |
| The address matches no route | The root route is the layout, and it has the `ErrorBoundary`. The user sees "not found" in the dashboard, not a blank screen |
| A page needs data before it renders | A loader calls `ensureQueryData` with the same query as the component. The data loads once, and the component reads it from the cache |
| A feature has several pages | It gets a child route with its own layout. It gets its own `ErrorBoundary` when its error must not hide the rest of the dashboard |
| A filter or a list position must stay after a reload | It is a search parameter in the URL. The list cursor goes in `?cursor=`, as the server returns it |
| Someone puts personal data in the URL | It goes into the history, the logs and the referrer. The rule forbids it |
| A page needs a login | The guard is not built yet: it needs a session endpoint in the contract. The pattern is a layout route with middleware that redirects to the login page with a `next` parameter |
| `next` points to another site | The login page returns to `next` only when it is a path of the dashboard |
| A new page breaks the route tree | A test renders the real tree in a memory router. Each new page adds a case |

## Decision

The dashboard uses react-router in data mode. The route tree is one file. The root route is the layout with the error boundary. Each page is a child route with `lazy`, so each page is a separate chunk. After the change, the build has a chunk for the home page of 8.95 kB.

We set the rule with one page, before the size is a problem. The rule costs one line for each route now. A later change of all routes costs more.

This would be wrong for a small set of pages that the user always opens together: the split then adds waits and saves nothing. It also does not help the site, which uses the file routes of Astro and sends no JavaScript.

The guard is for the user, not for security. The server checks each request again.

Record: [frontend standard 25](../../dev/frontend-standard/25-routing.md) and [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md).

## In our code

1. `web/apps/dashboard/src/routes.tsx`: the tree, the layout route and the `lazy` page.
2. `web/apps/dashboard/src/main.tsx`: `createBrowserRouter(routes)` and the providers around it.
3. `web/apps/dashboard/src/routes.test.tsx`: the real tree at `/` and at an unknown address.
4. `web/apps/dashboard/src/errors/RouteError.tsx`: the page that the error boundary shows ([lesson 0028](0028-server-errors-on-screen.md)).

## Check yourself

1. Why does a guard in a declarative `<Routes>` tree make the page flicker, and a guard in a data route does not?
2. A route has `lazy` and a loader. Why must the loader use the same query as the component?
3. The login page gets `?next=https://example.com/steal`. What happens if it redirects to `next` with no check?

<details><summary>Answers</summary>

1. In a declarative tree, the guard is a component. It runs during the render, so the protected page or an empty frame shows before the redirect. A data route runs its middleware or loader after the match and before the render. The redirect happens before anything shows.
2. The loader puts the data in the query cache. The component reads the cache with its query key. With the same query, the component finds the data and makes no request. With a different query, the data loads twice, and the component starts with a loading state.
3. The user logs in on the real site and then lands on a site of an attacker. The address began on the real domain, so the user trusts it. This is an open redirect. The login page must accept only a path of its own app.

</details>
