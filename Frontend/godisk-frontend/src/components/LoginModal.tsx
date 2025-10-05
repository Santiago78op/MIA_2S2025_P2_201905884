import { useState } from 'react'
import { runCmd } from '@/lib/api'

export function LoginModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const [usr, setUsr] = useState('')
  const [pwd, setPwd] = useState('')
  const [busy, setBusy] = useState(false)
  if (!open) return null

  async function doLogin() {
    setBusy(true)
    try {
      // Mientras no exista endpoint /api/login, usa el comando si tu backend lo soporta:
      // "login -usr=... -pwd=..."
      const res = await runCmd(`login -usr=${JSON.stringify(usr)} -pwd=${JSON.stringify(pwd)}`)
      if (res.ok) {
        onClose()
      } else {
        alert(res.error || 'Login falló')
      }
    } finally { setBusy(false) }
  }

  return (
    <div className="fixed inset-0 bg-black/30 grid place-items-center p-4">
      <div className="bg-white rounded-2xl w-full max-w-sm p-4 border shadow-sm">
        <h3 className="text-lg font-semibold">Iniciar sesión</h3>
        <p className="text-sm text-slate-500 mb-3">UI obligatoria en P2. Cuando implementes /api/login, cambia aquí.</p>
        <div className="space-y-2">
          <input className="w-full px-3 py-2 rounded-lg border" placeholder="Usuario" value={usr} onChange={e=>setUsr(e.target.value)} />
          <input type="password" className="w-full px-3 py-2 rounded-lg border" placeholder="Password" value={pwd} onChange={e=>setPwd(e.target.value)} />
        </div>
        <div className="mt-3 flex justify-end gap-2">
          <button className="px-3 py-2 rounded-lg border" onClick={onClose} disabled={busy}>Cancelar</button>
          <button className="px-3 py-2 rounded-lg bg-black text-white" onClick={doLogin} disabled={busy}>Entrar</button>
        </div>
      </div>
    </div>
  )
}