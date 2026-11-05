import { expect, test } from "@playwright/test";

// No JavaScript: the status must come from the server-rendered HTML.
test.use({ baseURL: "http://localhost:4330", javaScriptEnabled: false });

test("the public page shows that the core answers", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("status")).toHaveText("Core: ok");
});
