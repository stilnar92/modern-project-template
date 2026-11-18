# web (frontends)

The rules for the frontends. The full standard: [docs/dev/frontend-standard](../docs/dev/frontend-standard/README.md).

## Commands (from the repo root)

- `mise run lint:web`: oxlint, oxfmt and TypeScript 7 (`tsc`, `astro check` for the site). `pnpm format` (in `web/`) fixes the format.
- TypeScript 7 has no classic JavaScript API. `packages/api-client` keeps TypeScript 6 for `openapi-typescript`, and its `typecheck` runs the root `tsc` 7. `apps/site` keeps TypeScript 6 for `@astrojs/check`.
- `mise run test:web`: unit tests (Vitest). `mise run test:e2e`: Playwright, after `mise run up`. It starts the app dev servers itself. Install the browsers once: `pnpm --filter @app/e2e exec playwright install chromium webkit`.
- `mise run build:web`: build all apps. CI runs it.
- `pnpm dev` (in `web/`): the dashboard on <http://localhost:5173> and the public pages on <http://localhost:4330>. `CORE_URL` selects the core, for example `CORE_URL=http://127.0.0.1:4010` for the Prism mock. `mise run dev` starts both apps with the ports and `CORE_URL` from `.env`, so each git worktree has its own ports.
- `mise run gen:web`: regenerate `packages/api-client/src/schema.d.ts` (types, openapi-typescript) and `packages/api-client/src/gen/zod.gen.ts` (zod schemas, hey-api) from `api/openapi.yaml`.
- `mise run mock`: a mock API from the contract on `MOCK_PORT` from `.env` (Prism, 4010 in the main checkout).

## Layout

- `apps/dashboard`: the dashboard (Vite, React, Mantine). `apps/site`: the public pages (Astro, server rendering, no JavaScript in the browser). `packages/*`: shared code. `e2e`: Playwright tests.
- `packages/tokens`: the design tokens. `packages/i18n`: the UI texts (Paraglide). A shared package never imports an app or Mantine. The site never imports the dashboard or Mantine. oxlint blocks it.
- When Astro detects an AI agent, `astro dev` starts in the background and returns. Stop it with `npx astro dev stop` in `apps/site`.

## Rules

- Call the core only through `@app/api-client`. Do not write request types by hand.
- Do not edit `schema.d.ts` or `packages/api-client/src/gen/`. Change the contract and run `mise run gen`.
- Errors ([standard 27](../docs/dev/frontend-standard/27-error-handling.md#in-this-project)): call the API with `unwrap(await api.GET(...))`. A form uses `@mantine/form` with a zod schema from `@app/api-client` (`schemaResolver`), sets `meta: { handlesErrors: true }` on its mutation, and calls `applyProblem(form, error)`. Branch on `problemKind`, never on the text.
- The dashboard uses Mantine ([ADR 0008](../docs/dev/adr/0008-dashboard-mantine.md)).
- Design values come only from `packages/tokens` ([standard 19](../docs/dev/frontend-standard/19-design-tokens.md#in-this-project)). Change `packages/tokens/src/tokens.ts` and run `mise run gen`. Do not edit `packages/tokens/css/`. The site styles with Tailwind classes through the `tv()` variants in `apps/site/src/styles/ui.ts`. Both apps follow the system color scheme. A site page selects a palette from `palettes` with `<html data-palette="<id>">`. After a token change, republish the Claude Design view of the tokens ([ADR 0015](../docs/dev/adr/0015-design-system-and-claude-design.md)).
- Follow the [UX standard](../docs/dev/ux-standard/README.md) (U1–U35) for every change that a user sees: states, forms, errors, focus, sizes, text. A UI change is done only after the browser check in its "Verification" section.
- UI texts live in `packages/i18n/messages/ru.json`, not in components ([standard 28](../docs/dev/frontend-standard/28-i18n.md#in-this-project)). Call `m.key()` from `@app/i18n`. After a change of the messages, run `mise run gen`.
- A new dashboard page is a lazy child route in `apps/dashboard/src/routes.tsx` with a case in `routes.test.tsx` ([standard 25](../docs/dev/frontend-standard/25-routing.md)). Filters, tabs and the list cursor go in the URL.
- Do not copy a business rule from the backend. The backend owns the rule, the frontend shows its result.
- Guard every button that sends a change (create, delete, pay, submit, toggle) against a double tap with a `useRef` flag, not only `disabled={isPending}`: the state updates after a re-render, and two fast taps send two requests. Check: a double click sends one request. Details: [24-data-api-layer](../docs/dev/frontend-standard/24-data-api-layer.md).
- pnpm only. Commit `pnpm-lock.yaml`.
- Browser support: Baseline Widely available (`browserslist` in `package.json`). Use a newer feature only with the fallback from its guide. No polyfill by default ([10-browser-support](../docs/dev/frontend-standard/10-browser-support.md)).
- Before you write or change HTML, CSS or client code, look up the guide with the `modern-web-guidance` skill. `pnpm legacy-scan` (part of `lint:web`) blocks known legacy patterns. An exception goes to `legacy-scan.allow.json` with a reason.
- A key page must fit the real viewports in `e2e/viewports.ts`: DevTools docked, vertical tabs, 150% zoom, phone in landscape with the keyboard. Add each new key page to `e2e/tests/viewports.spec.ts`. For a manual check, use [Site Seeing](https://lab.ishadeed.com/tools/site-seeing/) with a localhost URL.
