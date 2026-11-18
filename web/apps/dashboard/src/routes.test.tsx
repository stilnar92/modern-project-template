import { MantineProvider } from "@mantine/core";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { expect, test } from "vitest";
import { routes } from "./routes";

function renderAt(path: string) {
  const router = createMemoryRouter(routes, { initialEntries: [path] });
  render(
    <MantineProvider env="test">
      <QueryClientProvider client={new QueryClient()}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </MantineProvider>,
  );
}

test("the home page loads on demand", async () => {
  renderAt("/");
  expect(await screen.findByRole("heading", { name: "Панель управления" })).toBeInTheDocument();
});

test("an unknown address shows the not-found page of the layout", async () => {
  renderAt("/nothing-here");
  expect(await screen.findByRole("alert")).toHaveTextContent("Не найдено");
});
