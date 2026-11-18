import { m } from "@app/i18n";
import { Text } from "@mantine/core";

export type HealthState = "loading" | "ok" | "unavailable";

const LABELS: Record<HealthState, () => string> = {
  loading: m.core_status_checking,
  ok: m.core_status_ok,
  unavailable: m.core_status_unavailable,
};

export function HealthStatus({ state }: { state: HealthState }) {
  // <output> has the implicit role "status", so screen readers announce changes.
  return <Text component="output">{LABELS[state]()}</Text>;
}
