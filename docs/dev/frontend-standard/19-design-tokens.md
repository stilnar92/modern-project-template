# 19 — Design Tokens

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

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
