# dashboard: the account dashboard

A shell of the dashboard: Vite, React, Mantine ([ADR 0008](../../../docs/dev/adr/0008-dashboard-mantine.md)), TanStack Query and React Router. React Compiler is part of the build. There are no screens yet: the page shows only that the core answers.

## Run

From `web/`: `pnpm dev` starts this app on http://localhost:5173 and the site. Or only this app: `pnpm --filter @app/dashboard dev`.

The browser calls the same origin. In development, Vite sends `/healthz` and `/v1` to the core:

| Variable            | Default                     | What it does                                                                                               |
| ------------------- | --------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `CORE_URL`          | `http://api.localhost:8088` | The core for the dev proxy. `task up` starts it. For the Prism mock (`task mock`): `http://127.0.0.1:4010` |
| `VITE_API_BASE_URL` | empty (same origin)         | The core URL in the browser, for a build that is not served next to the core                               |
| `VITE_SENTRY_DSN`   | empty                       | Error tracking and Web Vitals. Off when empty                                                              |

## Layout

- `src/features/<feature>/`: a hook with the logic (`useHealth`) and components that only show data (`HealthStatus`).
- `src/api.ts`: the one API client. Call the core only through it.

## Checks

`pnpm test` (Vitest and Testing Library), `pnpm typecheck` (TypeScript 7, `tsc`). The end-to-end test is `web/e2e/tests/apps/dashboard.spec.ts`.
