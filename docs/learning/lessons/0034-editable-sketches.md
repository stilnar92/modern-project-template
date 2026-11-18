# 0034 — A sketch that people can change by hand

From PR #139 · 15.11.2026 · Topics: diagrams as code, tooling choice, local write endpoints

## Problem

The owner reworks the product and wants rough sketches of ideas: boxes and arrows that look hand-drawn. An agent draws the first version, and then the owner wants to move a box or change a word by hand. The project already had Mermaid: diagrams as text, drawn by an automatic layout. For a loop of five steps, the layout put the boxes in a chaotic order, and its hand-drawn look cut the labels. Nobody can drag a box in Mermaid.

## Predict first

A diagram is stored as a PNG picture in the repository. The owner wants to change one word in it next month. What is missing, and where must the missing part live so that it does not get lost?

## Theory

A diagram has two parts: the **picture** (what a reader sees) and the **source** (what an editor needs to change it). Tools differ in where they keep the source:

- **Diagram as code** (Mermaid, PlantUML, D2). The source is text, and the tool computes the layout. Diffs are easy to read, and an agent writes it well. But the author does not control positions, and a manual change means a text change.
- **Drawing tool** (Excalidraw, draw.io, Figma). The source is a scene: shapes with positions. The author drags shapes. A diff of the scene is hard to read.
- **Picture only** (PNG, plain SVG). Nothing to edit. The source is lost, or it lives in another place and drifts away.

A strong pattern is **the source inside the picture**. Excalidraw and draw.io can embed the scene in the SVG or PNG metadata. Then one file is both: every viewer shows the picture, and the editor opens the scene again. Nothing can drift, because there is only one file.

To edit in a web page and save to disk, a local server needs a **write endpoint**. Any web page in the same browser can send a request to `localhost`, so such an endpoint needs guards: the same-origin check, a secret token that only the real page knows, a strict allow-list of paths, and a body check.

Sources:

- Excalidraw docs, "Export utilities: exportToSvg, exportEmbedScene": <https://docs.excalidraw.com/docs/@excalidraw/excalidraw/api/utils/export>
- OWASP, "Cross-Site Request Forgery Prevention Cheat Sheet": <https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html>

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Mermaid with the hand-drawn look | Text in the Markdown, automatic layout | No hand edits, chaotic layout for loops |
| `.excalidraw.svg` + VS Code or Obsidian | One file, edited in a desktop editor | The owner leaves the docs site to edit |
| `.excalidraw.svg` + editor in the docs site | One file, "Edit sketch" button in dev, a guarded save endpoint | React in the viewer, a local write endpoint |
| excalidraw.com and manual export | Draw online, download, commit | Manual steps, the scene can be lost |

## Edge cases

| What if | What happens |
| --- | --- |
| Another web page posts to `/__sketch/save` | Rejected: foreign `Origin`, no token |
| A path has `../` or points to `core/` | Rejected: only `.excalidraw.svg` inside a published root folder |
| The site is built for a host | No editor and no endpoint. The sketch is a normal image |
| The owner opens the file in excalidraw.com | It opens: the scene is inside the SVG |
| Excalidraw changes its scene format | The scene has a version, and Excalidraw restores old ones. The SVG picture stays valid |

## Decision

We keep each sketch as one `.excalidraw.svg` file and edit it in the docs site, in dev only, through a guarded save endpoint ([ADR 0017](../../dev/adr/0017-sketch-editor.md)). The owner selected the in-site editor to change sketches where they read them. The decision would be wrong if the docs site ran on a shared host with this endpoint on, or if sketches became exact designs: then a design tool is better.

## In our code

1. `tools/docs-viewer/scripts/sketch-save.mjs`: the endpoint and its guards.
2. `tools/docs-viewer/.vitepress/theme/sketch.js`: finds sketch images and adds the button.
3. `tools/docs-viewer/.vitepress/theme/sketch-editor.js`: opens Excalidraw, exports the SVG with the scene, saves.
4. `tools/docs-viewer/README.md`, section "Sketches (Excalidraw)".

## Check yourself

1. Why is "the source inside the picture" safer than a picture plus a separate source file?
2. The endpoint checks `Origin` and a token. Which attack does each check stop?
3. Why does the static build not include the editor?

<details>
<summary>Answers</summary>

1. There is only one file, so the picture and the source cannot drift apart, and a copy or a move of the file keeps both.
2. `Origin` stops a request from another site in the same browser. The token stops a request that fakes or omits the header, for example from a local script or a tool that sets any header. Only the page that the dev server served knows the token.
3. A built site can go to a host. A write endpoint there lets other people change files. The editor is also heavy, and readers do not need it.

</details>
