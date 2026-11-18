// hey-api generates only the zod schemas of the contract (ADR 0012). The
// types and the client stay openapi-typescript and openapi-fetch.
import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../../../api/openapi.yaml",
  output: { path: "src/gen", format: false, lint: false },
  // Only the schemas of components/schemas. The request and response
  // aliases repeat them, and forms do not need them.
  plugins: [{ name: "zod", requests: { enabled: false }, responses: { enabled: false } }],
});
