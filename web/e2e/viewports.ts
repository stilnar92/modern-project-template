// Real viewports, not only device presets (frontend standard 10).
// People seldom see a page at a "clean" size: DevTools take part of the
// window, vertical tabs take 150-200px, zoom makes the CSS viewport smaller,
// and a phone keyboard leaves a thin strip in landscape. The idea comes from
// Site Seeing by Ahmad Shadeed: https://lab.ishadeed.com/tools/site-seeing/
import { devices, type Project } from "@playwright/test";

const desktop = devices["Desktop Chrome"];

export const viewportProjects: Project[] = [
  {
    name: "vp-devtools-right",
    use: { ...desktop, viewport: { width: 1040, height: 900 } },
  },
  {
    name: "vp-devtools-bottom",
    use: { ...desktop, viewport: { width: 1440, height: 500 } },
  },
  {
    name: "vp-vertical-tabs",
    use: { ...desktop, viewport: { width: 1240, height: 900 } },
  },
  // A 1280x900 window at 150% zoom: the page gets 853x600 CSS pixels.
  {
    name: "vp-zoom-150",
    use: { ...desktop, viewport: { width: 853, height: 600 }, deviceScaleFactor: 1.5 },
  },
  // iPhone in landscape with the keyboard open.
  {
    name: "vp-phone-landscape-keyboard",
    use: { ...devices["iPhone 13 landscape"], viewport: { width: 750, height: 160 } },
  },
].map((p) => ({ ...p, testMatch: /viewports\.spec\.ts/ }));
