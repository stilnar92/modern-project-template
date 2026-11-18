// Builds the generated files of the tokens. build.ts writes them. A test
// checks that the committed files match.
import { font, gray, palettes, PREFIX, radius, SEMANTIC_USAGE, type Palette } from "./tokens.ts";

const HEADER = '/* Generated from packages/tokens/src/tokens.ts by "pnpm gen". Do not edit. */\n';

// Tailwind names its 10 shades 50, 100, ... 900.
const TAILWIND_STEPS = [50, 100, 200, 300, 400, 500, 600, 700, 800, 900];

const semanticNames = Object.keys(SEMANTIC_USAGE);

// The variables that a palette sets: its brand scale and the light and dark
// value of each semantic color.
function paletteLines(palette: Palette, indent: string): string[] {
  const lines = palette.brand.map((v, i) => `${indent}${PREFIX}-brand-${i}: ${v};`);
  for (const [name, { light, dark }] of Object.entries(palette.semantic)) {
    lines.push(
      `${indent}${PREFIX}-${name}-light: ${light};`,
      `${indent}${PREFIX}-${name}-dark: ${dark};`,
    );
  }
  return lines;
}

// tokensCss: the CSS variables for any page (modern-web-guidance "dark-mode").
// color-scheme on :root themes the browser UI. light-dark() is only Newly
// available, so the media query is the fallback for the Widely available
// target (frontend standard 10), and @supports upgrades it.
//
// A palette other than the default is a static block for
// <html data-palette="<id>">. It sets only the -light and -dark values. The
// semantic variables on :root read them, so the attribute must be on <html>
// (ADR 0015). All CSS is static: no inline style, which keeps the CSP strict.
export function tokensCss(): string {
  const lines: string[] = [];
  gray.forEach((v, i) => lines.push(`  ${PREFIX}-gray-${i}: ${v};`));
  lines.push(`  ${PREFIX}-font-sans: ${font.sans};`);
  for (const [name, v] of Object.entries(radius)) lines.push(`  ${PREFIX}-radius-${name}: ${v};`);
  lines.push(...paletteLines(palettes.default, "  "));
  for (const name of semanticNames)
    lines.push(`  ${PREFIX}-${name}: var(${PREFIX}-${name}-light);`);

  const dark = semanticNames.map((name) => `    ${PREFIX}-${name}: var(${PREFIX}-${name}-dark);`);
  const both = semanticNames.map(
    (name) =>
      `    ${PREFIX}-${name}: light-dark(var(${PREFIX}-${name}-light), var(${PREFIX}-${name}-dark));`,
  );

  const others = Object.entries(palettes)
    .filter(([id]) => id !== "default")
    .map(
      ([id, palette]) =>
        `\n:root[data-palette="${id}"] {\n${paletteLines(palette, "  ").join("\n")}\n}\n`,
    );

  return `${HEADER}:root {
  color-scheme: light dark;
${lines.join("\n")}
}

@media (prefers-color-scheme: dark) {
  :root {
${dark.join("\n")}
  }
}

@supports (color: light-dark(white, black)) {
  :root {
${both.join("\n")}
  }
}
${others.join("")}`;
}

// tailwindCss: the Tailwind 4 theme. "inline" makes each utility use the
// variable itself, so the color scheme and the palette switch without a
// rebuild. "--color-*: initial" removes the default Tailwind colors: a color
// class that is not a token (for example bg-red-500) produces no CSS.
export function tailwindCss(): string {
  const lines = ["  --color-*: initial;", "  --color-white: #ffffff;", "  --color-black: #000000;"];
  for (const name of ["brand", "gray"]) {
    TAILWIND_STEPS.forEach((step, i) =>
      lines.push(`  --color-${name}-${step}: var(${PREFIX}-${name}-${i});`),
    );
  }
  for (const name of semanticNames) lines.push(`  --color-${name}: var(${PREFIX}-${name});`);
  lines.push(`  --font-sans: var(${PREFIX}-font-sans);`);
  for (const name of Object.keys(radius))
    lines.push(`  --radius-${name}: var(${PREFIX}-radius-${name});`);
  return `${HEADER}@theme inline {
${lines.join("\n")}
}
`;
}

