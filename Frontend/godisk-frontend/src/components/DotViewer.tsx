import { useEffect, useState } from 'react'
import { getViz } from '@/lib/viz'

export default function DotViewer({ dot }: { dot: string }) {
  const [svg, setSvg] = useState<string>('')
  const [err, setErr] = useState<string>('')

  useEffect(() => {
    let cancelled = false
    async function run() {
      setErr(''); setSvg('')
      try {
        const viz = await getViz()
        const out = await viz.renderSVGElement(dot)
        if (!cancelled) setSvg(out.outerHTML)
      } catch (e: any) {
        if (!cancelled) setErr(e?.message || 'Error renderizando DOT')
      }
    }
    if (dot?.trim()) run()
    return () => { cancelled = true }
  }, [dot])

  if (err) return <pre className="text-red-600 text-sm">{err}</pre>
  if (!dot?.trim()) return <div className="text-slate-500 text-sm">No hay DOT para mostrar.</div>
  return <div className="overflow-auto max-h-[70vh] border rounded-xl p-3" dangerouslySetInnerHTML={{ __html: svg }} />
}