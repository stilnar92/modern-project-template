import { describe, expect, test } from "vitest";
import { ApiError, fieldErrors, problemKind, toProblem, unwrap } from "./problem";

const problem = (status: number, extra: object = {}) => ({
  type: "about:blank" as const,
  title: "x",
  status,
  ...extra,
});

describe("toProblem", () => {
  test("keeps the problem of an ApiError", () => {
    const p = problem(422, { errors: [{ pointer: "/price", detail: "must be more than 0" }] });
    expect(toProblem(new ApiError(p))).toEqual(p);
  });

  test("a network failure is a service problem without a status", () => {
    expect(toProblem(new TypeError("Failed to fetch"))).toMatchObject({ status: 0 });
  });

  test("a body that is not a problem (an HTML 502 of a proxy) keeps the status", () => {
    const err = new ApiError(undefined, 502);
    expect(toProblem(err)).toMatchObject({ status: 502, type: "about:blank" });
  });
});

describe("problemKind (UX standard U14)", () => {
  test.each([
    [400, "validation"],
    [401, "unauthenticated"],
    [403, "forbidden"],
    [404, "not_found"],
    [409, "refused"],
    [500, "service"],
    [503, "service"],
    [0, "service"],
  ])("status %i is %s", (status, kind) => {
    expect(problemKind(problem(status))).toBe(kind);
  });

  test("a 422 without field errors is a refusal by a rule, not a field error", () => {
    expect(problemKind(problem(422))).toBe("refused");
    expect(problemKind(problem(422, { errors: [{ pointer: "/a", detail: "d" }] }))).toBe(
      "validation",
    );
  });
});

test("fieldErrors maps JSON pointers to form paths", () => {
  const p = problem(422, {
    errors: [
      { pointer: "/price", detail: "must be more than 0" },
      { pointer: "/items/0/name", detail: "is empty" },
      { pointer: "/a~1b", detail: "escaped" },
    ],
  });
  expect(fieldErrors(p)).toEqual({
    price: "must be more than 0",
    "items.0.name": "is empty",
    "a/b": "escaped",
  });
});

describe("unwrap", () => {
  test("returns the data", () => {
    expect(
      unwrap({ data: { status: "ok" }, response: new Response(null, { status: 200 }) }),
    ).toEqual({ status: "ok" });
  });

  test("throws an ApiError with the problem", () => {
    const p = problem(409);
    expect(() => unwrap({ error: p, response: new Response(null, { status: 409 }) })).toThrow(
      ApiError,
    );
  });
});
