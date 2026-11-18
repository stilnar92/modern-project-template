# 0015 — Design system in the repository, Claude Design as a view

Date: 2026-11-12
Status: accepted

## Context

The owner uses Claude Design for the design system and for prototypes of screens ([dev process, step 3](../dev-process.md#cycle-for-each-task)). Three forces apply:

- Access to Claude can stop. Codex or a person must then continue with the files of the repository only.
- The project needs fast prototypes and a stable design system at the same time.
- The two frontends need different freedom. The dashboard has one strict look. On the site, the look changes from page to page. A product can offer a self-serve choice of look and a paid individual design.

[ADR 0012](0012-infrastructure-libraries.md) already sets the token source: our own TypeScript package `web/packages/tokens`. The project does not use Figma. A study with the `find-solution` loop (2 rounds, 9 candidates) compared DTCG with Style Dictionary or Terrazzo, CSS as the source, Panda CSS, Storybook and a design workbench app. The owner kept the TypeScript tokens.

## Decision

1. **The repository is the source of truth.** `web/packages/tokens/src/tokens.ts` holds all design values. `pnpm gen` (part of `mise run gen`) writes:
   - `css/tokens.css` and `css/tailwind.css` for the apps;
   - `design-system/project/tokens.json` and `tokens.css` for Claude Design.

   A test fails when a committed file differs.
2. **Claude Design is a view, not a source.** One "Design System" artifact shows the tokens: <https://claude.ai/artifact/X42t9JboqB9J5WzP8N9PHF>. Its index has `"editing": "code"`, so its page is read-only. An agent republishes `web/packages/tokens/design-system/project/` after each token change. Nobody edits the artifact.
3. **A palette is a closed list in `tokens.ts`.** A palette sets the brand scale and every semantic color, in light and dark. The site selects it with `<html data-palette="<id>">`. Every palette is static CSS: no inline style, so the CSP stays strict. The dashboard uses the `default` palette only.
   - **L0:** the `default` palette.
   - **L1, self-serve:** a user selects a `public` palette. Nobody deploys for that user. The storage of the choice comes with the domain work.
   - **L2, paid:** a designer adds a `private` palette through a pull request. The core records which user may use it. `tokens.ts` never holds a name or an ID of a customer.

   Font pairs and page layouts become the next closed lists, with the public page.
4. **Proofs run in `mise run check`:**
   - the contrast test checks every palette, both schemes and every text pair (`TEXT_PAIRS`, at least 4.5:1);
   - Tailwind has only the token colors (`--color-*: initial`), so a class such as `bg-red-500` produces no CSS.
5. **Prototypes stay in the repository.** For an approved prototype, the agent saves the screenshots and the exported HTML in `openspec/changes/<id>/design/`. Without Claude, a prototype is plain HTML that links `web/packages/tokens/css/tokens.css`.

## Trade-off

- **Prefer:** independence from one vendor, a stable system, fast prototypes.
- **Give up:** a hand-made sync step to Claude Design, and some freedom per page (only closed lists).

## Stress round

| Stressor | Affects | Verdict | Resolution |
| --- | --- | --- | --- |
| Claude access stops | prototypes, the view | survives | Tokens, CSS, prototypes and their HTML are in git. Codex makes HTML prototypes with `tokens.css` |
| The Claude Design token format changes | `designSystemTokens()` | accepted | One function changes. The apps do not depend on it |
| A designer adds a palette with low contrast | site | survives | The contrast test fails in `mise run check` |
| Hundreds of users want their own look | palettes | accepted | They select public palettes. A paid design adds one private palette, not a page fork |
| More than 4 palettes | the view | accepted | The view shows at most 8 themes (a light and a dark one per palette). A test guards it. Then the view shows a subset |

## Consequences

- Easier: a new look is one palette entry with a proof. The view always matches the code.
- Harder: an agent must republish the view after a token change. Free colors per user are not possible.
- Forbidden: a design value outside `tokens.ts`, an inline style on the site, a customer identifier in the tokens, an edit in the artifact.
