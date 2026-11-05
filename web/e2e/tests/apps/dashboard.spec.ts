import { expect, test } from "@playwright/test";

test.use({ baseURL: "http://localhost:5173" });

test("the dashboard shows that the core answers", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("status")).toHaveText("Core: ok");
});
