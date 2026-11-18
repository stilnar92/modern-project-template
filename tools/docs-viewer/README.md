# Docs viewer

A local site for the project documents ([ADR 0016](../../docs/dev/adr/0016-docs-viewer.md)). The Markdown files in `docs/`, `openspec/`, `api/` and `CONTEXT.md` stay the source. Agents read these files. The site only shows them, with navigation, search, Mermaid diagrams, Excalidraw sketches and the HTTP API.

## Commands

| Command | What it does |
| --- | --- |
| `mise run docs` | Build the site and serve it on <http://127.0.0.1:4173>. Another worktree gets the next free port. Only this computer can open it |
| `mise run docs:dev` | The site with live reload while you edit a document |
| `mise run docs:check` | Build the site and check each link, anchor, image and title in the HTML. CI runs it when a document or the viewer changes |

The first command in a worktree installs the packages once. The next runs work without the network.

## How it works

1. `scripts/link-content.mjs` makes `content/`. It holds relative symlinks to the published roots (`scripts/roots.mjs`), the home page and the API page. Git ignores `content/`, so the linters see each document only once.
2. VitePress builds `content/` into `.vitepress/dist/`. `.vitepress/config.mjs` sets these rules:
   - The title of a page is the H1 of the file. No document needs frontmatter.
   - Heading ids come from `github-slugger`, the same ids as on GitHub. So the links with Cyrillic anchors work.
   - A link to a file outside the published roots, to a folder without a README or to a file that is not Markdown opens on GitHub.
   - An unknown HTML tag in a document, such as `<placeholder>`, shows as text. Vue would fail on it.
3. The API page reads `api/openapi.yaml` at build time (Scalar). Telemetry and the request client are off.

To add a published root, change `ROOTS` in `scripts/roots.mjs` and the `docs` filter in `.github/workflows/ci.yml`.

## Sketches (Excalidraw)

A sketch is one file `<name>.excalidraw.svg` in a published root, for example `docs/images/`. It is an SVG with the Excalidraw scene inside ([ADR 0017](../../docs/dev/adr/0017-sketch-editor.md)). Link it from a document as a normal image.

To change a sketch:

1. Run `mise run docs:dev` and open the page.
2. Click **Edit sketch** on the image. Excalidraw opens in full screen.
3. Draw, then click **Save**. The file in the repository changes, and the page shows the new picture.

To make a new sketch, add a placeholder SVG with the `.excalidraw.svg` suffix and link it. The editor opens it empty.

An agent can draw the first version: open the editor in a browser, call `window.__sketch.api.updateScene({ elements: window.__sketch.convertToExcalidrawElements([...]) })`, then click **Save**.

## Known vulnerabilities (accepted)

VitePress 1.6.4 needs Vite 5. On 2026-11-15, osv-scanner finds 10 advisories in this lockfile: 8 in vite 5.4.21, 1 in esbuild 0.21.5 and 1 in katex 0.16.47. The fixes are in Vite 6 and later. VitePress 2 supports Vite 7, but VitePress 2 is still an alpha release.

The owner accepted them until **2027-02-14**. Each one is in `osv-scanner.toml` with its reason and that end date. `mise run audit:deps` scans this lockfile too, so a new advisory, for example in Scalar or mermaid, fails the check like in the other lockfiles. The reasons:

- The vite and esbuild advisories affect the dev server only (`mise run docs:dev`). The dev server listens on `localhost` and reads only the viewer and the published roots, never `.env`. `mise run docs` serves the static build with `vite preview` on `127.0.0.1`.
- The katex advisory needs a hostile math formula in a document. The documents come from this repository only.

On the review date, move to VitePress 2 if it is stable, and run `mise run docs:check` again.
