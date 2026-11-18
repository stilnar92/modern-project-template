import "@mantine/core/styles.css";
import "@mantine/notifications/styles.css";
import { MantineProvider } from "@mantine/core";
import { Notifications } from "@mantine/notifications";
import * as Sentry from "@sentry/react";
import { MutationCache, QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot, type RootOptions } from "react-dom/client";
import { createBrowserRouter } from "react-router";
import { RouterProvider } from "react-router/dom";
import { notifyProblem } from "./errors/notifyProblem";
import { routes } from "./routes";
import { cssVariablesResolver, theme } from "./theme";

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

// A failed mutation shows a notification (frontend standard 27). A form that
// shows the errors next to its fields sets meta: { handlesErrors: true } and
// calls applyProblem. A failed query shows its error in its own area.
const queryClient = new QueryClient({
  mutationCache: new MutationCache({
    onError: (error, _vars, _ctx, mutation) => {
      if (!mutation.meta?.handlesErrors) notifyProblem(error);
    },
  }),
});

const router = createBrowserRouter(routes);

const container = document.getElementById("root");
if (!container) throw new Error('index.html has no element with id="root"');

createRoot(container, rootOptions).render(
  <StrictMode>
    <MantineProvider
      theme={theme}
      cssVariablesResolver={cssVariablesResolver}
      defaultColorScheme="auto"
    >
      <Notifications />
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </MantineProvider>
  </StrictMode>,
);
