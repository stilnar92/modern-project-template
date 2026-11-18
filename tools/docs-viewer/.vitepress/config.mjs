// Docs viewer of the repository (ADR 0016).
// The source of truth is the Markdown in docs/, openspec/, api/ and CONTEXT.md.
// scripts/link-content.mjs links these roots into content/. This file only
// changes how the pages show. It never changes a document.
import { execFileSync } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'
import { slug } from 'github-slugger'
import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'
import { BLOB, CONTENT, REPO, ROOTS, TREE, VIEWER, isPublished } from '../scripts/roots.mjs'
import { sketchSavePlugin } from '../scripts/sketch-save.mjs'

// --- Page titles and the sidebar ------------------------------------------

const firstH1 = (file) => {
  const m = fs.readFileSync(file, 'utf8').match(/^# (.+)$/m)
  return m ? m[1].replace(/[`*]/g, '').trim() : path.basename(file, '.md')
}

const route = (rel) => '/' + rel.replace(/\.md$/, '')

// Files that git ignores are not published.
const ignored = execFileSync('git', ['ls-files', '--others', '--ignored', '--exclude-standard', '--directory', ...ROOTS], { cwd: REPO })
  .toString()
  .split('\n')
  .filter(Boolean)
  .map((p) => (p.endsWith('/') ? `${p}**` : p))

const isIgnored = (rel) => ignored.some((p) => (p.endsWith('**') ? rel.startsWith(p.slice(0, -2)) : rel === p))

// Sidebar names for folders without a README.
const folderNames = { adr: 'Architecture decisions (ADR)' }

function group(dir) {
  const abs = path.join(REPO, dir)
  const entries = fs.readdirSync(abs, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))
  const items = []
  for (const e of entries) {
    const rel = path.posix.join(dir, e.name)
    if (e.name.startsWith('.') || e.name === 'node_modules' || isIgnored(rel)) continue
    if (e.isDirectory()) {
      const sub = group(rel)
      if (sub.items.length || sub.link) items.push(sub)
    } else if (e.name.endsWith('.md') && e.name !== 'README.md') {
      items.push({ text: firstH1(path.join(abs, e.name)), link: route(rel) })
    }
  }
  const readme = path.join(abs, 'README.md')
  const hasReadme = fs.existsSync(readme)
  const name = path.basename(dir)
  return {
    text: hasReadme ? firstH1(readme) : (folderNames[name] ?? name[0].toUpperCase() + name.slice(1)),
    link: hasReadme ? route(path.posix.join(dir, 'README.md')) : undefined,
    collapsed: true,
    items,
  }
}

const sidebar = [
  { text: 'Shared model', link: '/CONTEXT' },
  { ...group('docs/dev'), text: 'Development', collapsed: false },
  { ...group('docs/learning'), text: 'Learning' },
  { ...group('openspec'), text: 'OpenSpec' },
  { text: 'HTTP API', link: '/openapi' },
]

// --- Markdown rules ---------------------------------------------------------

// VitePress compiles each page as a Vue template. GitHub shows an unknown tag
// such as <placeholder> as text or drops it, but Vue fails on it. So the
// documents show unknown HTML as text. rumdl permits only these tags (MD033).
const allowedTag = /^<\/?(details|summary|br|sub|sup|kbd)\b[^>]*>$/i
const withoutComments = (c) => c.replace(/<!--[\s\S]*?-->/g, '')
const isAllowedHtml = (c) => {
  const s = withoutComments(c)
  const tags = s.match(/<[^>]*>/g) || []
  return tags.every((t) => allowedTag.test(t)) && !/<[a-z/]/i.test(s.replace(/<[^>]*>/g, ''))
}

function safeHtml(md) {
  md.core.ruler.push('safe-html', (state) => {
    // The pages of the viewer itself (home, API) may use Vue components.
    if (!isPublished(state.env.relativePath || '')) return
    const fix = (t) => {
      if ((t.type === 'html_inline' || t.type === 'html_block') && !isAllowedHtml(t.content)) {
        t.type = t.type === 'html_block' ? 'raw_text_block' : 'text'
      }
    }
    for (const t of state.tokens) {
      fix(t)
      for (const c of t.children || []) fix(c)
    }
  })
  md.renderer.rules.raw_text_block = (tokens, i) => `<p>${md.utils.escapeHtml(tokens[i].content)}</p>`
  // Vue reads {{ }} in text as an expression. Show it as text.
  const text = md.renderer.rules.text
  md.renderer.rules.text = (...a) => text(...a).replace(/\{\{/g, '&#123;&#123;').replace(/\}\}/g, '&#125;&#125;')
}

// Links to files outside the published roots, to folders and to files that
// are not Markdown open on GitHub. A link to the API contract opens the API page.
function repoLinks(md) {
  md.core.ruler.push('repo-links', (state) => {
    const rel = state.env.relativePath
    if (!rel || !isPublished(rel)) return
    const walk = (tokens) => {
      for (const t of tokens) {
        if (t.children) walk(t.children)
        if (t.type !== 'link_open' || (t.attrGet('class') || '').includes('header-anchor')) continue
        const href = t.attrGet('href')
        if (!href || /^[a-z][a-z0-9+.-]*:/i.test(href) || href.startsWith('//') || href.startsWith('#')) continue
        const [p, hash] = href.split('#')
        const target = path.posix.normalize(path.posix.join(path.posix.dirname(rel), decodeURI(p)))
        const frag = hash ? `#${hash}` : ''
        const abs = path.join(REPO, target)
        if (target === 'api/openapi.yaml') t.attrSet('href', '/openapi')
        else if (!isPublished(target)) t.attrSet('href', BLOB + target + frag)
        else if (fs.existsSync(abs) && fs.statSync(abs).isDirectory()) {
          if (fs.existsSync(path.join(abs, 'README.md'))) t.attrSet('href', `${p.replace(/\/?$/, '/README.md')}${frag}`)
          else t.attrSet('href', TREE + target)
        } else if (!target.endsWith('.md')) t.attrSet('href', BLOB + target)
      }
    }
    walk(state.tokens)
  })
}

