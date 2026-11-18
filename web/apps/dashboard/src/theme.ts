// The Mantine theme of the dashboard from the shared tokens (frontend
// standard 19, ADR 0012). Mantine needs literal colors, not CSS variables,
// so it reads the token values directly.
import { color, font, radius, semantic } from "@app/tokens";
import { createTheme, type CSSVariablesResolver } from "@mantine/core";

export const theme = createTheme({
  primaryColor: "brand",
  colors: { brand: color.brand, gray: color.gray },
  fontFamily: font.sans,
  headings: { fontFamily: font.sans },
  radius,
  defaultRadius: "md",
});

// The semantic tokens set the page colors of Mantine in both schemes, so
// the dashboard and the public pages have the same surface and text.
const schemeVariables = (mode: "light" | "dark") => ({
  "--mantine-color-body": semantic.surface[mode],
  "--mantine-color-text": semantic.text[mode],
  "--mantine-color-dimmed": semantic["text-muted"][mode],
  "--mantine-color-default-border": semantic.border[mode],
});

export const cssVariablesResolver: CSSVariablesResolver = () => ({
  variables: {},
  light: schemeVariables("light"),
  dark: schemeVariables("dark"),
});
