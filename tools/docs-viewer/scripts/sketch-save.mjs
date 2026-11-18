// Dev-server endpoint that saves a sketch from the Excalidraw editor (ADR 0017).
// It exists only in `vitepress dev`. The static build has no endpoint.
//
// A sketch is one file "<name>.excalidraw.svg" in a published root. The SVG
// shows on GitHub and in the built site. The Excalidraw scene is embedded in
// it, so the editor can open it again.
//
// Guards: a random token for each server start, the same origin, the
// ".excalidraw.svg" suffix, a path inside a published root, and an SVG body.
import crypto from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'
import { REPO, ROOTS, VIEWER } from './roots.mjs'

const MAX_BYTES = 10 * 1024 * 1024
const SUFFIX = '.excalidraw.svg'

export const SKETCH_TOKEN = crypto.randomBytes(24).toString('hex')

const folders = ROOTS.filter((r) => r.endsWith('/'))

// Returns the absolute path of the sketch, or null when the path is not allowed.
export function sketchPath(rel) {
  if (typeof rel !== 'string' || !rel.endsWith(SUFFIX) || rel.includes('\0')) return null
  const norm = path.posix.normalize(rel.replace(/^\/+/, ''))
  if (norm.startsWith('..') || !folders.some((f) => norm.startsWith(f))) return null
  const abs = path.resolve(REPO, norm)
  if (!abs.startsWith(REPO + path.sep)) return null
  const dir = path.dirname(abs)
  if (!fs.existsSync(dir) || !fs.realpathSync(dir).startsWith(REPO + path.sep)) return null
  return abs
}

function send(res, code, text) {
  res.statusCode = code
  res.setHeader('content-type', 'text/plain; charset=utf-8')
  res.end(text)
}

export function sketchSavePlugin() {
  return {
    name: 'docs-viewer:sketch-save',
    apply: 'serve',
    config: () => ({
      define: { __SKETCH_TOKEN__: JSON.stringify(SKETCH_TOKEN), __VIEWER_DIR__: JSON.stringify(VIEWER) },
      // Pre-bundle the editor at start, so the first click does not reload the page.
      optimizeDeps: { include: ['@excalidraw/excalidraw', 'react', 'react-dom/client'] },
    }),
    configureServer(server) {
      server.middlewares.use('/__sketch/save', (req, res) => {
        if (req.method !== 'POST') return send(res, 405, 'POST only')
        const origin = req.headers.origin
        if (!origin || new URL(origin).host !== req.headers.host) return send(res, 403, 'wrong origin')
        if (req.headers['x-sketch-token'] !== SKETCH_TOKEN) return send(res, 403, 'wrong token')

        const chunks = []
        let size = 0
        req.on('data', (c) => {
          size += c.length
          if (size > MAX_BYTES) req.destroy()
          else chunks.push(c)
        })
        req.on('end', () => {
          let body
          try {
            body = JSON.parse(Buffer.concat(chunks).toString('utf8'))
          } catch {
            return send(res, 400, 'bad JSON')
          }
          const abs = sketchPath(body.path)
          if (!abs) return send(res, 400, `not an allowed sketch path: ${body.path}`)
          if (typeof body.svg !== 'string' || !body.svg.trimStart().startsWith('<svg')) return send(res, 400, 'not an SVG')
          fs.writeFileSync(abs, body.svg.endsWith('\n') ? body.svg : `${body.svg}\n`)
          server.config.logger.info(`sketch saved: ${path.relative(REPO, abs)}`)
          send(res, 200, 'saved')
        })
      })
    },
  }
}
