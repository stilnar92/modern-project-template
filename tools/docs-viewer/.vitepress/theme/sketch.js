// Adds an "Edit sketch" button to each .excalidraw.svg image (ADR 0017).
// Only in `vitepress dev`. The editor itself loads on the first click.

import './sketch.css'

const SELECTOR = 'img[src*=".excalidraw.svg"]'

// The repository path of an image. In dev, the URL path is the path under
// content/, and content/ links the published roots of the repository.
function repoPath(img) {
  const url = new URL(img.getAttribute('src'), window.location.href)
  return decodeURIComponent(url.pathname).replace(/^\/(@fs\/[^]*?\/content\/)?/, '')
}

function decorate(img) {
  if (img.dataset.sketch) return
  img.dataset.sketch = repoPath(img)
  const wrap = document.createElement('span')
  wrap.className = 'sketch-wrap'
  img.replaceWith(wrap)
  wrap.append(img)
  const button = document.createElement('button')
  button.type = 'button'
  button.className = 'sketch-edit'
  button.textContent = 'Edit sketch'
  button.addEventListener('click', async () => {
    const { openEditor } = await import('./sketch-editor.js')
    openEditor(img)
  })
  wrap.append(button)
}

export function startSketches() {
  const scan = () => document.querySelectorAll(SELECTOR).forEach(decorate)
  scan()
  new MutationObserver(scan).observe(document.body, { childList: true, subtree: true })
}
