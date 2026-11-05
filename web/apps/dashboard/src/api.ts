import { createApiClient } from "@app/api-client";

// An empty base URL sends the calls to the same origin. In development the
// Vite proxy sends them on to the core (vite.config.ts).
export const api = createApiClient(import.meta.env.VITE_API_BASE_URL ?? "");
