# web (frontends)

The rules for the frontends. The full standard: [docs/dev/frontend-standard](../docs/dev/frontend-standard/README.md).

## Commands (from the repo root)

- `task lint:web`: oxlint, oxfmt and TypeScript 7 (`tsc`, `astro check` for the site). `pnpm format` (in `web/`) fixes the format.
- TypeScript 7 has no classic JavaScript API. `packages/api-client` keeps TypeScript 6 for `openapi-typescript`, and its `typecheck` runs the root `tsc` 7. `apps/site` keeps TypeScript 6 for `@astrojs/check`.
- `task test:web`: unit tests (Vitest). `task test:e2e`: Playwright, after `task up`. It starts the app dev servers itself. Install the browsers once: `pnpm --filter @app/e2e exec playwright install chromium webkit`.
- `task build:web`: build all apps. CI runs it.
- `pnpm dev` (in `web/`): the dashboard on http://localhost:5173 and the public pages on http://localhost:4330. `CORE_URL` selects the core, for example `CORE_URL=http://127.0.0.1:4010` for the Prism mock.
- `task gen:web`: regenerate `packages/api-client/src/schema.d.ts` from `api/openapi.yaml`.
- `task mock`: a mock API from the contract on port 4010 (Prism).

## Layout

- `apps/dashboard`: the dashboard (Vite, React, Mantine). `apps/site`: the public pages (Astro, server rendering, no JavaScript in the browser). `packages/*`: shared code. `e2e`: Playwright tests.
- A shared package never imports an app or Mantine. The site never imports the dashboard or Mantine. oxlint blocks it.
- When Astro detects an AI agent, `astro dev` starts in the background and returns. Stop it with `npx astro dev stop` in `apps/site`.

## Rules

- Call the core only through `@app/api-client`. Do not write request types by hand.
- Do not edit `schema.d.ts`. Change the contract and run `task gen`.
- The dashboard uses Mantine ([ADR 0008](../docs/dev/adr/0008-dashboard-mantine.md)).
- Do not copy a business rule from the backend. The backend owns the rule, the frontend shows its result.
- pnpm only. Commit `pnpm-lock.yaml`.
