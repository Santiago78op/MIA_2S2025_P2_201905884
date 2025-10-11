import React, { useEffect, useRef, useState } from 'react';
import { runCmd } from '@/lib/api';
import { useToast } from '@/lib/useToast';

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
    <div className="flex flex-col gap-4">
      <div className="h-64 overflow-auto scroll-slim bg-gray-100 rounded-2xl p-4 text-sm font-mono space-y-2 border shadow-inner">
        {history.length === 0 && (
          <div className="text-slate-500">Escribe un comando <span className="font-semibold">(ej: <code>mkdisk -size=50 -unit=m -fit=ff -path=\"/tmp/Disco1.mia\"</code>)</span></div>
        )}
        {history.map((ln, i) => (
          <pre key={i} className="whitespace-pre-wrap text-xs leading-tight">{ln}</pre>
        ))}
      </div>
      <form className="flex flex-col sm:flex-row items-center gap-2" onSubmit={e => {e.preventDefault(); submit();}}>
        <input
          ref={inputRef}
          className="flex-1 px-4 py-2 rounded-lg border outline-none focus:ring-2 ring-blue-700/20 bg-white text-sm shadow"
          placeholder="Escribe un comando…"
          value={line}
          onChange={e => setLine(e.target.value)}
          onKeyDown={onKey}
          disabled={busy}
        />
        <button type="submit" disabled={busy} className="px-5 py-2 rounded-lg bg-blue-700 text-white font-semibold shadow hover:bg-blue-800 transition disabled:opacity-50">Enviar</button>
      </form>
      <Toasts />
    </div>
  )
}