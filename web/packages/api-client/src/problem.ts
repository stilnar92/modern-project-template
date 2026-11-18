// Errors of the core API in one form (backend standard 02, frontend
// standard 27, ADR 0012). Each failed call becomes a Problem (RFC 9457). The
// UI branches on problemKind and on the Problem "type", never on the text.
import { zProblem } from "./gen/zod.gen";
import type { components } from "./schema";

export type Problem = components["schemas"]["Problem"];

// The kinds of error that the UX standard (U14) shows differently:
// - validation: a field is wrong; the form shows it next to the field;
// - refused: a business rule refuses the action; the cause is not in a field;
// - unauthenticated: the session ended; log in again;
// - forbidden: no permission;
// - not_found: the thing does not exist (any more);
// - service: the server or the network failed; try again later.
export type ProblemKind =
  | "validation"
  | "refused"
  | "unauthenticated"
  | "forbidden"
  | "not_found"
  | "service";

// ApiError carries the problem of a failed call.
export class ApiError extends Error {
  readonly problem: Problem;

  // body is the parsed error body. When it is not a Problem (for example
  // the HTML page of a proxy), the error keeps only the status.
  constructor(body: unknown, status?: number) {
    const parsed = zProblem.safeParse(body);
    const problem: Problem = parsed.success ? parsed.data : serviceProblem(status ?? 0);
    super(problem.detail ?? problem.title);
    this.name = "ApiError";
    this.problem = problem;
  }
}

function serviceProblem(status: number): Problem {
  return { type: "about:blank", title: status ? "Service Error" : "Network Error", status };
}

// toProblem turns any thrown value into a Problem. A network failure (no
// response) becomes a service problem with status 0.
export function toProblem(error: unknown): Problem {
  if (error instanceof ApiError) return error.problem;
  return serviceProblem(0);
}

export function problemKind(p: Problem): ProblemKind {
  if (p.status === 400 || (p.status === 422 && p.errors?.length)) return "validation";
  if (p.status === 401) return "unauthenticated";
  if (p.status === 403) return "forbidden";
  if (p.status === 404) return "not_found";
  if (p.status === 409 || p.status === 422) return "refused";
  return "service";
}

// fieldErrors maps the "errors" extension to form paths: "/items/0/name"
// becomes "items.0.name", the path form of @mantine/form.
export function fieldErrors(p: Problem): Record<string, string> {
  const out: Record<string, string> = {};
  for (const e of p.errors ?? []) {
    const path = e.pointer
      .split("/")
      .slice(1)
      .map((part) => part.replaceAll("~1", "/").replaceAll("~0", "~"))
      .join(".");
    out[path] = e.detail;
  }
  return out;
}

// unwrap returns the data of an openapi-fetch result, or throws an ApiError.
// Use it in query and mutation functions, so React Query sees the error.
export function unwrap<T>(result: { data?: T; error?: unknown; response: Response }): T {
  if (result.error !== undefined || result.data === undefined) {
    throw new ApiError(result.error, result.response.status);
  }
  return result.data;
}
