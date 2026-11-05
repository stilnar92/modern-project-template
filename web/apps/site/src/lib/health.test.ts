import { createApiClient } from "@app/api-client";
import { afterEach, expect, test, vi } from "vitest";
import { readHealth } from "./health";

// The client keeps the fetch it gets at creation, so stub fetch first.
afterEach(() => vi.unstubAllGlobals());

test("ok when the core answers", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => Response.json({ status: "ok" })),
  );
  expect(await readHealth(createApiClient("http://core.test"))).toBe("ok");
});

test("unavailable when the core does not answer", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => {
      throw new TypeError("fetch failed");
    }),
  );
  expect(await readHealth(createApiClient("http://core.test"))).toBe("unavailable");
});
