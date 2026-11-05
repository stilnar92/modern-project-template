# Example — Modern Component Stack

*Illustrative reference for [20 — Component System](../20-component-system.md); not normative until elaboration.*

Demonstrates how the standard's pieces combine: design tokens as CSS variables (D7, tiers per [19](../19-design-tokens.md)), Tailwind + tailwind-variants as the only styling surface (D5/D6), native-first and headless-ARIA component sourcing ([20](../20-component-system.md), [22](../22-accessibility.md)), fluid-first sizing (D14), and modern CSS defaults (D2).

## Layer 1 — Tokens (`tokens.css`)

```css
:root {
  color-scheme: light dark;              /* native widgets follow the theme */

  /* 1 — primitives: what it is */
  --blue-500: oklch(62% 0.19 255);
  --blue-300: oklch(75% 0.13 255);
  --red-500:  oklch(58% 0.22 25);
  --gray-100: oklch(96% 0 0);
  --gray-900: oklch(21% 0 0);

  /* 2 — semantic: why it exists (intent, never appearance) */
  --color-accent:       light-dark(var(--blue-500), var(--blue-300));
  --color-accent-hover: oklch(from var(--color-accent) calc(l - 0.06) c h);
  --color-danger:       var(--red-500);
  --color-surface:      light-dark(#fff, var(--gray-900));
  --color-text:         light-dark(var(--gray-900), var(--gray-100));

  /* scales: rem-based (units policy), numeric with growth room */
  --radius-control: 0.5rem;
  --space-200: 0.5rem;
  --space-400: 1rem;
}
```

- `light-dark()` — theming with zero JS; `color-scheme` makes scrollbars/form controls follow.
- Hover color **derived** via relative color syntax — change one token, the family follows.
- Tailwind (v4 CSS-first) maps utilities onto these variables via `@theme`, so `bg-accent` / `px-400` can never contain raw values.

## Layer 2 — Simple component: native element + variants-only API (`button.tsx`)

```tsx
import { tv, type VariantProps } from "tailwind-variants";

const button = tv({
  base: [
    "inline-flex items-center justify-center gap-200",
    "rounded-control font-medium select-none transition-colors",
    "outline-offset-2 focus-visible:outline-2 focus-visible:outline-accent",
    "disabled:opacity-50 disabled:pointer-events-none",
  ],
  variants: {
    intent: {
      primary: "bg-accent text-white hover:bg-accent-hover",
      danger:  "bg-danger text-white hover:bg-danger/90",
      ghost:   "bg-transparent text-text hover:bg-text/10",
    },
    size: {
      sm: "h-8 px-200 text-sm",
      md: "h-10 px-400 text-base",
    },
  },
  defaultVariants: { intent: "primary", size: "md" },
});

type ButtonProps = VariantProps<typeof button> &
  Omit<React.ComponentPropsWithRef<"button">, "className" | "style">;
  // ^ className/style excluded from the public API — variants only (D6)

export function Button({ intent, size, type = "button", ...props }: ButtonProps) {
  return <button type={type} className={button({ intent, size })} {...props} />;
}
```

- Native `<button>`: role/keyboard/focus semantics for free (native-first).
- `focus-visible` + `outline-offset` from the modern-CSS one-liners.
- The type system enforces D6: consumers pick variants, cannot inject classes.

## Layer 3 — Behavior-rich component: headless ARIA + slots (`confirm-dialog.tsx`)

```tsx
import * as Ariakit from "@ariakit/react";
import { tv } from "tailwind-variants";
import { Button } from "./button";

const dialog = tv({
  slots: {
    backdrop: "fixed inset-0 bg-gray-900/40 backdrop-blur-sm",
    panel: [
      "fixed inset-0 m-auto h-fit w-[min(92vw,28rem)]",   // fluid-first (D14)
      "rounded-control bg-surface text-text p-400 shadow-lg",
    ],
    title:   "text-lg font-semibold text-balance",
    actions: "mt-400 flex justify-end gap-200",
  },
});
const s = dialog();

export function ConfirmDialog({ store, title, children, onConfirm }: Props) {
  return (
    <Ariakit.Dialog
      store={store}
      role="alertdialog"                    // ARIA taxonomy: requires a response
      backdrop={<div className={s.backdrop()} />}
      className={s.panel()}
    >
      <Ariakit.DialogHeading className={s.title()}>{title}</Ariakit.DialogHeading>
      {children}
      <div className={s.actions()}>
        <Ariakit.DialogDismiss render={<Button intent="ghost" />}>Cancel</Ariakit.DialogDismiss>
        <Button intent="danger" onClick={onConfirm}>Delete</Button>
      </div>
    </Ariakit.Dialog>
  );
}
```

- **Ariakit owns the ARIA contract** — focus lock, Escape, `aria-labelledby` via internal `useId` (instance-proof).
- **`tv` slots own every visual**; visuals speak only in tokens.
- Named `ConfirmDialog` with `role="alertdialog"` — behavior-based naming, not "Modal".
