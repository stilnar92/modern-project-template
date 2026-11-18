// Check the built site of the docs viewer (ADR 0016).
// For each tracked document it checks these things in the built HTML:
// - The page exists and shows the H1 of the file once.
// - Each internal link opens a built page, and each anchor exists on it.
// - Each image exists.
// lychee checks the same links in the Markdown. This script checks that the
// viewer did not break them. It exits with 1 when a check fails.
import { execFileSync } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'
import { BLOB, DIST, REPO, ROOTS, TREE } from './roots.mjs'

const tracked = execFileSync('git', ['ls-files', ...ROOTS], { cwd: REPO })
  .toString()
  .split('\n')
  .filter((f) => f.endsWith('.md'))

const cache = new Map()
const read = (f) => {
  if (!cache.has(f)) cache.set(f, fs.existsSync(f) && fs.statSync(f).isFile() ? fs.readFileSync(f, 'utf8') : null)
  return cache.get(f)
}
const unescape = (s) => s.replace(/&#39;/g, "'").replace(/&quot;/g, '"').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&amp;/g, '&')

// A route such as /docs/dev/core or core.html points to a built .html file.
const htmlFile = (from, p) => {
  const base = p.startsWith('/') ? path.join(DIST, decodeURIComponent(p)) : path.join(path.dirname(from), decodeURIComponent(p))
  if (base.endsWith('.html')) return base
  if (base.endsWith('/')) return path.join(base, 'index.html')
  return `${base}.html`
}

const s = { pages: 0, links: 0, anchors: 0, cyrillicAnchors: 0, github: 0, images: 0, titles: 0 }
const errors = []

for (const src of tracked) {
  const page = path.join(DIST, src.replace(/\.md$/, '.html'))
  const html = read(page)
  if (!html) {
    errors.push(`${src}: no built page`)
    continue
  }
  s.pages++
  // Only the document body, not the navigation and the sidebar.
  const main = (html.split('<main')[1] || '').split('</main>')[0]
  const doc = main.split('class="vp-doc')[1] || main
  const h1 = fs.readFileSync(path.join(REPO, src), 'utf8').match(/^# (.+)$/m)?.[1].replace(/[`*]/g, '').trim()
  // A template can have no H1 (rumdl MD041 exception). No page may get a second H1.
  if ((doc.match(/<h1[\s>]/g) || []).length !== (h1 ? 1 : 0)) errors.push(`${src}: the page must show the H1 of the file once`)

  const title = unescape(html.match(/<title>([^<|]*)/)?.[1].trim() || '')
  if (h1 && title === h1) s.titles++

  for (const m of doc.matchAll(/<a [^>]*href="([^"]*)"[^>]*>/g)) {
    if (/header-anchor/.test(m[0])) continue
    const href = unescape(m[1])
    if (href.startsWith(BLOB) || href.startsWith(TREE)) {
      s.github++
      continue
    }
    if (/^[a-z][a-z0-9+.-]*:/i.test(href) || href.startsWith('//')) continue
    s.links++
    const [p, hash] = href.split('#')
    const target = p ? htmlFile(page, p) : page
    const targetHtml = read(target)
    if (!targetHtml) {
      errors.push(`${src}: ${href} opens no page`)
      continue
    }
    if (hash) {
      const id = decodeURIComponent(hash)
      s.anchors++
      if (/[а-яё]/i.test(id)) s.cyrillicAnchors++
      if (!targetHtml.includes(`id="${id}"`)) errors.push(`${src}: ${href} has no heading #${id}`)
    }
  }

  for (const m of doc.matchAll(/<img [^>]*src="([^"]*)"/g)) {
    s.images++
    const src2 = m[1].split('?')[0]
    if (/^[a-z]+:/i.test(src2)) continue
    if (!fs.existsSync(path.join(DIST, src2))) errors.push(`${src}: image ${m[1]} is missing`)
  }
}

console.log(
  `docs viewer: ${s.pages}/${tracked.length} pages, ${s.links} internal links, ${s.anchors} anchors ` +
    `(${s.cyrillicAnchors} Cyrillic), ${s.github} links to GitHub, ${s.images} images, ` +
    `${s.titles} titles equal to the H1`,
)
if (errors.length) {
  console.error(`${errors.length} problems:\n${errors.join('\n')}`)
  process.exit(1)
}
