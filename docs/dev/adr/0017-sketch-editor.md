# 0017 — Sketch editor in the docs viewer: Excalidraw in `.excalidraw.svg` files

Date: 2026-11-15
Status: accepted

## Context

The owner reworks the product and wants rough sketches of ideas: boxes and arrows that look hand-drawn, not like a finished design. The owner wants to change a sketch by hand, and an agent must be able to draw the first version.

Mermaid ([ADR 0016](0016-docs-viewer.md)) does not fit. The owner cannot move a box by hand, the automatic layout of a loop is chaotic, and its hand-drawn look cut the labels.

The owner selected the Excalidraw editor inside the docs site, from three options: a `.excalidraw.svg` file edited in VS Code, the editor in the docs site, and Obsidian with a plugin.

ADR 0016 says that the viewer never changes a document. This editor writes files, so it needs its own decision.

## Decision

1. **A sketch is one file `<name>.excalidraw.svg`** in a published root. It is a normal SVG, so GitHub, the built site and `docs:check` show it as an image. The Excalidraw scene is embedded in the SVG, so the editor opens it again. A document links it as a normal image: `![...](../images/<name>.excalidraw.svg)`.
2. **The editor is `@excalidraw/excalidraw` 0.18.1 with React 19**, in the docs viewer only. It loads only in `vitepress dev` (`mise run docs:dev`), and only on the first click. The static build has no editor and no endpoint.
3. **In dev, each sketch image gets an "Edit sketch" button.** "Save" exports the SVG with the scene and sends it to the dev-server endpoint `/__sketch/save`, which writes the file.
4. **The endpoint accepts a write only when all guards pass:** a random token for each server start, an `Origin` of the same host, the suffix `.excalidraw.svg`, a path inside a published root folder with an existing parent folder, and a body that starts with `<svg`, at most 10 MB.
5. **Fonts come from the installed package**, not from a CDN. The saved SVG embeds the font subset, so it shows the same everywhere.
6. **An agent draws the first version** through `window.__sketch` (the Excalidraw API and `convertToExcalidrawElements`) while the editor is open, and then saves. The owner changes it by hand.

## Trade-off

- **Prefer:** Simplicity for the owner (edit in the page that shows the sketch), Correctness (one file is the picture and the source), Cost (no SaaS, no new service).
- **Give up:** Security (a dev server that can write files, local only, behind guards), Size (React and Excalidraw in the viewer packages, loaded only in dev).

## Stress round

| Stressor | Affects | Verdict | Resolution |
| --- | --- | --- | --- |
| A hostile web page sends a request to the dev server | files in the repository | survives | The request has a foreign `Origin` or no token. The token is only in the page that the dev server serves |
| A request tries `../` or a path outside the roots | files outside `docs/` | survives | The path must start with a published root folder and stay inside the repository after `resolve` |
| Excalidraw changes its file format | old sketches | accepted | The scene has a format version, and Excalidraw restores old versions. The SVG picture stays valid in any case |
| The owner edits a sketch in excalidraw.com or VS Code | the workflow | survives | Both open and save `.excalidraw.svg` with an embedded scene |
| A sketch has a large image inside | the repository size | accepted | The 10 MB limit. A review sees the file size |

## Consequences

- Easier: the owner changes a sketch in the docs site and saves it to the repository. An agent draws a first version in the same format.
- Harder: the viewer has React next to Vue, only in dev.
- Forbidden: a write endpoint in the static build, and a write to a file that is not `.excalidraw.svg`.
