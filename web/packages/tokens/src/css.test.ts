import { readFileSync } from "node:fs";
import { expect, test } from "vitest";
import { designSystemCss, designSystemTokens, tailwindCss, tokensCss } from "./css.ts";

const read = (name: string) => readFileSync(new URL(`../css/${name}`, import.meta.url), "utf8");

test("the committed CSS matches the tokens (run pnpm gen)", () => {
  expect(read("tokens.css")).toBe(tokensCss());
  expect(read("tailwind.css")).toBe(tailwindCss());
  const json = readFileSync(
    new URL("../design-system/project/tokens.json", import.meta.url),
    "utf8",
  );
  expect(json).toBe(designSystemTokens());
  const dsCss = readFileSync(
    new URL("../design-system/project/tokens.css", import.meta.url),
    "utf8",
  );
  expect(dsCss).toBe(designSystemCss());
});

test("a palette other than the default is a static block on <html>", () => {
  expect(tokensCss()).toMatch(/:root\[data-palette="teal"\] \{[\s\S]*--ms-accent-light: #087f5b;/);
});

test("Tailwind has no default colors, only tokens", () => {
  expect(tailwindCss()).toContain("--color-*: initial;");
  expect(tailwindCss()).not.toMatch(/--color-red/);
});

test("each semantic color has a fallback for browsers without light-dark()", () => {
  const css = tokensCss();
  expect(css).toContain("color-scheme: light dark;");
  expect(css).toMatch(
    /@media \(prefers-color-scheme: dark\)[\s\S]*--ms-surface: var\(--ms-surface-dark\)/,
  );
  expect(css).toMatch(
    /@supports \(color: light-dark\(white, black\)\)[\s\S]*--ms-surface: light-dark\(/,
  );
});

test("Tailwind utilities read the variables", () => {
  expect(tailwindCss()).toContain("--color-brand-600: var(--ms-brand-6);");
  expect(tailwindCss()).toContain("--color-surface: var(--ms-surface);");
});

// The Claude Design view shows at most 8 themes: a light and a dark one per
// palette. More palettes need a second view or a filter (ADR 0015).
test("the Claude Design view fits its theme limit", () => {
  const { color } = JSON.parse(designSystemTokens()) as { color: { themes: unknown[] } };
  expect(color.themes.length).toBeLessThanOrEqual(8);
});
