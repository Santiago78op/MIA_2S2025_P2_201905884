import { useEffect, useRef, useState } from 'react'
import { runCmd } from '@/lib/api'
import { useToast } from '@/lib/useToast'

export function CommandTerminal() {
  const [history, setHistory] = useState<string[]>([])
  const [line, setLine] = useState('')
  const [busy, setBusy] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const { push, View: Toasts } = useToast()

  useEffect(() => { inputRef.current?.focus() }, [])

  async function submit() {
    const trimmed = line.trim()
    if (!trimmed) return
    setBusy(true)
    try {
      const res = await runCmd(trimmed)
      if (res.ok) {
        setHistory(h => [`$ ${trimmed}`, res.output ?? '', ...h])
      } else {
        setHistory(h => [`$ ${trimmed}`, res.error ?? 'Error', ...h])
      }
    } catch (e: any) {
      push(e.message || 'Fallo de red', 'error')
    } finally {
      setBusy(false)
      setLine('')
      inputRef.current?.focus()
    }
  }

  function onKey(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') submit()
  }

  return (
    <div>
      <div className="h-60 overflow-auto scroll-slim bg-[var(--muted)] rounded-xl p-3 text-sm font-mono space-y-2">
        {history.length === 0 && (
          <div className="text-slate-500">Escribe un comando (ej.: <code>mkdisk -size=50 -unit=m -fit=ff -path="/tmp/Disco1.mia"</code>)</div>
        )}
        {history.map((ln, i) => (
          <pre key={i} className="whitespace-pre-wrap">{ln}</pre>
        ))}
      </div>
      <div className="mt-2 flex items-center gap-2">
        <input
          ref={inputRef}
          className="flex-1 px-3 py-2 rounded-lg border outline-none focus:ring-2 ring-black/10"
          placeholder="Escribe un comando…"
          value={line}
          onChange={e => setLine(e.target.value)}
          onKeyDown={onKey}
          disabled={busy}
        />
        <button onClick={submit} disabled={busy} className="px-3 py-2 rounded-lg bg-black text-white hover:opacity-90 disabled:opacity-50">Enviar</button>
      </div>
      <Toasts />
    </div>
  )
}