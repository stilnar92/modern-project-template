// The route tree of the dashboard (frontend standard 25). Data mode of
// react-router: each route can have a loader, middleware and an
// ErrorBoundary, and the router knows them before rendering.
import { Outlet, type RouteObject } from "react-router";
import { RouteError } from "./errors/RouteError";

export const routes: RouteObject[] = [
  {
    // The layout of the dashboard. Its ErrorBoundary catches each error below
    // it, also an unknown address: no blank screen.
    path: "/",
    element: <Outlet />,
    ErrorBoundary: RouteError,
    children: [
      {
        index: true,
        // Each page loads on demand: its code is a separate chunk.
        lazy: { Component: async () => (await import("./HomePage")).HomePage },
      },
    ],
  },
];
