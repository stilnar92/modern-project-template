// The design tokens of all frontends (frontend standard 19, D7, ADR 0012,
// ADR 0015). One source for the Mantine theme of the dashboard, the CSS
// variables of the site and the Claude Design view. "pnpm gen" writes
// css/ and design-system/ from this file. Do not edit those files by hand.
//
// The values are placeholders (the default palettes of Mantine) until the
// design sets the brand. Change only the values, not the names.

// A scale of 10 shades, light to dark. Mantine needs exactly 10.
export type Shades = readonly [
  string,
  string,
  string,
  string,
  string,
  string,
  string,
  string,
  string,
  string,
];

// A value for each color scheme: the page follows the scheme of the system.
export type Modes = { readonly light: string; readonly dark: string };

// Tier 1: primitive values, named for what they are.
export const gray = [
  "#f8f9fa",
  "#f1f3f5",
  "#e9ecef",
  "#dee2e6",
  "#ced4da",
  "#adb5bd",
  "#868e96",
  "#495057",
  "#343a40",
  "#212529",
] as const satisfies Shades;

export const font = {
  sans: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
} as const;

export const radius = { sm: "0.25rem", md: "0.5rem", lg: "1rem" } as const;

// Tier 2: semantic colors, named for their intent. Each palette gives all of
// them. The text says where to use each one (also shown in Claude Design).
export const SEMANTIC_USAGE = {
  surface: "Page background.",
  text: "Body text and headings on surface.",
  "text-muted": "Secondary text on surface: hints, dates, captions.",
  border: "Borders of cards, inputs and dividers on surface.",
  accent: "Links, primary buttons, the selected state. Text in accent is allowed on surface.",
  "on-accent": "Text and icons on an accent background, for example a primary button.",
} as const;

export type SemanticName = keyof typeof SEMANTIC_USAGE;

// The pairs that must be readable: [text color, background]. The contrast
// test checks each pair in each palette and scheme (WCAG 1.4.3 AA, 4.5:1).
export const TEXT_PAIRS = [
  ["text", "surface"],
  ["text-muted", "surface"],
  ["accent", "surface"],
  ["on-accent", "accent"],
] as const satisfies readonly (readonly [SemanticName, SemanticName])[];

// A palette is one look of a page: a brand scale and all semantic colors.
// The site selects it with <html data-palette="<id>"> (ADR 0015).
// - "public": a user can select it (self-serve).
// - "private": a paid individual design. The record of which user owns it
//   is data in the core, never in this file.
export type Palette = {
  readonly visibility: "public" | "private";
  readonly brand: Shades;
  readonly semantic: { readonly [K in SemanticName]: Modes };
};

const blue = [
  "#e7f5ff",
  "#d0ebff",
  "#a5d8ff",
  "#74c0fc",
  "#4dabf7",
  "#339af0",
  "#228be6",
  "#1c7ed6",
  "#1971c2",
  "#1864ab",
] as const satisfies Shades;

const teal = [
  "#e6fcf5",
  "#c3fae8",
  "#96f2d7",
  "#63e6be",
  "#38d9a9",
  "#20c997",
  "#12b886",
  "#0ca678",
  "#099268",
  "#087f5b",
] as const satisfies Shades;

// The neutral colors are the same in each palette.
const neutral = {
  surface: { light: "#ffffff", dark: gray[9] },
  text: { light: gray[9], dark: gray[0] },
  "text-muted": { light: gray[7], dark: gray[4] },
  border: { light: gray[3], dark: gray[7] },
} as const;

export const palettes = {
  default: {
    visibility: "public",
    brand: blue,
    semantic: {
      ...neutral,
      accent: { light: blue[8], dark: blue[3] },
      "on-accent": { light: "#ffffff", dark: gray[9] },
    },
  },
  teal: {
    visibility: "public",
    brand: teal,
    semantic: {
      ...neutral,
      accent: { light: teal[9], dark: teal[3] },
      "on-accent": { light: "#ffffff", dark: gray[9] },
    },
  },
} as const satisfies Record<string, Palette> & { default: Palette };

export type PaletteId = keyof typeof palettes;

// The default palette: the dashboard uses only this one.
export const color = { brand: palettes.default.brand, gray } as const;
export const semantic = palettes.default.semantic;

// The prefix of all generated CSS variables.
export const PREFIX = "--ms";
