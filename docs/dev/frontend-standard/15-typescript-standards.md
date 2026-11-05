# 15 — TypeScript Standards

**Status:** COLLECTING — raw points being gathered; will be elaborated into full rules later.

---

## Collected points

### From Locked Decisions

- **D11** — typechecking runs on **native TypeScript**. Since TypeScript 7 this is the stable `tsc` (the former tsgo); before it, `@typescript/native-preview` and the `tsgo` command. Practical notes:
  - install: `typescript@7`. The `@typescript/native-preview` dev builds stopped after the stable release.
  - remove legacy flags from tsconfig
  - TypeScript 7 has no classic JavaScript API. A tool that imports `typescript` as a library keeps TypeScript 6 in its own package: in this project `openapi-typescript` (in `packages/api-client`) and `@astrojs/check` (in `apps/site`).
  - editor: `"typescript.experimental.useTsgo": true` in VS Code settings (preview-era setting, not checked with TypeScript 7)
  - the native checker reportedly catches some errors the old tsc missed; treat its output as the source of truth.
