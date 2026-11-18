// The key pages must work in every real viewport from viewports.ts:
// no horizontal scroll, and the main heading is visible. A screenshot of
// each page goes to the report, so a person or an agent can review it.
// Add each new key page to the list.
import { expect, test } from "@playwright/test";
import { dashboardUrl, siteUrl } from "../urls";

const pages = [
  { name: "dashboard home", url: `${dashboardUrl}/`, javaScriptEnabled: true },
  { name: "site home", url: `${siteUrl}/`, javaScriptEnabled: false },
];

for (const p of pages) {
  test.describe(p.name, () => {
    test.use({ javaScriptEnabled: p.javaScriptEnabled });

    test("fits the viewport", async ({ page }, testInfo) => {
      await page.goto(p.url);
      await expect(page.getByRole("heading", { level: 1 })).toBeInViewport();

      const overflow = await page.evaluate(
        () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
      );
      expect(overflow, "horizontal scroll in px").toBeLessThanOrEqual(0);

      await testInfo.attach("screenshot", {
        body: await page.screenshot({ fullPage: true }),
        contentType: "image/png",
      });
    });
  });
}
