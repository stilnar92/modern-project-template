import { problemKind, toProblem } from "@app/api-client";
import { notifications } from "@mantine/notifications";
import { PROBLEM_MESSAGES } from "./messages";

// notifyProblem shows the error of an action as a notification. Use it for
// mutations whose error has no place on the page. A refusal by a rule shows
// the detail from the server: the backend owns the rule and its text.
export function notifyProblem(error: unknown): void {
  const problem = toProblem(error);
  const kind = problemKind(problem);
  const { title, action } = PROBLEM_MESSAGES[kind];
  notifications.show({
    color: "red",
    title: title(),
    message: kind === "refused" && problem.detail ? problem.detail : action(),
  });
}
