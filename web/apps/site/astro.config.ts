import node from "@astrojs/node";
import { defineConfig, envField } from "astro/config";

// One server renders all public pages (server-side rendering).
// The shell has one page that shows the core status.
export default defineConfig({
  output: "server",
  adapter: node({ mode: "standalone" }),
  // Not the Astro default 4321, so another local Astro app can use it.
  server: { port: 4330 },
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
