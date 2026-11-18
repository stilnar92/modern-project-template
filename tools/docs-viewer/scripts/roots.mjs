// The allow-list of the docs viewer (ADR 0016). Only these paths are published.
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))

export const VIEWER = path.resolve(here, '..')
export const REPO = path.resolve(VIEWER, '../..')
export const CONTENT = path.join(VIEWER, 'content')
export const DIST = path.join(VIEWER, '.vitepress/dist')

// A folder ends with a slash. A file has no slash at the end.
export const ROOTS = ['docs/', 'openspec/', 'api/', 'CONTEXT.md']

// Links to other files of the repository open on GitHub.
export const GITHUB = 'https://github.com/stilnar92/modern-project-template'
export const BLOB = `${GITHUB}/blob/main/`
export const TREE = `${GITHUB}/tree/main/`

export const isPublished = (p) => ROOTS.some((r) => (r.endsWith('/') ? p.startsWith(r) : p === r))
