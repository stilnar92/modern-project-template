# 0033 — People and agents must read the same documents

From PR #127 · 15.11.2026 · Topics: docs as code, tooling choice, link checks

## Problem

The project keeps its documents as Markdown in git. Agents read them as files, and that works well. The owner reads them in an editor, where Mermaid shows as code, a drawing is a file name and the links between documents are hard to follow. The owner wants a comfortable site. But a second copy of the text, in a wiki or a SaaS tool, starts to drift from the files on the first day. The documents also have about 270 links to headings, 68 of them with Cyrillic anchors, and nobody wants to change 160 files for a viewer.

## Predict first

Two site tools show the same Markdown file with the heading "Открытые вопросы". One tool makes the id `открытые-вопросы`, the other makes `открытые-вопросы-1` or drops the Cyrillic letters. The link check in CI reads only the Markdown files. Does CI find the broken links on the site? Why, or why not?

## Theory

**Docs as code** means: the documents live in the repository next to the code, in plain text, and the same tools check them: review, linters, CI. A site is only a **view** of these files. The view must not become a second source. The same idea applies to diagrams as code (Mermaid) and to a contract file (OpenAPI) that tools show as pages.

A view adds its own rules, and each rule can break a document:

- **Titles.** Many tools want a title in frontmatter. Files with only an H1 then show no title or two titles.
- **Heading ids (slugs).** A link `file.md#section` works only if the viewer makes the same id as the place where the author checked the link. Slug algorithms differ on Unicode letters, dashes and punctuation.
- **Link paths.** A relative link to `../x.md` must become a route on the site. A link to a file outside the site must go somewhere useful.
- **Templates.** Some tools compile Markdown as a template (Vue, MDX). Then text such as `<placeholder>` or `{{ }}` is code, not text.

So "the tool supports Markdown" says little. The useful question is: **does the real build of our files keep each link and anchor?** Only a build of the real repository and a check of its HTML answer it. A check of the source alone (lychee on the Markdown) does not see what the viewer breaks.

A second lesson from the choice: **measurements beat opinions**. Three rounds of comparison on paper gave a different ranking each round. One prototype on the real files gave counts that nobody could argue with.

Sources:

- [Docs as code](https://www.writethedocs.org/guide/docs-as-code/), Write the Docs: the idea and the workflow.
- [Markdown extensions: header anchors](https://vitepress.dev/guide/markdown), VitePress: how a viewer makes heading ids.
- [github-slugger](https://github.com/Flet/github-slugger): the slug algorithm of GitHub as a library.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| A wiki or SaaS (Notion, Confluence) | A second copy of the text | It drifts from git. Agents need a connector |
| An editor or a desktop app (VS Code, Obsidian) | The owner opens the files | Not a site. Obsidian does not follow GitHub anchors |
| A site generator with built-in graph (Quartz) | Builds the files, shows backlinks and a graph | A fork to keep, Mermaid from a CDN, custom title code |
| A docs framework of our web stack (Starlight) | A site in the Astro workspace | Title frontmatter in each file, plugins without Astro 7 support |
| A small site generator (VitePress) | Builds the files with a small config | No graph, an older Vite line |

## Edge cases

| What if | What happens |
| --- | --- |
| A document has `<Decision title>` as a placeholder | Vue fails on the whole build. A rule shows unknown tags as text |
| A heading has Cyrillic letters and an en dash | The viewer uses `github-slugger`, so the id is the same as on GitHub. `docs:check` checks all 68 Cyrillic anchors |
| A document links to `CLAUDE.md`, which the site does not publish | The link opens the file on GitHub |
| The viewer packages get a security advisory | `audit:deps` scans the viewer lockfile too. The 10 known dev-server advisories of Vite 5 are accepted in `osv-scanner.toml` with an end date. A new one fails the check |
| Two worktrees run `mise run docs` | `vite preview` takes the next free port. It listens on `127.0.0.1` only, so nobody on the same network can open the documents |
| The tracked symlinks point to `CONTEXT.md` | lychee checks the symlink from its own folder and fails. So the symlinks are made before each build and git ignores them |

## Decision

VitePress 1.6.4 in `tools/docs-viewer/`, over an allow-list of roots, with no change to the documents. It was the only option that a real build proved: 800 of 800 internal links and 243 of 243 anchors worked. The choice would be wrong if the owner needs a link graph more than correct links, or if the docs move to MDX components. See [ADR 0016](../../dev/adr/0016-docs-viewer.md).

## In our code

1. `tools/docs-viewer/scripts/roots.mjs`: the allow-list.
2. `tools/docs-viewer/.vitepress/config.mjs`: the title, slug, link and safe-HTML rules.
3. `tools/docs-viewer/scripts/check-html.mjs`: the check of the built HTML.
4. `mise.toml`: the tasks `docs`, `docs:dev` and `docs:check`.

## Check yourself

1. lychee already checks every link in the Markdown. Why does CI also need a check of the built HTML?
2. Why is a tracked symlink to `CONTEXT.md` a problem, when a tracked symlink to `docs/` is not?
3. A colleague proposes to add `title:` frontmatter to all documents, so that any docs tool works. What does the project lose?

<details>
<summary>Answers</summary>

1. lychee checks the source. The viewer can change ids, routes and HTML. Only the built HTML shows what a reader really gets.
2. `git ls-files '*.md'` lists a file symlink as a Markdown file, so lychee checks it from its own folder, where the relative links point to the wrong place. A folder symlink is not a `.md` file, so the linters skip it. The project avoids both: it makes the symlinks at build time.
3. A second place for the title, which can differ from the H1. A large change with no value for agents or for GitHub. And the documents start to serve one tool.

</details>
