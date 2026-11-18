// The style variants of the site components (frontend standard 20, D6):
// components take variants, not raw class strings. tailwind-variants runs on
// the server here, so it sends no JavaScript to the browser.
import { tv } from "tailwind-variants";

export const page = tv({ base: "mx-auto max-w-prose px-4 py-8" });

export const heading = tv({
  base: "font-semibold text-text",
  variants: { level: { 1: "text-3xl", 2: "text-xl" } },
  defaultVariants: { level: 1 },
});

export const status = tv({
  base: "mt-4 inline-block rounded-md border px-3 py-1 text-sm",
  variants: {
    state: { ok: "border-accent text-accent", unavailable: "border-border text-text-muted" },
  },
});