// --- Site -------------------------------------------------------------------

export default withMermaid(
  defineConfig({
    title: 'modern-project-template docs',
    description: 'The project documents, from the repository',
    srcDir: 'content',
    srcExclude: ['**/node_modules/**', ...ignored],
    ignoreDeadLinks: false,
    lastUpdated: false,
    vite: {
      resolve: { preserveSymlinks: true },
      // The sketch editor saves .excalidraw.svg files, only in dev (ADR 0017).
      plugins: [sketchSavePlugin()],
      // mermaid imports fastdom, a CommonJS package. vitepress-plugin-mermaid
      // pre-bundles only its own list of mermaid dependencies. Without this
      // line the dev server sends raw CommonJS and each page is blank.
      optimizeDeps: { include: ['fastdom', 'fastdom/extensions/fastdom-promised.js'] },
      // The dev server reads only the viewer and the published roots, never
      // the repository root with .env.
      server: { host: 'localhost', fs: { strict: true, allow: [VIEWER, CONTENT, ...ROOTS.map((r) => path.join(REPO, r))] } },
    },
    markdown: {
      // GitHub makes heading ids with github-slugger. The same ids keep the
      // links with Cyrillic anchors working.
      anchor: { slugify: slug },
      config: (md) => {
        safeHtml(md)
        repoLinks(md)
      },
    },
    themeConfig: {
      nav: [
        { text: 'Development', link: '/docs/dev/how-it-works' },
        { text: 'Learning', link: '/docs/learning/README' },
        { text: 'API', link: '/openapi' },
      ],
      sidebar,
      search: { provider: 'local' },
      outline: [2, 3],
      socialLinks: [{ icon: 'github', link: BLOB.replace(/\/blob\/main\/$/, '') }],
    },
  }),
)
