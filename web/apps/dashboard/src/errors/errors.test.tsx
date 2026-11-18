import { ApiError } from "@app/api-client";
import { MantineProvider } from "@mantine/core";
import { useForm } from "@mantine/form";
import { Notifications, notifications } from "@mantine/notifications";
import { act, render, renderHook, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { afterEach, expect, test } from "vitest";
import { applyProblem } from "./applyProblem";
import { notifyProblem } from "./notifyProblem";
import { RouteError } from "./RouteError";

const problem = (status: number, extra: object = {}) =>
  new ApiError({ type: "about:blank", title: "x", status, ...extra }, status);

function withNotifications() {
  render(
    <MantineProvider env="test">
      <Notifications />
    </MantineProvider>,
  );
}

afterEach(() => act(() => notifications.clean()));

test("field errors go next to the fields, and the values stay", () => {
  const { result } = renderHook(() => useForm({ initialValues: { price: "0" } }));
  let handled = false;
  act(() => {
    handled = applyProblem(
      result.current,
      problem(422, { errors: [{ pointer: "/price", detail: "Цена должна быть больше 0" }] }),
    );
  });
  expect(handled).toBe(true);
  expect(result.current.errors).toEqual({ price: "Цена должна быть больше 0" });
  expect(result.current.getValues()).toEqual({ price: "0" });
});

test("a refusal by a rule shows the server text in a notification, not on a field", async () => {
  withNotifications();
  const { result } = renderHook(() => useForm({ initialValues: { price: "0" } }));
  let handled = true;
  act(() => {
    handled = applyProblem(result.current, problem(409, { detail: "Слот уже занят" }));
  });
  expect(handled).toBe(false);
  expect(result.current.errors).toEqual({});
  expect(await screen.findByText("Слот уже занят")).toBeInTheDocument();
});

test("a network failure is a service message", async () => {
  withNotifications();
  act(() => notifyProblem(new TypeError("Failed to fetch")));
  expect(await screen.findByText("Сервис временно недоступен")).toBeInTheDocument();
});

test("a route error shows a page with a next step, not a blank screen", async () => {
  const router = createMemoryRouter([
    {
      path: "/",
      ErrorBoundary: RouteError,
      loader: () => {
        throw problem(503);
      },
      Component: () => null,
    },
  ]);
  render(
    <MantineProvider env="test">
      <RouterProvider router={router} />
    </MantineProvider>,
  );
  expect(await screen.findByRole("alert")).toHaveTextContent("Сервис временно недоступен");
  expect(screen.getByRole("button", { name: "Обновить страницу" })).toBeInTheDocument();
});

test("an unknown address is 'not found' with a link home, not a service failure", async () => {
  const router = createMemoryRouter(
    [{ path: "/", ErrorBoundary: RouteError, children: [{ index: true, Component: () => null }] }],
    { initialEntries: ["/nothing-here"] },
  );
  render(
    <MantineProvider env="test">
      <RouterProvider router={router} />
    </MantineProvider>,
  );
  expect(await screen.findByRole("alert")).toHaveTextContent("Не найдено");
  expect(screen.getByRole("link", { name: "На главную" })).toHaveAttribute("href", "/");
});
