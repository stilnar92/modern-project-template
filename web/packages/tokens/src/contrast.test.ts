import { expect, test } from "vitest";
import { palettes, TEXT_PAIRS } from "./tokens.ts";

// WCAG 2.2 relative luminance and contrast ratio for #rrggbb colors.
function luminance(hex: string): number {
  const [r, g, b] = [1, 3, 5].map((i) => {
    const c = Number.parseInt(hex.slice(i, i + 2), 16) / 255;
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  });
  return 0.2126 * r! + 0.7152 * g! + 0.0722 * b!;
}

function contrast(a: string, b: string): number {
  const [hi, lo] = [luminance(a), luminance(b)].sort((x, y) => y - x);
  return (hi! + 0.05) / (lo! + 0.05);
}

// Each text pair has at least 4.5:1 in each palette and both schemes (WCAG
// 1.4.3 AA, UX standard). The palettes are a closed list, so this checks
// every look that a page can have. A new palette cannot break it.
const cases = Object.entries(palettes).flatMap(([id, palette]) =>
  TEXT_PAIRS.map(([fg, bg]) => [id, fg, bg, palette] as const),
);

test.each(cases)("%s: %s has AA contrast on %s", (_id, fg, bg, palette) => {
  for (const mode of ["light", "dark"] as const) {
    const ratio = contrast(palette.semantic[fg][mode], palette.semantic[bg][mode]);
    expect(ratio, `${mode}: ${ratio.toFixed(2)}`).toBeGreaterThanOrEqual(4.5);
  }
});

test("each color is a #rrggbb value, so the test can read it", () => {
  for (const palette of Object.values(palettes)) {
    for (const v of [
      ...palette.brand,
      ...Object.values(palette.semantic).flatMap((m) => [m.light, m.dark]),
    ]) {
      expect(v).toMatch(/^#[0-9a-f]{6}$/);
    }
  }
});
