import node from "@astrojs/node";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, envField } from "astro/config";

// One server renders all public pages (server-side rendering).
// The shell has one page that shows the core status.
export default defineConfig({
  output: "server",
  adapter: node({ mode: "standalone" }),
  // Not the Astro default 4321, so another local Astro app can use it.
  // Each git worktree has its own port in .env (mise run dev).
  // Tailwind 4 runs as a Vite plugin: it builds the CSS, no browser script.
  vite: { plugins: [tailwindcss()] },
  server: { port: Number(process.env.SITE_PORT ?? 4330) },
  env: {
    schema: {
      // The core API. "secret" makes Astro read the value when the server
      // starts, not when it builds. The value itself is not a secret.
      CORE_URL: envField.string({
        context: "server",
        access: "secret",
        default: "http://api.localhost:8088",
      }),
    },
  },
});
