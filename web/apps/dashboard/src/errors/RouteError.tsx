import { problemKind, toProblem, type Problem } from "@app/api-client";
import { Button, Container, Stack, Text, Title } from "@mantine/core";
import { isRouteErrorResponse, Link, useRouteError } from "react-router";
import { m } from "@app/i18n";
import { PROBLEM_MESSAGES } from "./messages";

// The router throws its own responses, for example a 404 for an unknown
// address. They are not service failures (UX standard U14).
function routeProblem(error: unknown): Problem {
  if (isRouteErrorResponse(error)) {
    return { type: "about:blank", title: error.statusText, status: error.status };
  }
  return toProblem(error);
}

// RouteError is the ErrorBoundary of the routes (frontend standard 27): an
// error in a loader or in rendering shows this page instead of a blank
// screen. Sentry gets the error from the root options in main.tsx.
export function RouteError() {
  const kind = problemKind(routeProblem(useRouteError()));
  const { title, action } = PROBLEM_MESSAGES[kind];
  return (
    <Container py="xl">
      <Stack align="flex-start" role="alert">
        <Title order={1}>{title()}</Title>
        <Text>{kind === "not_found" ? m.error_page_not_found() : action()}</Text>
        {kind === "not_found" ? (
          <Button component={Link} to="/">
            {m.error_go_home()}
          </Button>
        ) : (
          <Button onClick={() => window.location.reload()}>{m.error_reload()}</Button>
        )}
      </Stack>
    </Container>
  );
}
