// Writes the generated files of the tokens: "pnpm gen" (part of "mise run gen").
import { writeFileSync } from "node:fs";
import { designSystemCss, designSystemTokens, tailwindCss, tokensCss } from "./css.ts";

const css = new URL("../css/", import.meta.url);
writeFileSync(new URL("tokens.css", css), tokensCss());
writeFileSync(new URL("tailwind.css", css), tailwindCss());
writeFileSync(
  new URL("../design-system/project/tokens.json", import.meta.url),
  designSystemTokens(),
);
writeFileSync(new URL("../design-system/project/tokens.css", import.meta.url), designSystemCss());
