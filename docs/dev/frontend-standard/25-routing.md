# 25 — Routing & Navigation

**Status:** DRAFT — the rules of this project. The dashboard uses them now.

---

## Router

- The dashboard uses react-router in data mode (`createBrowserRouter`, [ADR 0012](../adr/0012-infrastructure-libraries.md)). Not framework mode: the dashboard is a single-page app and needs no file routes or build plugin. Not declarative `<Routes>`: it has no loaders and no middleware, so a guard runs after the first render and the page flickers.
- The site uses the file routes of Astro (`src/pages`). The rules below are for the dashboard.

## Route tree

- The tree is in `apps/dashboard/src/routes.tsx`. One root route is the layout of the dashboard: it renders `<Outlet />` and has `RouteError` as its `ErrorBoundary` ([27](27-error-handling.md#in-this-project)). An unknown address shows the "not found" page of the layout.
- Each page is a child route with `lazy`, so its code is a separate chunk: `lazy: { Component: async () => (await import("./features/x/XPage")).XPage }`.
- A feature that has several pages gets a child route with its own layout. Give it an `ErrorBoundary` when an error must not hide the rest of the dashboard.
- Paths are lowercase, with hyphens, and use the words of the product in English: `/services`, `/services/:serviceId`. A path parameter ends with `Id`.

## Data

- A route that needs data before it renders uses a loader that calls `queryClient.ensureQueryData(...)` with the same query as the component. So the data loads once, and the component reads it from the cache.
- Mutations stay in components (React Query). After a mutation, invalidate the queries. Do not reload the route.

## URL as state

- Filters, the sort order, the selected tab and the list cursor are search parameters (`useSearchParams`), so a link, the back button and a reload keep them (UX standard U6).
- The list cursor goes in `?cursor=`, exactly as the core returns it ([backend standard 02](../backend-standard/02-api.md#lists)).
- Do not put personal data in the URL: it goes into the history, the logs and the referrer.

## Guard

The guard needs a session endpoint in the contract. It comes with the login flow. The pattern:

- One layout route around the protected pages has a `middleware` (or a loader) that calls `queryClient.ensureQueryData(sessionQuery)`.
- A 401 throws `redirect("/login?next=" + encodeURIComponent(path))`. The login page returns to `next` only when it is a path of the dashboard, never another origin.
- The session cookie is `HttpOnly`, so the dashboard asks the core and never reads the cookie. The core checks each request again: the guard is for the user, not for security.

## Check

- `apps/dashboard/src/routes.test.tsx` renders the real route tree at a path. Add a case for each new page and each guard.
- A key page goes into `web/e2e/tests/viewports.spec.ts` ([web rules](../../../web/CLAUDE.md)).
