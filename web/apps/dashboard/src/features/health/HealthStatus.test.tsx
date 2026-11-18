import { MantineProvider } from "@mantine/core";
import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import { HealthStatus } from "./HealthStatus";

function renderStatus(state: "loading" | "ok" | "unavailable") {
  render(
    <MantineProvider env="test">
      <HealthStatus state={state} />
    </MantineProvider>,
  );
  return screen.getByRole("status");
}

test("shows that the core is available", () => {
  expect(renderStatus("ok")).toHaveTextContent("Ядро: ok");
});

test("shows that the core is unavailable", () => {
  expect(renderStatus("unavailable")).toHaveTextContent("Ядро: недоступно");
});
