# 0030 — Two applications must look like one product

From PRs #110, #117 · 12.11.2026 · Topics: design tokens, theming, accessibility

## Problem

The product has two frontends. The dashboard is a React app on a component library with its own theme object. The site is static pages with utility classes. They must have the same colours, sizes and radii, in light and in dark mode. A user must also select a look for the public page. If each app keeps its own values, they drift with the first change.

## Predict first

Both apps read the same colour values. One uses a component library that has its own dark theme. You change the system to dark mode. Do the two apps have the same background? Why, or why not?

## Theory

A **design token** is a design decision with a name: a colour, a size, a radius. Code uses the name and never the value. Tokens have tiers:

- **Primitive:** named for what it is, for example `gray-7`. It is the full set of permitted values.
- **Semantic:** named for its intent, for example `text-muted`. It points to a primitive.

Components use semantic tokens. So a change of look is a change of the mapping, not of the components. Dark mode is one more mapping: each semantic colour has a light and a dark value. A palette is another mapping of the same names.

Each platform reads tokens in its own form: CSS custom properties, a theme object, a file for a design tool. So tokens need **one source and generated outputs**. A test must compare the committed outputs with the source, or someone edits an output by hand.

For the colour scheme, CSS has two tools. `color-scheme` tells the browser which schemes the page supports, so the first paint already has the correct background. `light-dark()` selects one of two values by the used scheme. It is a newer function, so an older browser needs a `prefers-color-scheme` media query as the fallback.

A token system can also prove properties. WCAG needs a contrast of at least 4.5:1 for normal text. The contrast is a function of two colours, so a test can calculate it for each pair of text and background.

Sources:

- [Design Tokens Format Module](https://www.designtokens.org/tr/drafts/format/), Design Tokens Community Group: the terms, token types and aliases.
- [MDN: `light-dark()`](https://developer.mozilla.org/en-US/docs/Web/CSS/color_value/light-dark): the function and its browser support.
- [Improved dark mode default styling with the `color-scheme` property](https://web.dev/articles/color-scheme), web.dev: the property and the meta tag.
- [WCAG 2.2, Understanding 1.4.3 Contrast (Minimum)](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html): the 4.5:1 rule.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| DTCG files with Style Dictionary or Terrazzo | Tokens are JSON in a standard format. A tool builds the outputs | A dependency. The format helps a design tool that reads it, and the project has none |
| CSS as the source | The custom properties are the tokens | The theme of the dashboard needs literal colours, not CSS variables |
| Our own TypeScript package | One typed module holds the values. A script writes the CSS files, and the dashboard imports the values | We write and test the generator |

## Edge cases

| What if | What happens |
| --- | --- |
| The accent colour is hard to read | The first light accent had 4.20:1 on white. A contrast test now checks each text pair, in each palette and both schemes. The accent moved one shade darker, to 5.02:1 |
| The component library has its own dark colours | The dashboard showed the dark background of the library, not the token. A resolver now maps the semantic tokens to the page colours of the library |
| The browser does not know `light-dark()` | The CSS uses a `prefers-color-scheme` media query. `@supports` changes it to `light-dark()` where possible |
| A developer writes `bg-red-500` | The Tailwind theme has only the token colours. The class produces no CSS |
| Someone edits a generated CSS file | A test compares the committed files with the source and fails |
| Each user wants a different look | A palette is an entry in a closed list. The page selects it with `<html data-palette="…">`. Each palette is static CSS, so no inline style is necessary and the CSP stays strict |
| A paid design for one user | A designer adds a private palette through a pull request. The core records who can use it. The token file never holds a customer name or ID |

## Decision

One TypeScript file is the source of all design values. A script generates the CSS variables and the Tailwind theme for the site. The dashboard builds its Mantine theme from the same values. The reasons: no dependency, and no design tool in the project needs DTCG files. A later study compared DTCG, CSS as the source and other tools again, and the owner kept this choice.

The design tool shows a generated view of the tokens. It is not a source: nobody edits it.

This would be wrong if the project starts to use a design tool that reads and writes DTCG files. It is also a limit: a user cannot have free colours, only a palette from the list.

Record: [ADR 0015](../../dev/adr/0015-design-system-and-claude-design.md) and [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md), "Design tokens".

## In our code

1. `web/packages/tokens/src/tokens.ts`: the tiers, `TEXT_PAIRS` and `palettes`.
2. `web/packages/tokens/css/tokens.css`: the generated variables, the fallback and the `data-palette` blocks.
3. `web/packages/tokens/src/contrast.test.ts`: the contrast proof.
4. `web/apps/dashboard/src/theme.ts`: the Mantine theme and `cssVariablesResolver`.

## Check yourself

1. Why do components use `text-muted` and not `gray-7`, when both give the same colour today?
2. Both apps read the same tokens, and the dark backgrounds were still different. Why?
3. A new palette has an accent with a contrast of 4.2:1 on the surface. What stops it, and when?

<details><summary>Answers</summary>

1. `gray-7` is a value. `text-muted` is an intent. In dark mode, and in another palette, the muted text needs a different value. With the semantic name, only the mapping changes. With the primitive name, each component must change.
2. The component library sets its own page colours for the dark scheme. The tokens gave it only the colour scales, so its body background stayed its own. The apps matched only after a resolver mapped the semantic tokens to those page colours.
3. The contrast test calculates each text pair for each palette in both schemes. The pair of accent and surface is below 4.5:1, so the test fails in the local check and in CI. The palette cannot merge.

</details>
