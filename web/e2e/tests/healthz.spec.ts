import { createApiClient } from "@app/api-client";
import { expect, test } from "@playwright/test";

test("the core answers on /healthz", async ({ request }) => {
  const response = await request.get("/healthz");
  expect(response.status()).toBe(200);
  expect(await response.json()).toEqual({ status: "ok" });
});

test("the generated client reads /healthz", async ({ baseURL }) => {
  const api = createApiClient(baseURL!);
  const { data, error } = await api.GET("/healthz");
  expect(error).toBeUndefined();
  expect(data?.status).toBe("ok");
});
