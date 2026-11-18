// End-to-end tests (backend standard 07, frontend standard 29).
// They run against a started environment: "mise run up", then "mise run test:e2e".
// The addresses come from urls.ts. E2E_API_URL selects the core, for example the Prism mock.
// Playwright starts the dev servers of the apps and points them at the same core.
import { defineConfig, devices } from "@playwright/test";
import { dashboardUrl, coreUrl, siteUrl } from "./urls";
import { viewportProjects } from "./viewports";

const apps = /apps\/.*\.spec\.ts/;

export default defineConfig({
  testDir: "tests",
  timeout: 10_000,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? "github" : "list",
  projects: [
    // The API project needs no browser.
    { name: "api", testMatch: "healthz.spec.ts", use: { baseURL: coreUrl } },
    { name: "chromium", testMatch: apps, use: { ...devices["Desktop Chrome"] } },
    // Most users use phones, mostly iPhones (frontend standard D8, D9).
    { name: "webkit", testMatch: apps, use: { ...devices["iPhone 13"] } },
    // The key pages in real viewports: DevTools, vertical tabs, zoom, keyboard.
    ...viewportProjects,
  ],
  webServer: [
    {
      command: "pnpm --filter @app/dashboard dev",
      url: dashboardUrl,
      reuseExistingServer: !process.env.CI,
      env: { CORE_URL: coreUrl },
    },
    {
      command: "pnpm --filter @app/site dev",
      url: siteUrl,
      reuseExistingServer: !process.env.CI,
      // When Astro detects an AI agent, "astro dev" starts in the background
      // and exits. Playwright must own the server process, so keep it in front.
      env: { CORE_URL: coreUrl, ASTRO_DEV_BACKGROUND: "0" },
    },
  ],
});
