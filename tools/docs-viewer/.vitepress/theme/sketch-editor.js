// Full-screen Excalidraw editor for one .excalidraw.svg sketch (ADR 0017).
// "Save" exports an SVG with the scene inside and sends it to the dev server,
// which writes the file into the repository.
import { Excalidraw, convertToExcalidrawElements, exportToSvg, loadFromBlob } from '@excalidraw/excalidraw'
import '@excalidraw/excalidraw/index.css'
import { createElement } from 'react'
import { createRoot } from 'react-dom/client'

// Fonts come from the installed package, not from a CDN.
window.EXCALIDRAW_ASSET_PATH = `/@fs${__VIEWER_DIR__}/node_modules/@excalidraw/excalidraw/dist/prod/`

async function loadScene(img) {
  try {
    const res = await fetch(`${img.getAttribute('src').split('?')[0]}?t=${Date.now()}`)
    return await loadFromBlob(await res.blob(), null, null)
  } catch {
    // A new sketch or an SVG without a scene: start empty.
    return { elements: [], appState: {}, files: {} }
  }
}

async function save(api, path) {
  const svg = await exportToSvg({
    elements: api.getSceneElements(),
    appState: { ...api.getAppState(), exportEmbedScene: true, exportBackground: true, viewBackgroundColor: '#ffffff' },
    files: api.getFiles(),
  })
  const res = await fetch('/__sketch/save', {
    method: 'POST',
    headers: { 'content-type': 'application/json', 'x-sketch-token': __SKETCH_TOKEN__ },
    body: JSON.stringify({ path, svg: svg.outerHTML }),
  })
  if (!res.ok) throw new Error(await res.text())
}

export async function openEditor(img) {
  const path = img.dataset.sketch
  const scene = await loadScene(img)

  const overlay = document.createElement('div')
  overlay.className = 'sketch-overlay'
  const bar = document.createElement('div')
  bar.className = 'sketch-bar'
  const title = document.createElement('span')
  title.textContent = path
  const status = document.createElement('span')
  status.className = 'sketch-status'
  const saveButton = document.createElement('button')
  saveButton.textContent = 'Save'
  const closeButton = document.createElement('button')
  closeButton.textContent = 'Close'
  bar.append(title, status, saveButton, closeButton)
  const canvas = document.createElement('div')
  canvas.className = 'sketch-canvas'
  overlay.append(bar, canvas)
  document.body.append(overlay)
  document.documentElement.style.overflow = 'hidden'

  let api
  const root = createRoot(canvas)
  root.render(
    createElement(Excalidraw, {
      initialData: { ...scene, scrollToContent: true },
      excalidrawAPI: (a) => {
        api = a
        // For agents that draw the first version (see the viewer README).
        window.__sketch = { api, convertToExcalidrawElements }
      },
    }),
  )

  const close = () => {
    root.unmount()
    overlay.remove()
    document.documentElement.style.overflow = ''
    delete window.__sketch
  }

  saveButton.addEventListener('click', async () => {
    status.textContent = 'Saving…'
    try {
      await save(api, path)
      status.textContent = 'Saved'
      img.src = `${img.getAttribute('src').split('?')[0]}?t=${Date.now()}`
    } catch (e) {
      status.textContent = `Not saved: ${e.message}`
    }
  })
  closeButton.addEventListener('click', close)
}
