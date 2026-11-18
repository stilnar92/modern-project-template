import "@tanstack/react-query";

// The meta fields that the dashboard reads (main.tsx).
declare module "@tanstack/react-query" {
  interface Register {
    mutationMeta: { handlesErrors?: boolean };
  }
}
