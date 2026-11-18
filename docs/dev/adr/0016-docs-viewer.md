# 0016 — Docs viewer: a VitePress site over the Markdown in the repository

Date: 2026-11-15
Status: accepted

## Context

The project has about 160 Markdown documents: product, standards, ADRs, lessons and specs. Agents read them as files. The owner reads them in an editor, where diagrams and the links between documents are hard to follow. The owner wants a comfortable site: navigation, search, Mermaid diagrams, drawings, the HTTP API, and links between documents.

The requirements:

- The Markdown in git stays the only source. No SaaS and no copy of the text.
- No change to the documents: they have an H1 and no title frontmatter, and about 270 links have GitHub heading anchors, 68 of them Cyrillic.
- Local first, with a mise task, and a static build that a host can serve later.
- No new language runtime, no extra work in the CI of other pull requests, and `mise run check` stays green.

A study with the `find-solution` loop ran 3 rounds. It compared Quartz, Astro Starlight, Hugo with Hextra, VitePress, Docusaurus, MkDocs, Obsidian and a VS Code workspace. Only VitePress was built and checked on the real repository: all 800 internal links and all 243 anchors worked in the HTML, and Mermaid, search and the API page worked with the network blocked. The other candidates had failures that the study found in their sources:

- Starlight requires title frontmatter, and its link and graph plugins do not support Astro 7 or files outside `src/content`.
- Quartz loads Mermaid from a CDN and needs a vendored fork.
- Hugo needs our own templates for the title, backlinks and the graph.
- Obsidian does not follow GitHub anchors.
- MkDocs needs Python.

## Decision

1. **VitePress 1.6.4 in `tools/docs-viewer/`**, with its own `package.json` and `pnpm-lock.yaml`, outside the web workspace. Each package is at least 14 days old (`minimumReleaseAge`).
2. **The published roots are an allow-list:** `docs/`, `openspec/`, `api/` and `CONTEXT.md`. A script links them into the ignored folder `content/` before each build. Git does not track the symlinks, so the linters see each document once.
3. **The documents do not change.** The viewer config takes the title from the H1, makes heading ids with `github-slugger`, sends links outside the roots to GitHub, and shows unknown HTML tags as text.
4. **The API page** reads `api/openapi.yaml` at build time with Scalar.
5. **Commands:** `mise run docs` (build and a static preview on `127.0.0.1` only), `mise run docs:dev` (live reload) and `mise run docs:check` (build and an HTML check of each link, anchor, image and title). A CI job runs `docs:check` only when a document or the viewer changes.
6. **Backlinks and a link graph are not part of this decision.** No established VitePress plugin gives them. A later change can make a backlinks list from the links that the build already reads.
7. **Accepted vulnerabilities:** VitePress 1.x needs Vite 5, which has 10 advisories, mostly for the dev server. They are in `osv-scanner.toml` with reasons and the end date 2027-02-14, and `audit:deps` scans the viewer lockfile. The details are in [the viewer README](../../../tools/docs-viewer/README.md#known-vulnerabilities-accepted).

## Trade-off

- **Prefer:** Correctness (each link and anchor checked), Simplicity (no change to the documents, one small config), Cost (one install, a fast build).
- **Give up:** Security (a dev server with known advisories, local only), Evolvability (no graph and no backlinks, VitePress 1.x is in maintenance).

## Stress round

| Stressor | Affects | Verdict | Resolution |
| --- | --- | --- | --- |
| A document gets a raw `<tag>` or `{{ }}` | the whole build | survives | The safe-HTML rule shows them as text. `docs:check` in CI fails on a new case, before the merge |
| A hostile web page reads files from the dev server | the owner's machine | accepted | The dev server reads only the viewer and the roots, never `.env`. Daily use is the static preview |
| VitePress 2 stays in alpha for a long time | upgrades | accepted | The viewer is isolated in its own lockfile. On the review date, decide again |
| The number of documents grows ten times | build time, sidebar | survives | The build takes about 30 s now. The sidebar groups are collapsed |
| A new root, such as `core/README.md` | the allow-list | survives | One line in `scripts/roots.mjs` and the CI filter |

## Consequences

- Easier: the owner reads all documents as one site, with search and diagrams. A broken link or anchor fails CI.
- Harder: a document with unusual HTML must pass the viewer build too.
- Forbidden: a copy of a document inside the viewer, a document edit only to please the viewer, and a published path outside the allow-list.
