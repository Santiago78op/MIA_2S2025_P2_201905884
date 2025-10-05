import { useEffect, useState } from 'react'
import { runCmd } from '@/lib/api'

export function JournalPanel() {
  const [id, setId] = useState('')
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)

  async function load() {
    if (!id) return
    setBusy(true)
    try {
      const res = await runCmd(`journaling -id=${id}`)
      if (res.ok) setText(res.output || '')
      else setText(res.error || 'error')
    } finally { setBusy(false) }
  }

  async function recovery() {
    if (!id) return
    setBusy(true)
    try {
      const res = await runCmd(`recovery -id=${id}`)
      alert(res.ok ? 'Recovery OK' : (res.error || 'error'))
      await load()
    } finally { setBusy(false) }
  }

  async function loss() {
    if (!id) return
    setBusy(true)
    try {
      const res = await runCmd(`loss -id=${id}`)
      alert(res.ok ? 'Loss OK' : (res.error || 'error'))
      await load()
    } finally { setBusy(false) }
  }

  useEffect(() => { /* opcional: autoload por id */ }, [id])

  return (
    <div className="space-y-2">
      <div className="flex gap-2">
        <input className="flex-1 px-3 py-2 rounded-lg border" placeholder="ID de partición montada (p.ej. vd12)" value={id} onChange={e=>setId(e.target.value)} />
        <button className="px-3 py-2 rounded-lg border" onClick={load} disabled={busy}>Listar</button>
      </div>
      <div className="flex gap-2">
        <button className="px-3 py-2 rounded-lg bg-black text-white" onClick={recovery} disabled={busy}>Recovery</button>
        <button className="px-3 py-2 rounded-lg bg-red-600 text-white" onClick={loss} disabled={busy}>Loss</button>
      </div>
      <textarea className="w-full h-56 rounded-xl border p-2 font-mono text-sm" value={text} onChange={()=>{}} />
    </div>
  )
}