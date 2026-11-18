import DefaultTheme from 'vitepress/theme'
import './custom.css'
import { defineAsyncComponent } from 'vue'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    // The sketch editor works only in dev (ADR 0017).
    if (import.meta.env.DEV && !import.meta.env.SSR) import('./sketch.js').then((m) => m.startSketches())
    // Load the API viewer only on the API page.
    app.component('OpenApi', defineAsyncComponent(() => import('./OpenApi.vue')))
  },
}
