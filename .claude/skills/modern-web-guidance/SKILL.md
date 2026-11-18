---
name: modern-web-guidance
description: Look up current web platform guidance (Chrome Modern Web Guidance) before you write or change HTML, CSS or client-side code in web/. Use for dialogs, popovers, forms and validation, layout, container queries, animations, accessibility, Core Web Vitals (LCP, INP, CLS), images, security and dates. Do not use for backend, CI, build config or tests.
---

# Modern Web Guidance

The guides are in the pinned `modern-web-guidance` package in `web/`. They are
vanilla HTML, CSS and JavaScript. Use the modern feature in React or Astro, and
keep the feature when you adapt the code.

Run the commands from `web/`. Set `DISABLE_TELEMETRY=1`: the tool sends usage
data to Google by default.

## 1. Search

Use 2 to 5 words for one capability. Name the UI primitive. Do not put
framework, library or product names in the query. Split a task with two
features into two searches.

```sh
DISABLE_TELEMETRY=1 pnpm exec modern-web-guidance search "light dismiss dialog"
```

If the results do not match, rephrase one time with a concrete noun
(`dialog`, `popover`, `table`, `form`). If they still do not match, continue
without a guide.

## 2. Retrieve

Retrieve the matching guides and the broad guide of the discipline (`html`,
`css`, `forms`, `accessibility`, `performance`, `security`). Read all of the
output. If the tool writes it to a file, read that file.

```sh
DISABLE_TELEMETRY=1 pnpm exec modern-web-guidance retrieve "light-dismiss-a-dialog,accessibility"
```

## 3. Apply and check

- Browser target: Baseline Widely available (`browserslist` in
  `web/package.json`, frontend standard 10). A feature that is not Widely
  available needs the fallback from the guide.
- Check the code against the guide before you finish.
- In the PR, write one line per guide that changed the code: the guide ID and
  the pitfall that it prevents.

## Legacy scan

`mise run lint:web` runs `pnpm legacy-scan`. It fails on a legacy pattern
that the scan knows (about 20% of the guides). Fix the code with the guide in
the message. If the code is a fallback that a guide requires, add one entry to
`web/legacy-scan.allow.json` with the file, the rule and the reason.
