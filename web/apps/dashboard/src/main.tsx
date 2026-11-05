import "@mantine/core/styles.css";
import { MantineProvider } from "@mantine/core";
import * as Sentry from "@sentry/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot, type RootOptions } from "react-dom/client";
import { createBrowserRouter } from "react-router";
import { RouterProvider } from "react-router/dom";
import { HomePage } from "./HomePage";

// Error tracking is on only when VITE_SENTRY_DSN is set, as in the core
// (frontend standard D13). Browser tracing also reports Web Vitals.
const dsn = import.meta.env.VITE_SENTRY_DSN;
let rootOptions: RootOptions = {};
if (dsn) {
  Sentry.init({ dsn, integrations: [Sentry.browserTracingIntegration()], tracesSampleRate: 0.1 });
  rootOptions = {
    onUncaughtError: Sentry.reactErrorHandler(),
    onCaughtError: Sentry.reactErrorHandler(),
    onRecoverableError: Sentry.reactErrorHandler(),
  };
}

const queryClient = new QueryClient();
const router = createBrowserRouter([{ path: "/", element: <HomePage /> }]);

const container = document.getElementById("root");
if (!container) throw new Error('index.html has no element with id="root"');

createRoot(container, rootOptions).render(
  <StrictMode>
    <MantineProvider>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </MantineProvider>
  </StrictMode>,
);
