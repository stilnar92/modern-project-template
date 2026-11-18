import { color, semantic } from "@app/tokens";
import { DEFAULT_THEME, mergeMantineTheme } from "@mantine/core";
import { expect, test } from "vitest";
import { cssVariablesResolver, theme } from "./theme";

test("the Mantine theme uses the shared tokens", () => {
  expect(theme.primaryColor).toBe("brand");
  expect(theme.colors?.brand).toEqual(color.brand);
});

test("the page colors of both schemes come from the semantic tokens", () => {
  const vars = cssVariablesResolver(mergeMantineTheme(DEFAULT_THEME, theme));
  expect(vars.light["--mantine-color-body"]).toBe(semantic.surface.light);
  expect(vars.dark["--mantine-color-body"]).toBe(semantic.surface.dark);
  expect(vars.dark["--mantine-color-text"]).toBe(semantic.text.dark);
});
