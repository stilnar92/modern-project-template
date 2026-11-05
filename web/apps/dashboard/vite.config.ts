import babel from "@rolldown/plugin-babel";
import react, { reactCompilerPreset } from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

// In development the browser calls the same origin, and Vite sends /healthz
// and /v1 to the core. The core runs behind Caddy on api.localhost ("task
// up"). CORE_URL selects another target, for example the Prism mock.
// changeOrigin is necessary: Caddy selects the service by the Host header.
const core = { target: process.env.CORE_URL ?? "http://api.localhost:8088", changeOrigin: true };

export default defineConfig({
  // React Compiler is part of the build (frontend standard 09, 30).
  plugins: [react(), babel({ presets: [reactCompilerPreset()] })],
  server: {
    port: 5173,
    strictPort: true,
    proxy: { "/healthz": core, "/v1": core },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./vitest.setup.ts"],
  },
});
