import { expect, test } from "@playwright/test";
import { dashboardUrl } from "../../urls";

test.use({ baseURL: dashboardUrl });

test("the dashboard shows that the core answers", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("status")).toHaveText("Ядро: ok");
});
