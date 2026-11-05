import { useQuery } from "@tanstack/react-query";
import { api } from "../../api";
import type { HealthState } from "./HealthStatus";

// The state of the core for the status line.
export function useHealth(): HealthState {
  const health = useQuery({
    queryKey: ["health"],
    queryFn: async () => {
      const { data } = await api.GET("/healthz");
      if (!data) throw new Error("The core did not answer");
      return data.status;
    },
    retry: false,
  });
  if (health.isPending) return "loading";
  return health.isSuccess ? "ok" : "unavailable";
}
