import { useRef, useState } from 'react'
import { runCmd } from '@/lib/api'

export function ScriptRunner() {
  const [busy, setBusy] = useState(false)
  const [log, setLog] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  async function handleFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    const text = await file.text()
    // Ejecuta cada línea no vacía como un comando
    const lines = text.split('\n').map(l => l.trim()).filter(Boolean)
    setBusy(true)
    const out: string[] = []
    for (const line of lines) {
      const res = await runCmd(line)
      out.push(`$ ${line}`)
      out.push(res.ok ? (res.output || '') : (res.error || 'error'))
    }
    setLog(out.join('\n'))
    setBusy(false)
    if (inputRef.current) inputRef.current.value = ''
  }

  return (
    <div>
      <input ref={inputRef} type="file" accept=".smia,.txt" onChange={handleFile} className="block w-full text-sm" />
      <pre className="h-56 overflow-auto scroll-slim bg-[var(--muted)] rounded-xl p-3 text-sm mt-2">{log || 'Carga un archivo .smia para ejecutar los comandos.'}</pre>
      <p className="text-xs text-slate-500 mt-1">Según el enunciado, la ejecución del script desde la app web es requerida.</p>
    </div>
  )
}