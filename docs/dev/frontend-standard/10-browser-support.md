# 10 — Browser Support & Compatibility

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Decided

### Browser target

- The target is **Baseline Widely available**: `"browserslist": ["baseline widely available"]` in `web/package.json`. The apps find it in the parent `package.json`.
- A feature that is Baseline Widely available needs no fallback.
- A newer feature needs the fallback from its Modern Web Guidance guide. No polyfill by default (D1).
- Vite 8 and Astro use the same target by default (`build.target: "baseline-widely-available"`). Do not set a lower target.

### Modern Web Guidance

- [Modern Web Guidance](https://developer.chrome.com/docs/modern-web-guidance) from Chrome gives agents current guides for the web platform. The package is pinned in `web/package.json` (`modern-web-guidance`).
- The project skill `.claude/skills/modern-web-guidance` (copy in `.agents/skills`) tells the agent to search and retrieve a guide before it writes HTML, CSS or client code.
- `pnpm legacy-scan` runs in `mise run lint:web` and in CI. It fails on a legacy pattern that the scan knows. The scan covers about 20% of the guides. The tool exits with 0 on findings, so `web/scripts/legacy-scan.mjs` reads its JSON output.
- An exception covers one file and one rule and has a reason: `web/legacy-scan.allow.json`.
- The commands set `DISABLE_TELEMETRY=1`. By default the tool sends usage data to Google.

### Real viewports (D9, D14)

People seldom see a page at a "clean" device size. The e2e projects in `web/e2e/viewports.ts` copy real cases from [Site Seeing](https://lab.ishadeed.com/tools/site-seeing/) by Ahmad Shadeed:

| Project | CSS viewport | Case |
| --- | --- | --- |
| `vp-devtools-right` | 1040 × 900 | DevTools docked to the right |
| `vp-devtools-bottom` | 1440 × 500 | DevTools docked to the bottom |
| `vp-vertical-tabs` | 1240 × 900 | Vertical tabs take 150–200px |
| `vp-zoom-150` | 853 × 600 | A 1280 × 900 window at 150% zoom |
| `vp-phone-landscape-keyboard` | 750 × 160 (WebKit) | iPhone in landscape, keyboard open |

- `web/e2e/tests/viewports.spec.ts` opens each key page in each project. There must be no horizontal scroll, and the `h1` must be in the viewport. A full-page screenshot goes to the report for a review.
- Add each new key page to the list in the spec.
- For a manual check, open Site Seeing and enter the localhost URL of the app.

## Collected points

### From Locked Decisions

- **D1** — target modern evergreen browsers with an explicit browserslist; polyfills only where that policy requires them, never by default.
- **D2** — modern CSS by default (nesting, container queries, `:has()`, logical properties, etc.) within the bounds of the browser support policy. This document defines exactly what "modern CSS" may rely on.
- **D9** — cross-browser behavior is always considered: engine differences (Safari/WebKit quirks, etc.) are accounted for during development and verified during testing.
