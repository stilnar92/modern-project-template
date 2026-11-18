# 19 — Design Tokens

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## In this project

- `web/packages/tokens/src/tokens.ts` is the one source ([ADR 0012](../adr/0012-infrastructure-libraries.md), [ADR 0015](../adr/0015-design-system-and-claude-design.md)): tier 1 primitives (`gray`, a 10-shade `brand` scale per palette, `font`, `radius`) and tier 2 semantic colors (`surface`, `text`, `text-muted`, `border`, `accent`, `on-accent`) with a light and a dark value. `SEMANTIC_USAGE` says where to use each one. The colors are placeholders until the design sets the brand.
- **Palettes:** `palettes` is a closed list. Each palette sets the brand scale and all semantic colors. `default` is for the dashboard and the site. A site page selects another one with `<html data-palette="<id>">`. A `public` palette is a self-serve choice of the master, a `private` one is a paid individual design. Never put a customer name or ID in `tokens.ts`.
- `pnpm gen` (part of `mise run gen`) writes `css/tokens.css` (the `--ms-*` variables, `color-scheme: light dark`, a `prefers-color-scheme` fallback, `light-dark()` in `@supports` and one static `:root[data-palette]` block per palette), `css/tailwind.css` (the Tailwind `@theme inline`, with the default Tailwind colors removed) and the Claude Design view in `design-system/project/`. A test fails when a committed file differs.
- The site imports both CSS files. The dashboard builds its Mantine theme from the same values (`apps/dashboard/src/theme.ts`), and `cssVariablesResolver` maps the semantic colors to the Mantine page colors in both schemes.
- The contrast test checks each pair in `TEXT_PAIRS` (at least 4.5:1) in each palette and both schemes. It found the first accent value too low (4.2:1).
- Add a value as a token first. Do not write a color, size or radius in a component. Tailwind knows only the token colors: `bg-red-500` produces no CSS.
- After a token change, republish the Claude Design view ([ADR 0015](../adr/0015-design-system-and-claude-design.md)).

## Collected points

### From Locked Decisions

- **D7** — all design values flow through CSS custom properties: design tokens are implemented as CSS variables, and Tailwind consumes them. No design value (color, spacing, radius, shadow, font, motion) is hardcoded outside the token layer.

### Token naming conventions

Source: [Design Token Naming Conventions](https://www.alwaystwisted.com/articles/design-token-naming-conventions) (Stuart Robson).

**Three tiers** (confirms the tiering in this area's charter):

- **Primitive** — raw values, named for *what* they are (`blue.500`, `spacing.8`);
- **Semantic** — named for *intent/why* (`color.brand.primary`, `color.text.error`);
- **Component** — map semantic intent to UI elements (`button.primary.background`).

**Naming rules:**

- One convention, applied everywhere — never mix models within a tier; "the best naming system is the one the whole team actually follows."
- Clarity over brevity — `button-background-primary-hover`, never `btn-bg-pri-hvr`.
- Semantic tiers name intent, not appearance — `color-text-error`, never `color-text-red` (intent survives value changes).
- Order broad → specific, so alphabetical sorting naturally groups related tokens.
- Numeric scales that leave room for growth — `spacing.100/200`, never `spacing-small` / `spacing-new-2`.
- States/variants/modes live in one fixed position (stable suffix or stable prefix — either works, switching between them doesn't).
- Names must be self-explanatory to a newcomer with no extra context.
- Naming is governed, not one-shot: a naming guide, review checklist, periodic cleanup, shared design–engineering ownership.

**Anti-patterns:** mixed conventions across tiers; positional naming; visual-property names in semantic tiers (`color.background.grey`); ambiguous abbreviations; inconsistent state placement.

**Convention candidates** for the elaboration phase (article surveys five: CPM, tier-based, OPM, context-first, comprehensive) — pick per system scale; context-first suits multi-theme systems, which matters given light/dark theming is in scope.
