import { Container, Title } from "@mantine/core";
import { HealthStatus } from "./features/health/HealthStatus";
import { useHealth } from "./features/health/useHealth";

// The app shell: no screens yet. The page only shows that the core answers.
export function HomePage() {
  const health = useHealth();
  return (
    <Container py="xl">
      <Title order={1}>Dashboard</Title>
      <HealthStatus state={health} />
    </Container>
  );
}
