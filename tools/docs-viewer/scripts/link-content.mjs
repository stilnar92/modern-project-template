// Make the content folder of the docs viewer (ADR 0016).
// content/ is not in git. It holds relative symlinks to the published roots
// and two pages of the viewer: the home page and the API page.
// Git does not track the symlinks, so lychee, rumdl and typos see each
// document only once, at its real path.
import fs from 'node:fs'
import path from 'node:path'
import { CONTENT, REPO, ROOTS } from './roots.mjs'

fs.rmSync(CONTENT, { recursive: true, force: true })
fs.mkdirSync(CONTENT, { recursive: true })

for (const root of ROOTS) {
  const name = root.replace(/\/$/, '')
  const target = path.relative(CONTENT, path.join(REPO, name))
  fs.symlinkSync(target, path.join(CONTENT, name))
}

fs.writeFileSync(
  path.join(CONTENT, 'index.md'),
  `---
layout: home
hero:
  name: modern-project-template
  text: Project documents
  tagline: The same files that the agents read. Edit them in the repository.
  actions:
    - theme: brand
      text: Brief
      link: /docs/dev/brief
    - theme: alt
      text: Shared model
      link: /CONTEXT
features:
  - title: Product
    details: What we build, on one page.
    link: /docs/dev/brief
  - title: Development
    details: How it works, standards and architecture decisions.
    link: /docs/dev/how-it-works
  - title: Learning
    details: Lessons from the real problems of the project.
    link: /docs/learning/README
  - title: Specs
    details: OpenSpec changes and capabilities.
    link: /openspec/README
  - title: HTTP API
    details: The contract in api/openapi.yaml.
    link: /openapi
---
`,
)

fs.writeFileSync(
  path.join(CONTENT, 'openapi.md'),
  `---
layout: page
title: HTTP API
---

<ClientOnly><OpenApi /></ClientOnly>
`,
)

console.log(`docs viewer: linked ${ROOTS.join(', ')} into ${path.relative(REPO, CONTENT)}`)