// designSystemTokens: tokens.json of the Claude Design view (ADR 0015). It is
// the flat list format of the "Design System" artifact type, not DTCG. Each
// palette gives a light and a dark theme, so the view can switch between
// them. The first theme is the default palette in the light scheme.
export function designSystemTokens(): string {
  const themes = Object.keys(palettes).flatMap((id) =>
    (["light", "dark"] as const).map((mode) => ({
      id: `${id}-${mode}`,
      name: `${id[0]!.toUpperCase()}${id.slice(1)} ${mode}`,
    })),
  );
  const perTheme = (pick: (p: Palette, mode: "light" | "dark") => string) =>
    Object.fromEntries(
      Object.entries(palettes).flatMap(([id, p]) =>
        (["light", "dark"] as const).map((mode) => [`${id}-${mode}`, pick(p, mode)]),
      ),
    );

  const colorTokens = [
    ...semanticNames.map((name) => ({
      name,
      value: perTheme((p, mode) => p.semantic[name as keyof Palette["semantic"]][mode]),
      usage: SEMANTIC_USAGE[name as keyof typeof SEMANTIC_USAGE],
    })),
    ...TAILWIND_STEPS.map((step, i) => ({
      name: `brand-${step}`,
      value: perTheme((p) => p.brand[i]!),
      usage: "Brand scale of the palette. Use a semantic color in components.",
    })),
    ...TAILWIND_STEPS.map((step, i) => ({
      name: `gray-${step}`,
      value: gray[i],
      usage: "Neutral scale. Use a semantic color in components.",
    })),
  ];

  const tokens = {
    name: "modern-project-template",
    version: 1,
    meta: { source: "web/packages/tokens/src/tokens.ts" },
    color: { themes, tokens: colorTokens },
    type: {
      fonts: [],
      families: { sans: font.sans },
      // The Tailwind 4 text sizes that the pages use (no custom type scale yet).
      groups: [
        {
          name: "Text",
          family: "sans",
          styles: [
            { name: "text-sm", fontSize: "14px", lineHeight: "20px", fontWeight: 400 },
            { name: "text-base", fontSize: "16px", lineHeight: "24px", fontWeight: 400 },
            { name: "text-xl", fontSize: "20px", lineHeight: "28px", fontWeight: 600 },
            { name: "text-3xl", fontSize: "30px", lineHeight: "36px", fontWeight: 600 },
          ],
        },
      ],
    },
    radius: {
      tokens: Object.entries(radius).map(([name, value]) => ({
        name: `radius-${name}`,
        value,
        usage: {
          sm: "Small controls, tags.",
          md: "Buttons, inputs, cards (default).",
          lg: "Large surfaces, sheets.",
        }[name],
      })),
    },
  };
  return `${JSON.stringify(tokens, null, 2)}\n`;
}

// designSystemCss: tokens.css of the Claude Design view, in the shape that the
// view compiles itself: one block per theme on [data-theme], then the other
// tokens on :root. A view synced from code does not compile it, so we send it.
export function designSystemCss(): string {
  const {
    color,
    radius: r,
    type,
  } = JSON.parse(designSystemTokens()) as {
    color: {
      themes: { id: string }[];
      tokens: { name: string; value: string | Record<string, string> }[];
    };
    radius: { tokens: { name: string; value: string }[] };
    type: { families: Record<string, string> };
  };
  const blocks = color.themes.map(({ id }, i) => {
    const selector = i === 0 ? `:root,\n[data-theme="${id}"]` : `[data-theme="${id}"]`;
    const vars = color.tokens.map(
      (t) => `  --${t.name}: ${typeof t.value === "string" ? t.value : t.value[id]};`,
    );
    return `${selector} {\n${vars.join("\n")}\n}\n`;
  });
  const rest = [
    ...r.tokens.map((t) => `  --${t.name}: ${t.value};`),
    ...Object.entries(type.families).map(([k, v]) => `  --font-${k}: ${v};`),
  ];
  return `/* modern-project-template — generated from tokens.json */\n${blocks.join("\n")}\n:root {\n${rest.join("\n")}\n}\n`;
}
