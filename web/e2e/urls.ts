// The addresses of the core and the dev servers. They come from .env
// (mise run test:e2e), so each git worktree tests its own stack.
// E2E_API_URL selects another core, for example the Prism mock.
export const coreUrl =
  process.env.E2E_API_URL ?? process.env.CORE_URL ?? "http://api.localhost:8088";
export const dashboardUrl = `http://localhost:${process.env.DASHBOARD_PORT ?? "5173"}`;
export const siteUrl = `http://localhost:${process.env.SITE_PORT ?? "4330"}`;
