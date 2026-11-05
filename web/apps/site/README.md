# site: the public pages

A shell of the public pages: Astro with server rendering and the Node adapter. One server renders the public pages of all accounts ([architecture](../../../docs/dev/architecture.md)). There are no content pages yet: the page shows only that the core answers. The page sends no JavaScript to the browser.

## Run

From `web/`: `pnpm dev` starts this app on http://localhost:4330 and the dashboard. Or only this app: `pnpm --filter @app/site dev`.

The port is not the Astro default 4321, so it does not collide with other Astro apps on the same machine.

| Variable   | Default                     | What it does                                                       |
| ---------- | --------------------------- | ------------------------------------------------------------------ |
| `CORE_URL` | `http://api.localhost:8088` | The core. Astro reads it when the server starts, not at build time |

When Astro detects an AI agent, `astro dev` starts in the background and returns. Stop it with `npx astro dev stop`. Set `ASTRO_DEV_BACKGROUND=0` to keep it in the foreground.

## Production build

`pnpm build`, then `node dist/server/entry.mjs` (`PORT` selects the port).

## Checks

`pnpm test` (Vitest), `pnpm typecheck` (`astro check` with TypeScript 6: `tsc` cannot read `.astro` files, and `@astrojs/check` does not accept TypeScript 7 yet). The end-to-end test is `web/e2e/tests/apps/site.spec.ts`. It runs with JavaScript off, so it proves that the server renders the page.

Not in the shell yet: error tracking. The `@sentry/astro` integration adds a browser script with Session Replay to each page by default, and the public pages must stay free of scripts. The first feature that needs error tracking selects a server-only setup.
