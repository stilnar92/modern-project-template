import { fieldErrors, problemKind, toProblem } from "@app/api-client";
import { notifyProblem } from "./notifyProblem";

// The part of a @mantine/form form that applyProblem needs.
type FormErrors = { setErrors: (errors: Record<string, string>) => void };

// applyProblem shows the error of a form submit (UX standard U6, U13, U14).
// Field errors go next to their fields, and the form keeps its values. Any
// other error goes to a notification, because its cause is not in a field.
// It returns true when the error went to the fields.
export function applyProblem(form: FormErrors, error: unknown): boolean {
  const problem = toProblem(error);
  if (problemKind(problem) === "validation" && problem.errors?.length) {
    form.setErrors(fieldErrors(problem));
    return true;
  }
  notifyProblem(error);
  return false;
}
