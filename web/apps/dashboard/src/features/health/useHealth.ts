import { unwrap } from "@app/api-client";
import { useQuery } from "@tanstack/react-query";
import { api } from "../../api";
import type { HealthState } from "./HealthStatus";

// The state of the core for the status line.
export function useHealth(): HealthState {
  const health = useQuery({
    queryKey: ["health"],
    queryFn: async () => unwrap(await api.GET("/healthz")).status,
    retry: false,
  });
  if (health.isPending) return "loading";
  return health.isSuccess ? "ok" : "unavailable";
}
