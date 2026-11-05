import type { createApiClient } from "@app/api-client";

type Api = ReturnType<typeof createApiClient>;
export type Health = "ok" | "unavailable";

// The page always renders. A core that does not answer is a state of the
// page, not an error (frontend standard 27).
export async function readHealth(api: Api): Promise<Health> {
  try {
    const { data } = await api.GET("/healthz");
    return data?.status === "ok" ? "ok" : "unavailable";
  } catch {
    return "unavailable";
  }
}
