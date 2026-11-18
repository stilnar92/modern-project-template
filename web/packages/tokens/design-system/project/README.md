# modern-project-template design system

The design system of modern-project-template, a platform for beauty masters. It has two frontends:

- **Dashboard:** the dashboard of the master (React and Mantine). It has one strict look and uses the `default` palette only.
- **Site:** the public pages of masters (Astro, no browser JavaScript). Each page selects a palette. Later it will also select a font pair and a layout.

**Source of truth:** the repository, `web/packages/tokens/src/tokens.ts`. This view is generated from it and is read-only. To change a value, change the file and run `mise run gen`.

**Status:** the colors are placeholders (Mantine default palettes). The brand is not chosen yet. The token names are stable. Only the values will change.

## Tokens

- **Tiers:** primitive scales (`brand-50` … `brand-900`, `gray-50` … `gray-900`), then semantic colors. Components use only semantic colors: `surface`, `text`, `text-muted`, `border`, `accent`, `on-accent`.
- **Schemes:** each semantic color has a light and a dark value. The page follows the system color scheme.
- **Palettes:** a palette is one look of a page: its brand scale and all semantic colors. The themes of this view are `<palette> light` and `<palette> dark`. A public palette is a self-serve choice of the master. A private palette is a paid individual design.
- **Contrast:** a test checks every palette in both schemes. Text pairs have at least 4.5:1: `text`, `text-muted` and `accent` on `surface`, and `on-accent` on `accent`.
- **Text:** the system font stack. Sizes use the Tailwind scale (`text-sm`, `text-base`, `text-xl`, `text-3xl`).
- **Radius:** `radius-sm` for small controls, `radius-md` (default) for buttons, inputs and cards, `radius-lg` for large surfaces.

## Rules for mockups

- Use only the tokens of this system. Do not add a color, size or radius in a mockup. If a mockup needs a new value, write it in the notes of the mockup. The new value becomes a token through a pull request.
- Design the phone screen first. Show every state: loading, empty, error, success.
- The text of the product is in Russian.
- The dashboard uses Mantine components. Make dashboard mockups with Mantine patterns, so the step to code is short.
- Site mockups must work without JavaScript.

## Without Claude Design

The same tokens are in `web/packages/tokens/css/tokens.css`. A mockup can be plain HTML that links this file and sets `<html data-palette="<id>">`.
