import { expect, test } from "@playwright/test";
import { siteUrl } from "../../urls";

// No JavaScript: the status must come from the server-rendered HTML.
test.use({ baseURL: siteUrl, javaScriptEnabled: false });

test("the public page shows that the core answers", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("status")).toHaveText("Ядро: ok");
});
