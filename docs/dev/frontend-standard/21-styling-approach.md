# 21 — Styling Approach

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D5** — the styling stack is **Tailwind CSS + tailwind-variants**.
- **D6** — no ad-hoc `className` styling in components: all component styling is expressed through tailwind-variants definitions (`tv()`), including one-off cases.
- **D7** — Tailwind consumes design tokens as CSS custom properties; styles reference tokens, never raw values.
- **D2** — modern CSS by default (nesting, container queries, `:has()`, logical properties, etc.) within the browser support policy ([10](10-browser-support.md)).
- **D8** — responsive utilities are required by default; a component without responsive behavior is incomplete.
- **D14** — responsiveness is implemented fluid-first: intrinsic layout (Grid/Flexbox, content-driven sizing) and container queries before viewport breakpoints; breakpoints are the last resort, not the starting point.

### CSS before JS

Source: [You Don't Need JS](https://lyra.horse/blog/2025/08/you-dont-need-js/) (lyra.horse). Prefer native CSS for behavior traditionally done in JavaScript — better performance (compositor-thread animations vs event loop), fewer failure modes. All bounded by the browser support policy ([10](10-browser-support.md), D2).

- **`@starting-style`** — entrance/fade-in animations on DOM insertion, no keyframes or JS listeners.
- **`:has()`** — parent/sibling styling from child state (e.g. `:has(input:checked)`) for tabs, accordions, conditional UI; also presence-based component states (`figure:has(figcaption)`) instead of a variant/class. Caveat: scope carefully (e.g. `body:has()`), it can get expensive.
- **`color-scheme: light dark` + `light-dark()`** — OS-driven theming; `color-scheme` makes native widgets (scrollbars, buttons) follow too.
- **CSS nesting** — no preprocessor needed.
- **Relative colors + `color-mix()`** — derive hover/variant colors in CSS (`hsl(from var(--color) …)`), no JS color math.
- **Modern media query range syntax** — `(width <= 768px)` over `(max-width: 768px)`.
- **`svh` / `lvh` / `dvh` instead of `vh`** (also: [New Viewport Units](https://ishadeed.com/article/new-viewport-units/), Ahmad Shadeed). Classic `100vh` ignores mobile browser UI (address bar/toolbars), causing overflow and hidden content on iOS Safari and Android.
  - `svh` — viewport with browser UI expanded (smallest): must-fit UI at initial load — hero sections (`svh` + `calc()` with header height), modals with sticky headers/footers;
  - `lvh` — viewport with browser UI collapsed (largest): backgrounds that tolerate cutoff;
  - `dvh` — tracks browser UI in real time: full-screen modals/overlays, but sparingly — continuous recalculation costs performance and causes jank; never for `font-size`.
  - Caveats: pair with `<meta name="viewport" content="interactive-widget=resizes-content">` for on-screen keyboards; logical-axis equivalents (`svi`/`svb`, etc.) exist and follow the logical-properties preference (D2).
- **`scrollbar-gutter`** — reserve scrollbar space, no layout shift; `stable both-edges` for visual balance. Caveat: no effect with overlay scrollbars (macOS default) — keep padding independent.
- **`:focus-visible`** — keyboard-only focus rings, no JS focus tracking.
- **Animations in CSS by default** — compositor thread beats JS event-loop animation, especially on low-end devices; Web Animations API (`element.animate()`) only as progressive enhancement when JS logic must trigger the animation.

### Modern CSS one-liners

Source: [12 Modern CSS One-Line Upgrades](https://moderncss.dev/12-modern-css-one-line-upgrades/) (Stephanie Eckles). Small default upgrades — candidates for the base stylesheet/reset. (`aspect-ratio` + `object-fit` are covered under images/CLS in [30](30-performance.md); `color-scheme` and `scrollbar-gutter` are already above.)

- **`margin-inline` / logical shorthands** — instead of left/right pairs; follows the logical-properties preference (D2), writing-mode-safe.
- **`text-underline-offset`** — legible underlines clearing descenders; part of the base link reset, with `text-decoration-color`/`-thickness` as companions.
- **`outline-offset`** — custom focus-outline positioning without box-shadow hacks; outlines don't affect layout size.
- **`scroll-margin-top`** — on `[id]` targets so sticky headers don't obscure anchor-linked content.
- **`accent-color`** — themes native checkboxes/radios/progress/range without custom styling; pairs with `color-scheme` as the theming baseline.
- **`width: fit-content`** (logical: `inline-size`) — shrink-wrap without `display: inline-block` side effects.
- **`overscroll-behavior`** — stop scroll chaining from contained scroll regions (sidebars, modals) to the page.
- **`text-wrap: balance`** for headlines (≤6 lines), **`text-wrap: pretty`** against orphans elsewhere — progressive enhancement; doesn't change computed width, may add surrounding whitespace.

### CSS units policy

Source: [CSS Units — When to Use Each One](https://www.freecodecamp.org/news/css-units-when-to-use-each-one/) (Esther Christopher). Since all design values flow through tokens (D7), this policy applies primarily where tokens are defined.

- **`rem`** — the default for font sizes and the spacing scale: scales with the user's root font-size preference (accessibility); root-consistent, no parent surprises.
- **`em`** — only when scaling relative to the *local* font size is the point (e.g. padding inside a component that tracks its own text size).
- **`%`** — container-relative width/height/positioning.
- **`ch`** — readable line lengths (`max-width` on text blocks, e.g. `65ch`).
- **`px`** — borders and intentionally fixed dimensions only; never for font sizes — px ignores user preferences.
- **Viewport units** — per the `svh`/`lvh`/`dvh` entry above; and for component-relative sizing, container query units (below) are preferred over viewport units (D14).

Source: [CSS Container Query Units](https://ishadeed.com/article/container-query-units/) (Ahmad Shadeed). Complements D14: components scale to their **container**, not the viewport, so they behave predictably wherever they're placed. (Article uses pre-spec names `qw`/`qh`; final units are `cqw`, `cqh`, `cqi`, `cqb`, `cqmin`, `cqmax`.)

- Units: `cqw`/`cqh` = 1% of container width/height; `cqi`/`cqb` = inline/block-axis equivalents (writing-mode-safe, prefer these); `cqmin`/`cqmax` = smaller/larger of the two.
- **Fluid typography per component** — `font-size: clamp(1rem, 3cqi, 2rem)`: type scales with the container, no breakpoint ladder.
- **Context-aware sizing** — the same component scales differently in `<aside>` vs `<main>` because each is its own inline-size container; no variant props needed for size context.
- **Container-relative dimensions and spacing** — element sizes (`--size: calc(60px + 10cqi)`) and gaps derived from the container instead of viewport units.
- Caveat: with no container defined, the units fall back to small-viewport units — components using cq units must declare their container (`container-type: inline-size`) explicitly.
- Caveat: a component can't query itself — container queries need a wrapper element as the container, or they'd loop; name containers (`container-name`) when nesting.

### Layout patterns

Source: [Solved by Modern CSS: Feature Image](https://ishadeed.com/article/modern-css-feature-image/) (Ahmad Shadeed) — worked example combining the techniques above.

- **Grid overlap instead of absolute positioning** — `display: grid` + `grid-area: 1 / -1` layers elements in the same cell; keeps layering in flow, no positioning context gymnastics.
- **Component-internal derived sizing** — define one custom property (e.g. `--caption-size: clamp(…)`) and derive related dimensions with `calc()` from it, so the component scales as one system (fits D7's variables-first approach at component scope).
