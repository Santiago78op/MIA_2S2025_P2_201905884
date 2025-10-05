// Carga perezosa de Viz.js (WASM-lite build)
import type { default as VizType } from 'viz.js'

let vizPromise: Promise<any> | null = null

export async function getViz() {
  if (!vizPromise) {
    vizPromise = (async () => {
      const [{ default: Viz }, { default: render }] = await Promise.all([
        import('viz.js'),
        import('viz.js/lite.render.js'),
      ])
      return new (Viz as typeof VizType)({ render })
    })()
  }
  return vizPromise
}