# 20 — Component System

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.
**Locked decisions that apply here:** D6 (variants are the only styling surface — tailwind-variants APIs, no raw `className` styling).

---

## Collected points

### Bulletproof components

Source: [Build Bulletproof React Components](https://shud.in/thoughts/build-bulletproof-react-components) (Shu Ding).

Components must survive hostile real-world conditions — SSR, hydration, multiple instances, concurrent rendering, portals — not just the happy path:

1. **Server-proof** — browser APIs (`localStorage`, `window`, …) only inside `useEffect`, never during render, so SSR doesn't crash.
2. **Hydration-proof** — when client state must be known before first paint (e.g. theme), use a synchronous inline script that runs before hydration; no flash of wrong content, no hydration mismatch.
3. **Instance-proof** — IDs via `useId`, never hardcoded; multiple instances of a component must coexist without colliding over DOM ids.
4. **Concurrent-proof** — deduplicate per-request work in Server Components with `React.cache()`.
5. **Composition-proof** — pass data to children via Context, not `React.cloneElement`; cloning breaks with Server Components, lazy components, and opaque children.
6. **Portal-proof** — attach listeners via `ref.current?.ownerDocument.defaultView`, not global `window`, so components work in iframes, pop-out windows, and portals.
7. **Transition-proof** — state updates that trigger view transitions go through `startTransition()`.
8. **Activity-proof** — component-scoped global side effects (e.g. injected `<style>`) must deactivate when the component is hidden/deactivated, not persist.
9. **Leak-proof** — sensitive server values are tainted (`taintUniqueValue` / `taintObjectReference`) so React throws if they cross into Client Components.
10. **Future-proof** — never use `useMemo` for correctness (it's a performance hint that React may discard); use `useState`/refs when persistence is required.

### Native elements first

Source: [You Don't Need JS](https://lyra.horse/blog/2025/08/you-dont-need-js/) (lyra.horse).

- Before building a JS-stateful component, check whether a native element covers it: `<details>/<summary>` for accordions (keyboard nav and Ctrl+F search for free), hidden `<input>` + styled `<label>` with `:checked` for radio pickers/tabs.
- Native-first is not dogma: complex interactions (date pickers, dynamic forms, persisted preferences) legitimately need JS — the rule is "simplest tool that covers the requirement," not "no JS."

### React implementation tricks

Source: [React Tricks](https://molefrog.com/notes/react-tricks) (Alexey Taktarov, wouter maintainer).

- **asChild pattern** — customize a child element (inject props, e.g. `href`, or override `ref`) via `cloneElement` instead of wrapping in extra DOM nodes. ⚠️ Tension with bulletproof pt. 5 (Context over `cloneElement` — cloning breaks with Server Components / lazy children): resolve during elaboration — likely "Context for data flow, asChild only for prop injection on direct known children."
- **`useState` initializer for one-time setup** — `const [value] = useState(() => expensiveInit())`: runs once per component lifetime, reference never changes; also the pattern for per-instance lazy singletons (`const [client] = useState(() => new Client())`). Caveat: Strict Mode runs initializers twice in dev. Minifies better than `useRef`.
- **`useEvent`-style stable callbacks** — stable reference that always calls the latest closure, no deps array; unlike `useCallback` it never invalidates. Still an RFC — use a userland implementation deliberately, not ad hoc.
- **Stable references matter** — new object/function identities per render break memoization; stabilize via module scope, `useState`, or `useEvent`. ⚠️ Note: the React Compiler ([30](30-performance.md)) does much of this automatically — during elaboration decide which of these manual tricks remain justified compiler-side.

### Atomic design

Source: Atomic Design (Brad Frost) — adopted as the component taxonomy, per user decision.

- **Five levels**: atoms (indivisible UI — button, input, icon) → molecules (small compositions — search field, form row) → organisms (self-contained sections — header, card list) → templates (page-level layout with placeholder content) → pages (templates with real content).
- The levels are a **shared vocabulary and folder taxonomy for the reusable UI kit**; they say where a component lives and what it may depend on (atoms depend on nothing above tokens; molecules on atoms; …).
- ✅ Resolved (user decision, see [12](12-architecture.md)): the UI kit and business components are separate layers — atoms/molecules/organisms organize the kit; business components are feature-owned, compose the kit, and are not forced into atomic naming. Still to define at elaboration: the classification rule (molecule vs organism) to avoid ambiguity debates.
- Folder mapping lives in [14 — Project Structure](14-project-structure.md).

### Headless component patterns

The behavior/rendering split ([D3](00-overview.md) at component scale): logic and ARIA behavior in a headless layer, visuals in variants (D6).

- **Hook-based headless** — behavior as a hook returning state + prop getters (`useSelect()` → `getToggleProps()`, `getMenuProps()`); the component consuming it is purely presentational.
- **Compound components** — parent + children sharing implicit state via Context (`<Select><Select.Trigger/><Select.List/></Select>`); composition-proof (Context, not `cloneElement` — see bulletproof pt. 5).
- **Controlled / uncontrolled duality** — components work uncontrolled by default but accept `value`/`onChange` to become controlled; both modes, one API.
- **State reducer pattern** — consumer can intercept/override internal state transitions (inversion of control) for edge-case behavior without forking the component.
- **Slot / `render` composition** — inject the rendered element (e.g. Ariakit's `render` prop) so behavior attaches to arbitrary markup.
- Sourcing ladder stays: native element → headless library (Ariakit / React Aria) → hand-built per APG ([22](22-accessibility.md)); these patterns apply to the last two rungs.

### Extending component styles (under D6)

With `className` banned from component APIs, every styling need gets a sanctioned channel:

1. **New look → new variant at the source** — even one-offs become named variants in the `tv()` definition (per D6); the name makes the style discoverable and reviewable.
2. **Contextual spacing → the parent owns layout** — components carry **no external margins**; containers position children via `gap`, grid, or wrappers. Kills the most common motivation for `className` passthrough.
3. **Specialization → `tv({ extend })`** — a derived component inherits base + variants and adds its own; extension yields a new catalogued component, never an anonymous call-site patch.
4. **Context-dependent appearance → token indirection (D7)** — components paint with semantic tokens; a surface remaps the CSS variables (`.inverted-surface { --color-text: …; }`) and every component on it adapts untouched. The mechanism for theming, dark surfaces, brand sections.
5. **Combination fixes → `compoundVariants`** — variant-combination knowledge stays inside the definition.

⚠️ Deliberately absent: the shadcn-style `className` + `twMerge` escape hatch — the strict D6 reading rejects it; developers arriving from that ecosystem should expect the difference. Revisit at elaboration only if practice shows channels 1–5 leave a real gap.

Source: [A11Y is UI Architecture](https://dev.to/artalar/a11y-is-ui-architecture-2ibn) (Artyom). Input for this area's headless-vs-styled / build-vs-buy policy:

- Headless UI libraries (Ariakit, React Aria/Spectrum) decouple WAI-ARIA behavior (roles, focus, keyboard) from visual rendering — the behavioral layer comes proven from the library, visuals come from our variants (D6). Aligns component layering with ARIA semantics; see [22](22-accessibility.md).

### Illustrative example — how the pieces combine

*Illustrative, not normative until elaboration.* Tokens (D7, tiered, `light-dark()` theming, relative-color derivation) → Tailwind utilities mapped onto the variables → components: native element + `tv()` variants with `className`/`style` excluded from the public API (D6) for simple components; headless ARIA layer (e.g. Ariakit `Dialog` with `role="alertdialog"`) + `tv()` slots for behavior-rich ones. Behavior-based names from the ARIA taxonomy ([22](22-accessibility.md)); fluid sizing (`min(92vw, 28rem)`, `text-balance`) per D14. Full worked code: [examples/modern-component.md](examples/modern-component.md).

### Icon system (SVG)

Sources: [SVGO](https://github.com/svg/svgo); trigger article on vite-svg-loader ([Managing SVG icons with vite-svg-loader](https://medium.com/@muhammederdinc/managing-svg-icons-effectively-using-vite-svg-loader-in-vue-vite-projects-990d2dc7ee02) — Vue-specific and unfetchable at collection time (429); points below are from the tools' own docs, adapted to React).

- **All icons are SVG**, optimized with **SVGO** in the build pipeline (strips metadata/editor cruft, configurable plugins) — unoptimized exports never reach the repo or the bundle.
- **Icons import as React components** (SVGR — e.g. `vite-plugin-svgr`, the React equivalent of the article's vite-svg-loader), with SVGO applied by the loader: inline SVG means no extra network requests, tree-shaking of unused icons, and full CSS control.
- **Icons inherit color via `currentColor`** — fills/strokes hardcoded in exports are rewritten, so icon color flows from text color and therefore from the design tokens (D7).
- **One icon set, one convention** — a single directory with a naming convention and a thin `Icon` wrapper component (sized via tokens); ad-hoc SVG imports scattered around the codebase are a defect (P5: one way to do each thing).
